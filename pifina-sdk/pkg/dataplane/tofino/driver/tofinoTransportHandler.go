package driver

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"google.golang.org/grpc"
)

// Connects to a tofino switch.
func (driver *TofinoDriver) Connect(ctx context.Context, endpoint string, p4name string, connectTimeout int) error {
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

	// Open stream channel
	driver.streamChannel, err = driver.client.StreamChannel(driver.ctx)

	reqSub := bfruntime.StreamMessageRequest_Subscribe{
		Subscribe: &bfruntime.Subscribe{
			DeviceId: 0,
			Notifications: &bfruntime.Subscribe_Notifications{
				EnablePortStatusChangeNotifications: false,
				EnableIdletimeoutNotifications:      true,
				EnableLearnNotifications:            true,
			},
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

	// Bind client
	reqFPCfg := bfruntime.SetForwardingPipelineConfigRequest{
		ClientId: driver.clientId,
		DeviceId: 0,
		Action:   bfruntime.SetForwardingPipelineConfigRequest_BIND,
	}
	reqFPCfg.Config = append(reqFPCfg.Config, &bfruntime.ForwardingPipelineConfig{P4Name: p4name})

	var setForwardPipelineConfigResponse *bfruntime.SetForwardingPipelineConfigResponse
	setForwardPipelineConfigResponse, err = driver.client.SetForwardingPipelineConfig(driver.ctx, &reqFPCfg)

	if setForwardPipelineConfigResponse == nil || setForwardPipelineConfigResponse.GetSetForwardingPipelineConfigResponseType() != bfruntime.SetForwardingPipelineConfigResponseType_WARM_INIT_STARTED {
		driver.Disconnect()
		return errors.New(fmt.Sprintf("tofino ASIC driver: Warm Init Failed : %s", err))
	}

	driver.logger.Info("Warm INIT Started")

	// Request Runtome CFG
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

func (driver *TofinoDriver) getEntityBySingleKey(tblName, keyName string, keyValue []byte) ([]*bfruntime.Entity, error) {
	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, errors.New(fmt.Sprintf("Cannot find table id of %s", tblName))
	}

	keyId := driver.GetKeyIdByName(tblName, keyName)
	if keyId == 0 {
		return nil, errors.New(fmt.Sprintf("Cannot find key id of %s", keyName))
	}

	keyFields := []*bfruntime.KeyField{
		{
			FieldId: keyId,
			MatchType: &bfruntime.KeyField_Exact_{
				Exact: &bfruntime.KeyField_Exact{
					Value: keyValue,
				},
			},
		},
	}
	tblEntries := []*bfruntime.Entity{}

	tblEntries = append(tblEntries,
		&bfruntime.Entity{
			Entity: &bfruntime.Entity_TableEntry{
				TableEntry: &bfruntime.TableEntry{
					TableId: tblId,
					Value: &bfruntime.TableEntry_Key{
						Key: &bfruntime.TableKey{
							Fields: keyFields,
						},
					},
				},
			},
		},
	)

	readReq := &bfruntime.ReadRequest{
		ClientId: driver.clientId,
		Entities: tblEntries,
		Target: &bfruntime.TargetDevice{
			DeviceId:  0,
			PipeId:    0xffff,
			PrsrId:    255,
			Direction: 255,
		},
	}
	// Send read request
	readClient, err := driver.client.Read(driver.ctx, readReq)
	if err != nil {
		return nil, err
	}
	// Read response
	resp, err := readClient.Recv()
	if err != nil {
		return nil, err
	}

	return resp.Entities, nil
}

// Low Level read request handler
func (driver *TofinoDriver) SendReadRequest(tblEntries []*bfruntime.Entity) ([]*bfruntime.Entity, error) {

	readReq := &bfruntime.ReadRequest{
		ClientId: driver.clientId,
		Entities: tblEntries,
		Target: &bfruntime.TargetDevice{
			DeviceId:  0,
			PipeId:    0xffff,
			PrsrId:    255,
			Direction: 255,
		},
	}

	// Send read request
	readClient, err := driver.client.Read(driver.ctx, readReq)
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

// Enable Sync operation on register
func (driver *TofinoDriver) EnableSyncOperationOnRegister(tblName string) error {
	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return errors.New(fmt.Sprintf("Cannot find table id of %s", tblName))
	}

	tblEntry := &bfruntime.TableOperation{
		TableId:             tblId,
		TableOperationsType: "Sync",
	}

	updateItems := []*bfruntime.Update{
		{
			Type: bfruntime.Update_INSERT,
			Entity: &bfruntime.Entity{
				Entity: &bfruntime.Entity_TableOperation{
					TableOperation: tblEntry,
				},
			},
		},
	}

	writeReq := bfruntime.WriteRequest{
		ClientId:  driver.clientId,
		Atomicity: bfruntime.WriteRequest_CONTINUE_ON_ERROR,
		Target: &bfruntime.TargetDevice{
			DeviceId:  0,
			PipeId:    0xffff,
			PrsrId:    255,
			Direction: 255,
		},
		Updates: updateItems,
	}

	_, err := driver.client.Write(driver.ctx, &writeReq)
	if err != nil {
		driver.logger.Error("Enable sync operation on register failed.", "register", tblName, "err", err)
		return err
	}
	return nil

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
