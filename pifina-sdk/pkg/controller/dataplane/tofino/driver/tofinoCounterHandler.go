package driver

import (
	"encoding/binary"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

// Retrieve egress start packet counter by a list of sessionIds, which are used as index
func (driver *TofinoDriver) GetEgressStartCounter(sessionIds []uint32) ([]*bfruntime.Entity, error) {
	driver.logger.Trace("Requesting Egress start byte counter", "sessionIds", sessionIds)
	metrics, err := driver.GetMetricFromCounterRequest(sessionIds, PROBE_EGRESS_START_CNT)
	/*if err == nil {
		driver.ResetCounter(sessionIds, PROBE_EGRESS_START_CNT)
	}*/

	return metrics, err
}

func (driver *TofinoDriver) GetMetricFromCounter(sessionIds []uint32, shortTblName string) ([]*model.MetricItem, error) {
	if len(sessionIds) == 0 {
		driver.logger.Debug("Given list of session ids is empty. Skipping collecting egress start counter.")
		return nil, nil
	}

	tblName, ok := driver.probeTableMap[shortTblName]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: shortTblName}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	keyId := driver.GetKeyIdByName(tblName, COUNTER_INDEX_KEY_NAME)
	if keyId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find key id for table name", Entity: tblName}
	}

	tblEntries := []*bfruntime.Entity{}

	for _, sessionId := range sessionIds {
		// Convert to byte slice
		byteEntryId := make([]byte, 4)
		binary.BigEndian.PutUint32(byteEntryId, sessionId)

		tblEntries = append(tblEntries,
			&bfruntime.Entity{
				Entity: &bfruntime.Entity_TableEntry{
					TableEntry: &bfruntime.TableEntry{
						TableId:        tblId,
						IsDefaultEntry: false,
						TableFlags: &bfruntime.TableFlags{
							FromHw: true,
						},
						Value: &bfruntime.TableEntry_Key{
							Key: &bfruntime.TableKey{
								Fields: []*bfruntime.KeyField{
									{
										FieldId: keyId,
										MatchType: &bfruntime.KeyField_Exact_{
											Exact: &bfruntime.KeyField_Exact{
												Value: byteEntryId,
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
	driver.logger.Trace("Requesting Egress start byte counter", "sessionIds", sessionIds)

	// Send read request to switch.
	entities, err := driver.SendReadRequest(tblEntries)
	if err != nil {
		return nil, err
	}
	// Get key Ids
	counterBytesKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_BYTES)
	counterPktsKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_PKTS)

	// Transform response
	transformedMetrics := make([]*model.MetricItem, 0, len(entities))
	timeNow := time.Now()
	for i := range entities {
		sessionId := binary.BigEndian.Uint32(entities[i].GetTableEntry().GetKey().GetFields()[0].GetExact().GetValue())
		dataEntries := entities[i].GetTableEntry().GetData().GetFields()
		for data_i := range dataEntries {
			// If the key indicates a byte counter
			if dataEntries[data_i].FieldId == counterBytesKeyId {
				transformedMetrics = append(transformedMetrics, &model.MetricItem{
					SessionId:   sessionId,
					Value:       binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
					Type:        model.METRIC_BYTES,
					MetricName:  PROBE_EGRESS_START_CNT,
					LastUpdated: timeNow,
				})
			}
			// If the key indicates a packet counter
			if dataEntries[data_i].FieldId == counterPktsKeyId {
				transformedMetrics = append(transformedMetrics, &model.MetricItem{
					SessionId:   sessionId,
					Value:       binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
					Type:        model.METRIC_PKTS,
					MetricName:  PROBE_EGRESS_START_CNT,
					LastUpdated: timeNow,
				})
			}
		}
	}

	return transformedMetrics, nil
}

func (driver *TofinoDriver) GetMetricFromCounterRequest(sessionIds []uint32, shortTblName string) ([]*bfruntime.Entity, error) {
	if len(sessionIds) == 0 {
		driver.logger.Debug("Given list of session ids is empty. Skipping collecting egress start counter.")
		return nil, nil
	}

	tblName, ok := driver.probeTableMap[shortTblName]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: shortTblName}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	keyId := driver.GetKeyIdByName(tblName, COUNTER_INDEX_KEY_NAME)
	if keyId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find key id for table name", Entity: tblName}
	}

	tblEntries := []*bfruntime.Entity{}

	for _, sessionId := range sessionIds {
		// Convert to byte slice
		byteEntryId := make([]byte, 4)
		binary.BigEndian.PutUint32(byteEntryId, sessionId)

		tblEntries = append(tblEntries,
			&bfruntime.Entity{
				Entity: &bfruntime.Entity_TableEntry{
					TableEntry: &bfruntime.TableEntry{
						TableId:        tblId,
						IsDefaultEntry: false,
						TableFlags: &bfruntime.TableFlags{
							FromHw: true,
						},
						Value: &bfruntime.TableEntry_Key{
							Key: &bfruntime.TableKey{
								Fields: []*bfruntime.KeyField{
									{
										FieldId: keyId,
										MatchType: &bfruntime.KeyField_Exact_{
											Exact: &bfruntime.KeyField_Exact{
												Value: byteEntryId,
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

	return tblEntries, nil
}

func (driver *TofinoDriver) ProcessCounterResponse(entity *bfruntime.Entity) ([]*model.MetricItem, error) {
	tblName := driver.GetTableNameById(entity.GetTableEntry().GetTableId())
	// Get key Ids
	counterBytesKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_BYTES)
	counterPktsKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_PKTS)

	// Transform response
	transformedMetrics := make([]*model.MetricItem, 0, 2)
	sessionId := binary.BigEndian.Uint32(entity.GetTableEntry().GetKey().GetFields()[0].GetExact().GetValue())
	dataEntries := entity.GetTableEntry().GetData().GetFields()
	// Replace full tblname with short name
	shortTblName := driver.FindShortTableNameByName(tblName)
	if shortTblName != "" {
		tblName = shortTblName
	}

	for data_i := range dataEntries {
		// If the key indicates a byte counter
		if dataEntries[data_i].FieldId == counterBytesKeyId {
			transformedMetrics = append(transformedMetrics, &model.MetricItem{
				SessionId:  sessionId,
				Value:      binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
				Type:       model.METRIC_BYTES,
				MetricName: tblName,
			})
		}
		// If the key indicates a packet counter
		if dataEntries[data_i].FieldId == counterPktsKeyId {
			transformedMetrics = append(transformedMetrics, &model.MetricItem{
				SessionId:  sessionId,
				Value:      binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
				Type:       model.METRIC_PKTS,
				MetricName: tblName,
			})
		}
	}

	return transformedMetrics, nil
}

// Reset indirect counters on device given a list of sessionIds
func (driver *TofinoDriver) ResetCounter(sessionIds []uint32, shortTbleName string) {
	registerValueByteSize := 8
	allResetReq := make([]*bfruntime.Update, 0)
	// Build reset request
	for _, id := range sessionIds {
		resetReq, err := driver.getIndirectCounterResetRequest(shortTbleName, COUNTER_INDEX_KEY_NAME, id, []string{COUNTER_SPEC_BYTES, COUNTER_SPEC_PKTS}, registerValueByteSize)
		if err != nil {
			driver.logger.Error("cannot build bfrt reset request", "tblName", shortTbleName, "err", err)
			continue
		} else {
			allResetReq = append(allResetReq, &bfruntime.Update{
				Type:   bfruntime.Update_MODIFY,
				Entity: resetReq,
			})
		}
	}
	if len(allResetReq) > 0 {
		// Send reset requests
		err := driver.SendWriteRequest(allResetReq)
		if err != nil {
			driver.logger.Error("Register reset has failed", "tblName", shortTbleName, "err", err)
		}
	}
}
