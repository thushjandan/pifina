package driver

import (
	"encoding/binary"
	"sort"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

// Retrieve all MatchSelectorEntries from device
func (driver *TofinoDriver) GetMatchSelectorEntriesRequest() ([]*bfruntime.Entity, error) {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	tblEntries := []*bfruntime.Entity{}

	tblEntries = append(tblEntries,
		&bfruntime.Entity{
			Entity: &bfruntime.Entity_TableEntry{
				TableEntry: &bfruntime.TableEntry{
					IsDefaultEntry: false,
					TableId:        tblId,
					TableFlags: &bfruntime.TableFlags{
						FromHw: true,
					},
				},
			},
		},
	)
	return tblEntries, nil
}

func (driver *TofinoDriver) ProcessMatchActionResponse(entity *bfruntime.Entity) ([]*model.MetricItem, error) {
	tblName := driver.GetTableNameById(entity.GetTableEntry().GetTableId())
	dataEntries := entity.GetTableEntry().GetData().GetFields()
	// Skip default entry
	if len(dataEntries) < 3 {
		return nil, nil
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	if actionName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find full action name for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	// Get key Ids
	counterBytesKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_BYTES)
	counterPktsKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_PKTS)
	sessionIdDataId := driver.GetDataIdByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)

	if sessionIdDataId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find field id for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID}
	}

	transformedMetrics := make([]*model.MetricItem, 0)
	sessionId := uint32(0)
	for data_i := range dataEntries {
		if dataEntries[data_i].GetFieldId() == sessionIdDataId {
			rawValue := dataEntries[data_i].GetStream()
			buffer := make([]byte, 4)
			copy(buffer[len(buffer)-len(rawValue):], rawValue)
			// Parse to uint32
			sessionId = binary.BigEndian.Uint32(buffer)
		}
		// If the key indicates a byte counter
		if dataEntries[data_i].GetFieldId() == counterBytesKeyId {
			transformedMetrics = append(transformedMetrics, &model.MetricItem{
				SessionId:  sessionId,
				Value:      binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
				Type:       model.METRIC_BYTES,
				MetricName: PROBE_INGRESS_MATCH_CNT,
			})
		}
		// If the key indicates a packet counter
		if dataEntries[data_i].GetFieldId() == counterPktsKeyId {
			transformedMetrics = append(transformedMetrics, &model.MetricItem{
				SessionId:  sessionId,
				Value:      binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
				Type:       model.METRIC_PKTS,
				MetricName: PROBE_INGRESS_MATCH_CNT,
			})
		}
	}

	return transformedMetrics, nil
}

func (driver *TofinoDriver) GetResetTableSelectorRequests(selectorEntries []*model.MatchSelectorEntry) ([]*bfruntime.Update, error) {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	// Get key Ids
	counterBytesKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_BYTES)
	counterPktsKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_PKTS)

	// Convert to byte slice
	zeroValue := make([]byte, 8)

	dataFields := []*bfruntime.DataField{
		{
			FieldId: counterBytesKeyId,
			Value: &bfruntime.DataField_Stream{
				Stream: zeroValue,
			},
		},
		{
			FieldId: counterPktsKeyId,
			Value: &bfruntime.DataField_Stream{
				Stream: zeroValue,
			},
		},
	}

	updateRequests := []*bfruntime.Update{}

	for i := range selectorEntries {
		keyFields := []*bfruntime.KeyField{}
		for _, keyItem := range selectorEntries[i].Keys {
			switch keyItem.MatchType {
			case model.MATCH_TYPE_EXACT:
				keyFields = append(keyFields, &bfruntime.KeyField{
					FieldId: keyItem.FieldId,
					MatchType: &bfruntime.KeyField_Exact_{
						Exact: &bfruntime.KeyField_Exact{
							Value: keyItem.Value,
						},
					},
				})
			case model.MATCH_TYPE_TERNARY:
				keyFields = append(keyFields, &bfruntime.KeyField{
					FieldId: keyItem.FieldId,
					MatchType: &bfruntime.KeyField_Ternary_{
						Ternary: &bfruntime.KeyField_Ternary{
							Value: keyItem.Value,
							Mask:  keyItem.ValueMask,
						},
					},
				})
			case model.MATCH_TYPE_LPM:
				keyFields = append(keyFields, &bfruntime.KeyField{
					FieldId: keyItem.FieldId,
					MatchType: &bfruntime.KeyField_Lpm{
						Lpm: &bfruntime.KeyField_LPM{
							Value:     keyItem.Value,
							PrefixLen: keyItem.PrefixLength,
						},
					},
				})
			}
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

		updateRequests = append(updateRequests, &bfruntime.Update{
			Type:   bfruntime.Update_MODIFY,
			Entity: tblEntry,
		})
	}

	return updateRequests, nil
}

// Retrieve all MatchSelectorEntries from device
func (driver *TofinoDriver) GetMatchSelectorEntries() ([]*bfruntime.Entity, error) {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	tblEntries := []*bfruntime.Entity{}

	tblEntries = append(tblEntries,
		&bfruntime.Entity{
			Entity: &bfruntime.Entity_TableEntry{
				TableEntry: &bfruntime.TableEntry{
					IsDefaultEntry: false,
					TableId:        tblId,
					TableFlags: &bfruntime.TableFlags{
						FromHw: true,
					},
				},
			},
		},
	)
	entities, err := driver.SendReadRequest(tblEntries)
	return entities, err
}

// Retrieve a list of configured session Id from device
func (driver *TofinoDriver) GetKeysFromMatchSelectors() ([]*model.MatchSelectorEntry, error) {
	entries, err := driver.GetMatchSelectorEntries()
	if err != nil {
		return nil, err
	}

	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	if actionName == "" {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find full action name for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	sessionIdDataId := driver.GetDataIdByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if sessionIdDataId == 0 {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find field id for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID}
	}

	matchSelectorEntries := make([]*model.MatchSelectorEntry, 0, len(entries))
	for i := range entries {
		matchSelectorEntry := &model.MatchSelectorEntry{}
		keyFields := entries[i].GetTableEntry().GetKey().GetFields()
		matchSelectorKeys := make([]*model.MatchSelectorKey, 0, len(keyFields))
		for key_i := range keyFields {
			matchSelectorKey := &model.MatchSelectorKey{
				FieldId: keyFields[key_i].GetFieldId(),
			}
			switch matchType := keyFields[key_i].GetMatchType().(type) {
			case *bfruntime.KeyField_Exact_:
				matchSelectorKey.Value = matchType.Exact.GetValue()
				matchSelectorKey.MatchType = model.MATCH_TYPE_EXACT
			case *bfruntime.KeyField_Ternary_:
				matchSelectorKey.Value = matchType.Ternary.GetValue()
				matchSelectorKey.MatchType = model.MATCH_TYPE_TERNARY
				matchSelectorKey.ValueMask = matchType.Ternary.GetMask()
			case *bfruntime.KeyField_Lpm:
				matchSelectorKey.Value = matchType.Lpm.GetValue()
				matchSelectorKey.MatchType = model.MATCH_TYPE_LPM
				matchSelectorKey.PrefixLength = matchType.Lpm.GetPrefixLen()
			}
			matchSelectorKeys = append(matchSelectorKeys, matchSelectorKey)
		}
		actionFields := entries[i].GetTableEntry().GetData().GetFields()
		for action_i := range actionFields {
			// Search for sessionId field
			if actionFields[action_i].GetFieldId() == sessionIdDataId {
				rawValue := actionFields[action_i].GetStream()
				buffer := make([]byte, 4)
				copy(buffer[len(buffer)-len(rawValue):], rawValue)
				// Parse to uint32
				parsedSessionId := binary.BigEndian.Uint32(buffer)
				matchSelectorEntry.SessionId = parsedSessionId
				matchSelectorEntry.Keys = matchSelectorKeys
				matchSelectorEntries = append(matchSelectorEntries, matchSelectorEntry)
				break
			}
		}
	}

	// Sort the sessionIds
	sort.Slice(matchSelectorEntries, func(i, j int) bool { return matchSelectorEntries[i].SessionId < matchSelectorEntries[j].SessionId })

	return matchSelectorEntries, err
}

// Returns the width of the sessionId parameter
// Needed to generate new sessionId or to define the size of the bufferpool
func (driver *TofinoDriver) GetSessionIdBitWidth() (uint32, error) {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return 0, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return 0, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	if actionName == "" {
		return 0, &model.ErrNameNotFound{Msg: "Cannot find full action name for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	sessionIdWidth := driver.GetActionDataWidthByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if sessionIdWidth < 1 {
		return 0, &model.ErrNameNotFound{Msg: "Cannot find sessionId width on the device", Entity: tblName}
	}

	return sessionIdWidth, nil
}

func (driver *TofinoDriver) AddSelectorEntry(newEntry *model.MatchSelectorEntry) error {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	actionId := driver.GetActionIdByName(tblName, actionName)
	if actionId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find action name", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	dataId := driver.GetDataIdByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if dataId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find action param name", Entity: PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID}
	}

	keyFields := []*bfruntime.KeyField{}

	for _, keyItem := range newEntry.Keys {
		switch keyItem.MatchType {
		case model.MATCH_TYPE_EXACT:
			keyFields = append(keyFields, &bfruntime.KeyField{
				FieldId: keyItem.FieldId,
				MatchType: &bfruntime.KeyField_Exact_{
					Exact: &bfruntime.KeyField_Exact{
						Value: keyItem.Value,
					},
				},
			})
		case model.MATCH_TYPE_TERNARY:
			keyFields = append(keyFields, &bfruntime.KeyField{
				FieldId: keyItem.FieldId,
				MatchType: &bfruntime.KeyField_Ternary_{
					Ternary: &bfruntime.KeyField_Ternary{
						Value: keyItem.Value,
						Mask:  keyItem.ValueMask,
					},
				},
			})
		case model.MATCH_TYPE_LPM:
			keyFields = append(keyFields, &bfruntime.KeyField{
				FieldId: keyItem.FieldId,
				MatchType: &bfruntime.KeyField_Lpm{
					Lpm: &bfruntime.KeyField_LPM{
						Value:     keyItem.Value,
						PrefixLen: keyItem.PrefixLength,
					},
				},
			})
		}
	}

	// Convert to byte slice
	byteSessionId := make([]byte, 4)
	binary.BigEndian.PutUint32(byteSessionId, newEntry.SessionId)
	sessionIdWidth, err := driver.GetSessionIdBitWidth()
	if err != nil {
		return err
	}
	// Calculate, which from where to select the bytes
	byteArrayWidth := len(byteSessionId) - ((int(sessionIdWidth) / 8) + 1)

	dataFields := []*bfruntime.DataField{
		{
			FieldId: dataId,
			Value: &bfruntime.DataField_Stream{
				Stream: byteSessionId[byteArrayWidth:],
			},
		},
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
					ActionId: actionId,
					Fields:   dataFields,
				},
			},
		},
	}

	updateReq := []*bfruntime.Update{
		{
			Type:   bfruntime.Update_INSERT,
			Entity: tblEntry,
		},
	}

	err = driver.SendWriteRequest(updateReq)
	if err != nil {
		return err
	}

	return nil
}

func (driver *TofinoDriver) RemoveSelectorEntry(entry *model.MatchSelectorEntry) error {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)

	actionId := driver.GetActionIdByName(tblName, actionName)
	if actionId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find action name", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	dataId := driver.GetDataIdByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if dataId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find action param name", Entity: PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID}
	}

	keyFields := []*bfruntime.KeyField{}

	for _, keyItem := range entry.Keys {
		switch keyItem.MatchType {
		case model.MATCH_TYPE_EXACT:
			keyFields = append(keyFields, &bfruntime.KeyField{
				FieldId: keyItem.FieldId,
				MatchType: &bfruntime.KeyField_Exact_{
					Exact: &bfruntime.KeyField_Exact{
						Value: keyItem.Value,
					},
				},
			})
		case model.MATCH_TYPE_TERNARY:
			keyFields = append(keyFields, &bfruntime.KeyField{
				FieldId: keyItem.FieldId,
				MatchType: &bfruntime.KeyField_Ternary_{
					Ternary: &bfruntime.KeyField_Ternary{
						Value: keyItem.Value,
						Mask:  keyItem.ValueMask,
					},
				},
			})
		case model.MATCH_TYPE_LPM:
			keyFields = append(keyFields, &bfruntime.KeyField{
				FieldId: keyItem.FieldId,
				MatchType: &bfruntime.KeyField_Lpm{
					Lpm: &bfruntime.KeyField_LPM{
						Value:     keyItem.Value,
						PrefixLen: keyItem.PrefixLength,
					},
				},
			})
		}
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
			},
		},
	}

	updateReq := []*bfruntime.Update{
		{
			Type:   bfruntime.Update_DELETE,
			Entity: tblEntry,
		},
	}

	// Send delete request
	err := driver.SendWriteRequest(updateReq)
	if err != nil {
		return err
	}

	return nil
}

func (driver *TofinoDriver) GetIngressStartMatchSelectorSchema() ([]*model.MatchSelectorSchema, error) {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Create a DTO
		keys := make([]*model.MatchSelectorSchema, 0, len(driver.P4Tables[sliceIdx].Key))
		for key_i := range driver.P4Tables[sliceIdx].Key {
			keys = append(keys, &model.MatchSelectorSchema{
				FieldId:   driver.P4Tables[sliceIdx].Key[key_i].Id,
				Name:      driver.P4Tables[sliceIdx].Key[key_i].Name,
				MatchType: driver.P4Tables[sliceIdx].Key[key_i].MatchType,
				Type:      driver.P4Tables[sliceIdx].Key[key_i].Type.Type,
				Width:     driver.P4Tables[sliceIdx].Key[key_i].Type.Width,
			})
		}
		return keys, nil
	}

	return nil, &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
}
