package driver

import (
	"encoding/binary"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

// Retrieve Ingress End header byte counter by a list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressHdrStartCounter(sessionIds []uint32) ([]*model.MetricItem, error) {
	driver.logger.Trace("Requesting Ingress start header byte counter", "sessionIds", sessionIds)

	if len(sessionIds) == 0 {
		return nil, nil
	}
	tblName := driver.FindTableNameByShortName(PROBE_INGRESS_START_HDR_SIZE)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_START_HDR_SIZE}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	// Retrieve register values for selected sessionId
	metrics, err := driver.GetMetricFromRegister(registersToReq, model.METRIC_BYTES)
	// If no errors have occured, reset the register
	if err == nil {
		// Reset register values
		driver.ResetRegister(sessionIds, PROBE_INGRESS_START_HDR_SIZE)
		for i := range metrics {
			metrics[i].MetricName = PROBE_INGRESS_START_HDR_SIZE
		}
	}

	return metrics, err
}

// Retrieve Ingress End header byte counter by a list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressHdrEndCounter(sessionIds []uint32) ([]*model.MetricItem, error) {
	driver.logger.Trace("Requesting Ingress end header byte counter", "sessionIds", sessionIds)

	if len(sessionIds) == 0 {
		return nil, nil
	}
	tblName := driver.FindTableNameByShortName(PROBE_INGRESS_END_HDR_SIZE)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_END_HDR_SIZE}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	metrics, err := driver.GetMetricFromRegister(registersToReq, model.METRIC_BYTES)
	// If no errors have occured, reset the register
	if err == nil {
		// Reset register values
		driver.ResetRegister(sessionIds, PROBE_INGRESS_END_HDR_SIZE)
		for i := range metrics {
			metrics[i].MetricName = PROBE_INGRESS_END_HDR_SIZE
		}
	}

	return metrics, err
}

// Retrieve Egress End packet byte counter by a list of sessionIds.
// Retrieve byte count from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetEgressEndCounter(sessionIds []uint32) ([]*model.MetricItem, error) {
	driver.logger.Trace("Requesting Egress end byte counter", "sessionIds", sessionIds)
	if len(sessionIds) == 0 {
		return nil, nil
	}
	tblName := driver.FindTableNameByShortName(PROBE_EGRESS_END_CNT)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_EGRESS_END_CNT}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	metrics, err := driver.GetMetricFromRegister(registersToReq, model.METRIC_BYTES)
	// If no errors have occured, reset the register
	if err == nil {
		// Reset register values
		driver.ResetRegister(sessionIds, PROBE_EGRESS_END_CNT)
		for i := range metrics {
			metrics[i].MetricName = PROBE_EGRESS_END_CNT
		}
	}

	return metrics, err
}

// Retrieves register values by a list of appRegister structs, which are used as index.
func (driver *TofinoDriver) GetMetricFromRegister(appRegisters []*model.AppRegister, metricType string) ([]*model.MetricItem, error) {
	tblEntries := []*bfruntime.Entity{}

	for i := range appRegisters {
		tblId := driver.GetTableIdByName(appRegisters[i].Name)
		if tblId == 0 {
			return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: appRegisters[i].Name}
		}

		keyId := driver.GetKeyIdByName(appRegisters[i].Name, REGISTER_INDEX_KEY_NAME)
		if keyId == 0 {
			return nil, &model.ErrNameNotFound{Msg: "Cannot find key id for table name", Entity: appRegisters[i].Name}
		}

		// Convert to byte slice
		byteEntryId := make([]byte, 4)
		binary.BigEndian.PutUint32(byteEntryId, appRegisters[i].Index)

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

	// Send read request to switch.
	entities, err := driver.SendReadRequest(tblEntries)
	if err != nil {
		return nil, err
	}

	// Transform response
	transformedMetrics := make([]*model.MetricItem, 0, len(entities))
	timeNow := time.Now()
	for i := range entities {
		// Get sessionId from key field.
		sessionId := binary.BigEndian.Uint32(entities[i].GetTableEntry().GetKey().GetFields()[0].GetExact().GetValue())
		dataEntries := entities[i].GetTableEntry().GetData().GetFields()
		for data_i := range dataEntries {
			// Dataplane could return just a single byte instead of 4 bytes.
			// So we copy the response in a 4 byte slice.
			rawValue := dataEntries[data_i].GetStream()
			buffer := make([]byte, 4)
			copy(buffer[len(buffer)-len(rawValue):], rawValue)

			decodedValue := binary.BigEndian.Uint32(buffer)
			// Skip loop if value is 0
			if decodedValue == 0 {
				continue
			}
			tblName := driver.GetTableNameById(entities[i].GetTableEntry().GetTableId())
			transformedMetrics = append(transformedMetrics, &model.MetricItem{
				SessionId:   sessionId,
				Value:       uint64(decodedValue),
				Type:        metricType,
				MetricName:  tblName,
				LastUpdated: timeNow,
			})
		}
	}

	return transformedMetrics, nil
}

func (driver *TofinoDriver) ResetRegister(sessionIds []uint32, shortTbleName string) {
	registerValueByteSize := 4
	allResetReq := make([]*bfruntime.Update, 0)
	// Build reset request
	for _, id := range sessionIds {
		resetReq, err := driver.getIndirectCounterResetRequest(shortTbleName, REGISTER_INDEX_KEY_NAME, id, []string{shortTbleName}, registerValueByteSize)
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

// Returns from cache all available registers on device.
func (driver *TofinoDriver) GetAllRegisterNames() []string {
	registerNames := make([]string, 0)

	for i := range driver.P4Tables {
		if driver.P4Tables[i].TableType == TABLE_TYPE_REGISTER {
			registerNames = append(registerNames, driver.P4Tables[i].Name)
		}
	}

	return registerNames
}

// Converts a list of sessionIds to a list of AppRegister structs.
func (driver *TofinoDriver) transformSessionIdToAppRegister(sessionIds []uint32, tblName string) []*model.AppRegister {
	registerToRequest := make([]*model.AppRegister, 0, len(sessionIds))
	for i := range sessionIds {
		registerToRequest = append(registerToRequest, &model.AppRegister{
			Name:  tblName,
			Index: sessionIds[i],
		})
	}

	return registerToRequest
}
