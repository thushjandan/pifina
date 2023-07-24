package driver

import (
	"encoding/binary"
	"strings"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

// Retrieve Ingress End header byte counter by a list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressHdrStartCounter(sessionIds []uint32) ([]*model.MetricItem, error) {
	return driver.GetHdrSizeCounter(PROBE_INGRESS_START_HDR_SIZE, sessionIds)
}

// Retrieve Ingress End header byte counter by a list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressHdrEndCounter(sessionIds []uint32) ([]*model.MetricItem, error) {
	return driver.GetHdrSizeCounter(PROBE_INGRESS_END_HDR_SIZE, sessionIds)
}

// Retrieve Egress End packet byte counter by a list of sessionIds.
// Retrieve byte count from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetEgressEndCounter(sessionIds []uint32) ([]*model.MetricItem, error) {
	return driver.GetHdrSizeCounter(PROBE_EGRESS_END_CNT, sessionIds)
}

// Retrieve header byte counter by a short table name and list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetHdrSizeCounter(shortTblName string, sessionIds []uint32) ([]*model.MetricItem, error) {
	driver.logger.Trace("Requesting header byte counter", "tblName", shortTblName, "sessionIds", sessionIds)

	if len(sessionIds) == 0 {
		return nil, nil
	}

	tblName := driver.FindTableNameByShortName(shortTblName)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: shortTblName}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	// Retrieve register values for selected sessionId
	metrics, err := driver.GetMetricFromRegister(registersToReq, model.METRIC_BYTES)
	// If no errors have occured, reset the register
	if err == nil {
		// Reset register values
		driver.ResetRegister(sessionIds, shortTblName)
		for i := range metrics {
			metrics[i].MetricName = shortTblName
		}
	}

	return metrics, err
}

// Collect ingress jitter value from register.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressJitter(sessionIds []uint32) ([]*model.MetricItem, error) {
	driver.logger.Trace("Requesting ingress jitter", "sessionIds", sessionIds)

	if len(sessionIds) == 0 {
		return nil, nil
	}

	tblName := driver.FindTableNameByShortName(PROBE_INGRESS_JITTER_REGISTER)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_JITTER_REGISTER}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	// Retrieve register values for selected sessionId
	metrics, err := driver.GetMetricFromRegister(registersToReq, model.METRIC_EXT_VALUE)
	// If no errors have occured, reset the register
	if err == nil {
		for i := range metrics {
			metrics[i].MetricName = PROBE_INGRESS_JITTER_REGISTER
			// Convert ns to microsecond
			metrics[i].Value = metrics[i].Value / 1000
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
		tblName := driver.GetTableNameById(entities[i].GetTableEntry().GetTableId())

		for data_i := range dataEntries {
			// Dataplane could return just a single byte instead of 4 bytes.
			// So we copy the response in a 4 byte slice.
			rawValue := dataEntries[data_i].GetStream()
			var decodedValue uint64
			// Check if data value is 64-bit or 32 bit
			if len(rawValue) == 8 {
				if strings.Contains(tblName, PROBE_INGRESS_JITTER_REGISTER) {
					buffer := make([]byte, 4)
					copy(buffer[:], rawValue[0:3])
					decodedValue = uint64(binary.BigEndian.Uint32(buffer))
				} else {
					decodedValue = binary.BigEndian.Uint64(rawValue)
				}
			} else {
				buffer := make([]byte, 4)
				copy(buffer[len(buffer)-len(rawValue):], rawValue)
				decodedValue = uint64(binary.BigEndian.Uint32(buffer))
			}

			// Skip loop if value is 0
			if decodedValue == 0 && data_i != 0 {
				continue
			}

			transformedMetrics = append(transformedMetrics, &model.MetricItem{
				SessionId:   sessionId,
				Value:       decodedValue,
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
