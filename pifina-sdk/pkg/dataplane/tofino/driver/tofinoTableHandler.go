package driver

import (
	"encoding/binary"
	"sort"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
)

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
					TableId: tblId,
				},
			},
		},
	)
	entities, err := driver.SendReadRequest(tblEntries)
	return entities, err
}

// Retrieve all MatchSelectorEntries from device
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

	sessions := make([]uint32, len(entries))
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
