package driver

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

func (driver *TofinoDriver) GetPortIdByName(portName string) ([]byte, error) {
	// Check the cache first
	if portId, ok := driver.portCache[portName]; ok {
		return portId, nil
	}

	return nil, &model.ErrNameNotFound{Msg: "Port not found in port cache", Entity: portName}
}

// Get all port names from the port cache
func (driver *TofinoDriver) GetAvailablePortNames() []*model.DevPort {
	portNames := make([]*model.DevPort, 0)
	for key := range driver.portCache {
		portNames = append(portNames, &model.DevPort{Name: key, PortId: binary.BigEndian.Uint32(driver.portCache[key])})
	}
	return portNames
}

// Load port cache. Creates a mapping between port name and dev port (port id)
func (driver *TofinoDriver) LoadPortNameCache() error {
	// get table id from bfrtinfo
	sliceIdx, ok := driver.indexNonP4Tables[TABLE_NAME_PORT_INFO]
	// Find table name in index
	if !ok {
		return &model.ErrNameNotFound{Msg: "Table Id not found in non index table cache", Entity: TABLE_NAME_PORT_INFO}
	}

	tblEntries := []*bfruntime.Entity{
		{
			Entity: &bfruntime.Entity_TableEntry{
				TableEntry: &bfruntime.TableEntry{
					TableId: driver.NonP4Tables[sliceIdx].Id,
				},
			},
		},
	}

	// Read response
	entities, err := driver.SendReadRequest(tblEntries)
	if err != nil {
		return err
	}

	// Check if response is empty in case the item has not found
	if len(entities) == 0 {
		return &model.ErrNameNotFound{Msg: "No port information have been returned by device", Entity: TABLE_NAME_PORT_INFO}
	}

	for i := range entities {
		portId := entities[i].GetTableEntry().GetData().GetFields()[0].GetStream()
		portName := string(entities[i].GetTableEntry().GetKey().GetFields()[0].GetExact().GetValue())
		// We need only 2 bytes
		driver.portCache[portName] = portId
	}
	driver.logger.Info("Port cache have been loaded", "portCount", len(entities))

	return nil
}

// Retrieves ingress and egress port counters from the perspective of TM.
func (driver *TofinoDriver) GetTMCountersByPort(ports []string) ([]*model.MetricItem, error) {
	tblEntries := []*bfruntime.Entity{}
	tblId_ig := driver.GetTableIdByName(TABLE_NAME_TM_CNT_IG)
	tblId_eg := driver.GetTableIdByName(TABLE_NAME_TM_CNT_EG)
	keyId_ig := driver.GetKeyIdByName(TABLE_NAME_TM_CNT_IG, DEV_PORT_KEY_NAME)
	keyId_eg := driver.GetKeyIdByName(TABLE_NAME_TM_CNT_EG, DEV_PORT_KEY_NAME)

	for i := range ports {
		portId, err := driver.GetPortIdByName(ports[i])
		if err != nil {
			continue
		}
		// INGRESS
		tblEntries = append(tblEntries,
			&bfruntime.Entity{
				Entity: &bfruntime.Entity_TableEntry{
					TableEntry: &bfruntime.TableEntry{
						TableId: tblId_ig,
						TableFlags: &bfruntime.TableFlags{
							FromHw: true,
						},
						Value: &bfruntime.TableEntry_Key{
							Key: &bfruntime.TableKey{
								Fields: []*bfruntime.KeyField{
									{
										FieldId: keyId_ig,
										MatchType: &bfruntime.KeyField_Exact_{
											Exact: &bfruntime.KeyField_Exact{
												Value: portId,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		)
		// EGRESS
		tblEntries = append(tblEntries,
			&bfruntime.Entity{
				Entity: &bfruntime.Entity_TableEntry{
					TableEntry: &bfruntime.TableEntry{
						TableId: tblId_eg,
						TableFlags: &bfruntime.TableFlags{
							FromHw: true,
						},
						Value: &bfruntime.TableEntry_Key{
							Key: &bfruntime.TableKey{
								Fields: []*bfruntime.KeyField{
									{
										FieldId: keyId_eg,
										MatchType: &bfruntime.KeyField_Exact_{
											Exact: &bfruntime.KeyField_Exact{
												Value: portId,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		)
	}

	// Send read request to switch.
	entities, err := driver.SendReadRequest(tblEntries)
	if err != nil {
		return nil, err
	}

	// Transform response
	transformedMetrics := make([]*model.MetricItem, 0, len(entities))
	timeNow := time.Now()
	for i := range entities {
		tblEntry := entities[i].GetTableEntry()
		dataEntries := tblEntry.GetData().GetFields()
		for data_i := range dataEntries {
			// Dataplane could return just a single byte instead of 4 bytes.
			// So we copy the response in a 4 byte slice.
			rawValue := dataEntries[data_i].GetStream()

			decodedValue := binary.BigEndian.Uint64(rawValue)
			tblName := driver.GetTableNameById(tblEntry.GetTableId())
			decodedPortId := binary.BigEndian.Uint32(tblEntry.GetKey().GetFields()[0].GetExact().GetValue())
			dataFieldName := driver.GetSingletonDataNameById(tblName, dataEntries[data_i].FieldId)
			tblNameSplit := strings.Split(tblName, ".")
			shortTblName := tblNameSplit[len(tblNameSplit)-1]
			newMetric := &model.MetricItem{
				SessionId:   decodedPortId,
				Value:       uint64(decodedValue),
				Type:        model.METRIC_EXT_VALUE,
				MetricName:  fmt.Sprintf("PF_TM_%s_%s", shortTblName, dataFieldName),
				LastUpdated: timeNow,
			}
			transformedMetrics = append(transformedMetrics, newMetric)
		}
	}

	return transformedMetrics, nil
}

// Retrieves register values by a list of appRegister structs, which are used as index.
func (driver *TofinoDriver) GetTMPipelineCounter() ([]*model.MetricItem, error) {
	tblEntries := []*bfruntime.Entity{}
	tblId_pipe := driver.GetTableIdByName(TABLE_NAME_TM_CNT_PIPE)
	tblEntries = append(tblEntries,
		&bfruntime.Entity{
			Entity: &bfruntime.Entity_TableEntry{
				TableEntry: &bfruntime.TableEntry{
					TableId: tblId_pipe,
				},
			},
		},
	)

	// Transform response
	transformedMetrics := make([]*model.MetricItem, 0)
	timeNow := time.Now()
	for pipe_id := 0; pipe_id < 1; pipe_id++ {
		// Send read request to switch.
		entities, err := driver.SendReadRequestByPipeId(tblEntries, pipe_id)
		if err != nil {
			return nil, err
		}
		for i := range entities {
			tblEntry := entities[i].GetTableEntry()
			dataEntries := tblEntry.GetData().GetFields()
			for data_i := range dataEntries {
				// Dataplane could return just a single byte instead of 4 bytes.
				// So we copy the response in a 4 byte slice.
				rawValue := dataEntries[data_i].GetStream()

				decodedValue := binary.BigEndian.Uint64(rawValue)
				tblName := driver.GetTableNameById(tblEntry.GetTableId())
				dataFieldName := driver.GetSingletonDataNameById(tblName, dataEntries[data_i].FieldId)
				tblNameSplit := strings.Split(tblName, ".")
				shortTblName := tblNameSplit[len(tblNameSplit)-1]
				newMetric := &model.MetricItem{
					SessionId:   uint32(pipe_id),
					Value:       uint64(decodedValue),
					Type:        model.METRIC_EXT_VALUE,
					MetricName:  fmt.Sprintf("PF_TM_%s_%s", shortTblName, dataFieldName),
					LastUpdated: timeNow,
				}
				transformedMetrics = append(transformedMetrics, newMetric)
			}
		}
	}

	return transformedMetrics, nil
}
