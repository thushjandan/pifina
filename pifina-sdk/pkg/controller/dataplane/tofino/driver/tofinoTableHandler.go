package driver

import (
	"encoding/binary"
	"sort"
	"time"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

func (driver *TofinoDriver) GetIngressStartMatchSelectorCounter() ([]*model.MetricItem, error) {
	driver.logger.Trace("Requesting ingress start match selector counter")
	entities, err := driver.GetMatchSelectorEntries()
	if err != nil {
		return nil, err
	}

	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	if actionName == "" {
		return nil, &ErrNameNotFound{Msg: "Cannot find full action name for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	// Get key Ids
	counterBytesKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_BYTES)
	counterPktsKeyId := driver.GetSingletonDataIdByName(tblName, COUNTER_SPEC_PKTS)
	sessionIdDataId := driver.GetDataIdByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)

	if sessionIdDataId == 0 {
		return nil, &ErrNameNotFound{Msg: "Cannot find field id for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID}
	}

	transformedMetrics := make([]*model.MetricItem, 0, len(entities))
	updateRequests := make([]*bfruntime.Update, 0, len(entities))
	timeNow := time.Now()
	// Transform response
	for i := range entities {
		isCounterEntry := false
		dataEntries := entities[i].GetTableEntry().GetData().GetFields()
		// Skip default entry
		if len(dataEntries) < 3 {
			continue
		}
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
					SessionId:   sessionId,
					Value:       binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
					Type:        model.METRIC_BYTES,
					MetricName:  PROBE_INGRESS_MATCH_CNT,
					LastUpdated: timeNow,
				})
				// Prepare the reset counter request
				isCounterEntry = true
				dataEntries[data_i].Value = &bfruntime.DataField_Stream{
					Stream: make([]byte, 8),
				}
			}
			// If the key indicates a packet counter
			if dataEntries[data_i].GetFieldId() == counterPktsKeyId {
				transformedMetrics = append(transformedMetrics, &model.MetricItem{
					SessionId:   sessionId,
					Value:       binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
					Type:        model.METRIC_PKTS,
					MetricName:  PROBE_INGRESS_MATCH_CNT,
					LastUpdated: timeNow,
				})
				// Prepare the reset counter request
				isCounterEntry = true
				dataEntries[data_i].Value = &bfruntime.DataField_Stream{
					Stream: make([]byte, 8),
				}
			}
		}
		// If the entry contains any
		if isCounterEntry {
			updateRequests = append(updateRequests, &bfruntime.Update{
				Type: bfruntime.Update_MODIFY,
				Entity: &bfruntime.Entity{
					Entity: &bfruntime.Entity_TableEntry{
						TableEntry: entities[i].GetTableEntry(),
					},
				},
			})
		}
	}

	if len(updateRequests) > 0 {
		err := driver.SendWriteRequest(updateRequests)
		if err != nil {
			driver.logger.Error("Error occured during table counter reset.", "err", err)
		}
	}

	return transformedMetrics, nil
}

// Retrieve all MatchSelectorEntries from device
func (driver *TofinoDriver) GetMatchSelectorEntries() ([]*bfruntime.Entity, error) {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_MATCH_CNT]
	if !ok {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
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
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	if actionName == "" {
		return nil, &ErrNameNotFound{Msg: "Cannot find full action name for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	sessionIdDataId := driver.GetDataIdByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if sessionIdDataId == 0 {
		return nil, &ErrNameNotFound{Msg: "Cannot find field id for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID}
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
				matchSelectorKey.MatchType = "Exact"
			case *bfruntime.KeyField_Ternary_:
				matchSelectorKey.Value = matchType.Ternary.GetValue()
				matchSelectorKey.MatchType = "Ternary"
				matchSelectorKey.ValueMask = matchType.Ternary.GetMask()
			case *bfruntime.KeyField_Lpm:
				matchSelectorKey.Value = matchType.Lpm.GetValue()
				matchSelectorKey.MatchType = "LPM"
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
		return 0, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return 0, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	actionName := driver.FindFullActionName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME)
	if actionName == "" {
		return 0, &ErrNameNotFound{Msg: "Cannot find full action name for the match selector", Entity: PROBE_INGRESS_MATCH_ACTION_NAME}
	}

	sessionIdWidth := driver.GetActionDataWidthByName(tblName, actionName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if sessionIdWidth < 1 {
		return 0, &ErrNameNotFound{Msg: "Cannot find sessionId width on the device", Entity: tblName}
	}

	return sessionIdWidth, nil
}
