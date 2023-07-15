package driver

import (
	"encoding/binary"
	"strings"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

// Retrieve Ingress End header byte counter by a list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressHdrStartCounter(sessionIds []uint32) ([]*bfruntime.Entity, error) {
	return driver.GetHdrSizeCounter(PROBE_INGRESS_START_HDR_SIZE, sessionIds)
}

// Retrieve Ingress End header byte counter by a list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressHdrEndCounter(sessionIds []uint32) ([]*bfruntime.Entity, error) {
	return driver.GetHdrSizeCounter(PROBE_INGRESS_END_HDR_SIZE, sessionIds)
}

// Retrieve Egress End packet byte counter by a list of sessionIds.
// Retrieve byte count from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetEgressEndCounter(sessionIds []uint32) ([]*bfruntime.Entity, error) {
	return driver.GetHdrSizeCounter(PROBE_EGRESS_END_CNT, sessionIds)
}

// Retrieve header byte counter by a short table name and list of sessionIds.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetHdrSizeCounter(shortTblName string, sessionIds []uint32) ([]*bfruntime.Entity, error) {
	driver.logger.Trace("Requesting header byte counter", "tblName", shortTblName, "sessionIds", sessionIds)

	if len(sessionIds) == 0 {
		return nil, nil
	}

	tblName := driver.FindTableNameByShortName(shortTblName)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: shortTblName}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	requests, err := driver.GetMetricFromRegisterRequest(registersToReq, model.METRIC_BYTES)

	return requests, err
}

// Collect ingress jitter value from register.
// The byte counter are retrieved from a 32-bit register as the stateful ALU supports only values up to 32-bits.
func (driver *TofinoDriver) GetIngressJitter(sessionIds []uint32) ([]*bfruntime.Entity, error) {
	driver.logger.Trace("Requesting ingress jitter", "sessionIds", sessionIds)

	if len(sessionIds) == 0 {
		return nil, nil
	}

	tblName := driver.FindTableNameByShortName(PROBE_INGRESS_JITTER_REGISTER)

	if tblName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_JITTER_REGISTER}
	}

	registersToReq := driver.transformSessionIdToAppRegister(sessionIds, tblName)

	requests, err := driver.GetMetricFromRegisterRequest(registersToReq, model.METRIC_BYTES)

	return requests, err
}

// Retrieves register values by a list of appRegister structs, which are used as index.
func (driver *TofinoDriver) GetMetricFromRegisterRequest(appRegisters []*model.AppRegister, metricType string) ([]*bfruntime.Entity, error) {
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

	return tblEntries, nil
}

func (driver *TofinoDriver) GetResetRegisterRequest(sessionIds []uint32) []*bfruntime.Update {
	allResetReq := make([]*bfruntime.Update, 0)
	shortTblNames := []string{PROBE_INGRESS_START_HDR_SIZE, PROBE_INGRESS_END_HDR_SIZE, PROBE_EGRESS_END_CNT}
	extraProbes := driver.GetExtraProbes()
	shortTblNames = append(extraProbes, shortTblNames...)
	// Build reset request
	for _, shortTblName := range shortTblNames {
		tblName := driver.FindTableNameByShortName(shortTblName)
		_, dataName := driver.GetSingletonDataIdLikeName(tblName, shortTblName)
		dataWidth := driver.GetSingletonDataWidthByName(tblName, dataName)
		dataWidth = dataWidth / 8
		for _, id := range sessionIds {
			resetReq, err := driver.getIndirectCounterResetRequest(shortTblName, REGISTER_INDEX_KEY_NAME, id, []string{shortTblName}, int(dataWidth))
			if err != nil {
				driver.logger.Error("cannot build bfrt reset request", "tblName", shortTblName, "err", err)
				continue
			} else {
				allResetReq = append(allResetReq, &bfruntime.Update{
					Type:   bfruntime.Update_MODIFY,
					Entity: resetReq,
				})
			}
		}
	}
	return allResetReq
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

func (driver *TofinoDriver) ProcessRegisterResponse(entity *bfruntime.Entity) (*model.MetricItem, error) {
	tblName := driver.GetTableNameById(entity.GetTableEntry().GetTableId())
	// Get sessionId from key field.
	sessionId := binary.BigEndian.Uint32(entity.GetTableEntry().GetKey().GetFields()[0].GetExact().GetValue())
	dataEntries := entity.GetTableEntry().GetData().GetFields()
	for data_i := range dataEntries {
		// Dataplane could return just a single byte instead of 4 bytes.
		// So we copy the response in a 4 byte slice.
		rawValue := dataEntries[data_i].GetStream()
		var decodedValue uint64
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

		// Replace full tblname with short name
		shortTblName := driver.FindShortTableNameByName(tblName)
		if shortTblName != "" {
			tblName = shortTblName
		}

		metricType := model.METRIC_EXT_VALUE
		switch tblName {
		case PROBE_INGRESS_START_HDR_SIZE, PROBE_INGRESS_END_HDR_SIZE, PROBE_EGRESS_END_CNT:
			metricType = model.METRIC_BYTES
		default:
			metricType = model.METRIC_EXT_VALUE
		}

		return &model.MetricItem{
			SessionId:  sessionId,
			Value:      uint64(decodedValue),
			Type:       metricType,
			MetricName: tblName,
		}, nil
	}

	return nil, nil
}
