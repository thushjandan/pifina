package driver

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
	"google.golang.org/grpc"
)

// Connects to a tofino switch.
func (driver *TofinoDriver) Connect(ctx context.Context, endpoint string, connectTimeout int) error {
	// If a connection already exists, then return
	if driver.isConnected {
		return nil
	}
	driver.logger.Info("Connect to Tofino", "endpoint", endpoint)

	var err error

	maxSizeOpt := grpc.MaxCallRecvMsgSize(16 * 10e6) // increase incoming grpc message size to 16MB
	driver.conn, err = grpc.Dial(
		endpoint,
		grpc.WithTimeout(time.Duration(connectTimeout)*time.Second), // Set connect timeout
		grpc.WithDefaultCallOptions(maxSizeOpt),                     // Set incoming grpc message size
		grpc.WithInsecure(),                                         // Without SSL/TLS
		grpc.WithBlock(),
	)

	if err != nil {
		return errors.New(fmt.Sprintf("Could not connect to Tofino %v\n", err))
	}

	driver.logger.Info("Gen new Client", "clientId", strconv.FormatUint(uint64(driver.clientId), 10))
	driver.client = bfruntime.NewBfRuntimeClient(driver.conn)

	driver.ctx, driver.cancel = context.WithCancel(ctx)

	// Open stream channel to associate my client id with device id
	driver.streamChannel, err = driver.client.StreamChannel(driver.ctx)

	reqSub := bfruntime.StreamMessageRequest_Subscribe{
		Subscribe: &bfruntime.Subscribe{
			DeviceId: 0,
		},
	}

	err = driver.streamChannel.Send(&bfruntime.StreamMessageRequest{ClientId: driver.clientId, Update: &reqSub})

	counter := 0
	for err != nil && counter < 3 {
		driver.logger.Error("Subscribe failed: trying new id", "err", err, "clientId", fmt.Sprint(driver.clientId+1))
		counter += 1
		driver.clientId += 1
		err = driver.streamChannel.Send(&bfruntime.StreamMessageRequest{ClientId: driver.clientId, Update: &reqSub})
	}

	driver.isConnected = true

	// Request runtime configuration
	reqGFPCfg := bfruntime.GetForwardingPipelineConfigRequest{
		ClientId: driver.clientId,
		DeviceId: 0,
	}
	var getForwardPipelineConfigResponse *bfruntime.GetForwardingPipelineConfigResponse
	getForwardPipelineConfigResponse, err = driver.client.GetForwardingPipelineConfig(driver.ctx, &reqGFPCfg)

	if getForwardPipelineConfigResponse == nil {
		driver.Disconnect()
		return errors.New(fmt.Sprintf("Could not get ForwardingPipelineConfig : %s", err))
	}

	driver.logger.Info("Connection is ready to use")
	// Parse BfrtInfo
	driver.P4Tables, err = UnmarshalBfruntimeInfoJson(getForwardPipelineConfigResponse.Config[0].BfruntimeInfo)
	if err != nil {
		driver.Disconnect()
		return errors.New(fmt.Sprintf("Could not parse P4Table BfrtInfo payload. Error: %v", err))
	}
	// Create Hash table for faster retrieval of tables
	driver.createP4TableIndex()
	// Parse NonP4Tables BfrtInfo
	driver.NonP4Tables, err = UnmarshalBfruntimeInfoJson(getForwardPipelineConfigResponse.NonP4Config.BfruntimeInfo)
	if err != nil {
		driver.Disconnect()
		return errors.New(fmt.Sprintf("Could not parse NonP4Table BfrtInfo payload. Error: %v", err))
	}
	// Create Hash table for faster retrieval of tables
	driver.createNonP4TableIndex()

	// Create Hash map for port cache
	driver.portCache = make(map[string][]byte)

	return nil
}

func (driver *TofinoDriver) getIndirectCounterResetRequest(shortTblName string, keyName string, keyValue uint32, dataNames []string, dataSize int) (*bfruntime.Entity, error) {
	tblName := driver.FindTableNameByShortName(shortTblName)

	if tblName == "" {
		driver.logger.Error("cannot find table for the probe", "tblName", tblName)
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name", Entity: tblName}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name", Entity: tblName}
	}

	keyId := driver.GetKeyIdByName(tblName, keyName)
	if keyId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find key id for table name", Entity: tblName}
	}
	// Convert to byte slice
	byteEntryId := make([]byte, 4)
	binary.BigEndian.PutUint32(byteEntryId, keyValue)

	keyFields := []*bfruntime.KeyField{
		{
			FieldId: keyId,
			MatchType: &bfruntime.KeyField_Exact_{
				Exact: &bfruntime.KeyField_Exact{
					Value: byteEntryId,
				},
			},
		},
	}

	var dataFields []*bfruntime.DataField
	for _, dataName := range dataNames {
		dataId, _ := driver.GetSingletonDataIdLikeName(tblName, dataName)
		if dataId == 0 {
			return nil, &model.ErrNameNotFound{Msg: "Cannot data name to reset the counter", Entity: dataName}
		}
		// Set the counter value to 0. The byte array needs to have the same length as on the dataplane
		dataField := &bfruntime.DataField{
			FieldId: dataId,
			Value: &bfruntime.DataField_Stream{
				Stream: make([]byte, dataSize),
			},
		}
		dataFields = append(dataFields, dataField)
	}

	tblEntry := &bfruntime.Entity{
		Entity: &bfruntime.Entity_TableEntry{
			TableEntry: &bfruntime.TableEntry{
				TableId: tblId,
				Value: &bfruntime.TableEntry_Key{
					Key: &bfruntime.TableKey{
						Fields: keyFields,
					},
				},
				Data: &bfruntime.TableData{
					Fields: dataFields,
				},
			},
		},
	}

	return tblEntry, nil
}

// Low Level read request handler
func (driver *TofinoDriver) SendReadRequest(tblEntries []*bfruntime.Entity) ([]*bfruntime.Entity, error) {
	return driver.SendReadRequestByPipeId(tblEntries, TOFINO_PIPE_ID)
}

func (driver *TofinoDriver) SendReadRequestByPipeId(tblEntries []*bfruntime.Entity, pipeId int) ([]*bfruntime.Entity, error) {

	readReq := &bfruntime.ReadRequest{
		ClientId: driver.clientId,
		P4Name:   driver.p4Name,
		Entities: tblEntries,
		Target: &bfruntime.TargetDevice{
			DeviceId:  0,
			PipeId:    uint32(pipeId),
			PrsrId:    255,
			Direction: 255,
		},
	}

	if !driver.isConnected {
		return nil, &model.ErrNotReady{Msg: "Not connected to Tofino"}
	}
	// Only single access to device allowed.
	// Otherwise undefined behaviour of Tofino device could occur
	// pipe_mgr could complains that a batch is already in progress
	driver.lock.Lock()
	defer driver.lock.Unlock()

	ctx, cancel := context.WithTimeout(driver.ctx, 5*time.Second)
	defer cancel()
	// Send read request
	readClient, err := driver.client.Read(ctx, readReq)
	if err != nil {
		return nil, err
	}

	// Read response
	resp, err := readClient.Recv()
	if err != nil {
		return nil, err
	}

	return resp.GetEntities(), nil
}

func (driver *TofinoDriver) SendWriteRequest(updateItems []*bfruntime.Update) error {
	// Ignore empty write requests
	if updateItems == nil {
		return nil
	}

	writeReq := bfruntime.WriteRequest{
		ClientId:  driver.clientId,
		P4Name:    driver.p4Name,
		Atomicity: bfruntime.WriteRequest_CONTINUE_ON_ERROR,
		Target: &bfruntime.TargetDevice{
			DeviceId:  0,
			PipeId:    TOFINO_PIPE_ID,
			PrsrId:    255,
			Direction: 255,
		},
		Updates: updateItems,
	}

	if driver.isConnected {
		// Only single access to device allowed
		driver.lock.Lock()
		defer driver.lock.Unlock()
		ctx, cancel := context.WithTimeout(driver.ctx, 5*time.Second)
		defer cancel()
		_, err := driver.client.Write(ctx, &writeReq)
		if err != nil {
			return err
		}
	}

	return nil
}

// Process metric responses and transform to metric item objects
func (driver *TofinoDriver) ProcessMetricResponse(entities []*bfruntime.Entity) ([]*model.MetricItem, error) {
	// Transform response
	transformedMetrics := make([]*model.MetricItem, 0, len(entities))
	timeNow := time.Now()
	for i := range entities {
		tblId := entities[i].GetTableEntry().GetTableId()
		tableType := driver.GetTableTypeById(tblId)
		// Process match action metrics
		if tableType == TABLE_TYPE_MATCHACTION {
			metric, err := driver.ProcessMatchActionResponse(entities[i])
			if err != nil || len(metric) == 0 {
				continue
			}
			for metric_i := range metric {
				metric[metric_i].LastUpdated = timeNow
			}
			transformedMetrics = append(transformedMetrics, metric...)
		}

		// Process register metrics
		if tableType == TABLE_TYPE_REGISTER {
			metric, err := driver.ProcessRegisterResponse(entities[i])
			if err != nil || metric == nil {
				continue
			}
			metric.LastUpdated = timeNow
			transformedMetrics = append(transformedMetrics, metric)
		}

		// Process Counter metrics
		if tableType == TABLE_TYPE_COUNTER {
			metric, err := driver.ProcessCounterResponse(entities[i])
			if err != nil || len(metric) == 0 {
				continue
			}
			for metric_i := range metric {
				metric[metric_i].LastUpdated = timeNow
			}
			transformedMetrics = append(transformedMetrics, metric...)
		}
		// Process TM counters
		if tableType == TABLE_TYPE_TM_CNT_IG || tableType == TABLE_TYPE_TM_CNT_EG {
			metric, err := driver.ProcessTMCounters(entities[i])
			if err != nil || len(metric) == 0 {
				continue
			}
			for metric_i := range metric {
				metric[metric_i].LastUpdated = timeNow
			}
			transformedMetrics = append(transformedMetrics, metric...)
		}
	}
	return transformedMetrics, nil
}

// Disconnects from Tofino switch
func (driver *TofinoDriver) Disconnect() {
	if driver.isConnected {
		driver.logger.Info("Disconnecting from Tofino.", "endpoint", driver.conn.Target())
		driver.client = nil
		driver.conn.Close()
		driver.cancel()
		driver.ctx.Done()
		driver.isConnected = false
	}
}
