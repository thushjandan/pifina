package driver

import (
	"encoding/binary"
	"sort"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
)

func (driver *TofinoDriver) GetIngressStartMatchSelectorCounter() ([]*MetricItem, error) {
	driver.logger.Debug("Requesting ingress start match selector counter")
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

	// Transform response
	transformedMetrics := make([]*MetricItem, 0, len(entities))
	for i := range entities {
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
				transformedMetrics = append(transformedMetrics, &MetricItem{
					SessionId:  sessionId,
					Value:      binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
					Type:       METRIC_BYTES,
					MetricName: PROBE_INGRESS_MATCH_CNT,
				})
			}
			// If the key indicates a packet counter
			if dataEntries[data_i].GetFieldId() == counterPktsKeyId {
				transformedMetrics = append(transformedMetrics, &MetricItem{
					SessionId:  sessionId,
					Value:      binary.BigEndian.Uint64(dataEntries[data_i].GetStream()),
					Type:       METRIC_PKTS,
					MetricName: PROBE_INGRESS_MATCH_CNT,
				})
			}
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
				},
			},
		},
	)
	entities, err := driver.SendReadRequest(tblEntries)
	return entities, err
}

// Retrieve a list of configured session Id from device
func (driver *TofinoDriver) GetSessionsFromMatchSelectors() ([]uint32, error) {
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

	sessions := make([]uint32, 0, len(entries))
	for i := range entries {
		actionFields := entries[i].GetTableEntry().GetData().GetFields()
		for action_i := range actionFields {
			// Search for sessionId field
			if actionFields[action_i].GetFieldId() == sessionIdDataId {
				rawValue := actionFields[action_i].GetStream()
				buffer := make([]byte, 4)
				copy(buffer[len(buffer)-len(rawValue):], rawValue)
				// Parse to uint32
				sessions = append(sessions, binary.BigEndian.Uint32(buffer))
				break
			}
		}
	}

	// Sort the sessionIds
	sort.Slice(sessions, func(i, j int) bool { return sessions[i] < sessions[j] })

	return sessions, err
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

	sessionIdWidth := driver.GetActionDataWidthByName(tblName, PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID)
	if sessionIdWidth == 0 {
		return 0, &ErrNameNotFound{Msg: "Cannot find sessionId width on the device", Entity: tblName}
	}

	return sessionIdWidth, nil
}
