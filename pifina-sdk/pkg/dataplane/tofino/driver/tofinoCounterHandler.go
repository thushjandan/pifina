package driver

import (
	"encoding/binary"
	"fmt"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
)

func (driver *TofinoDriver) GetEgressStartCounter(sessionIds []uint32) ([]*MetricItem, error) {
	if len(sessionIds) == 0 {
		driver.logger.Debug("Given list of session ids is empty. Skipping collecting egress start counter.")
		return nil, nil
	}

	tblName, ok := driver.probeTableMap[PROBE_EGRESS_START_CNT]
	if !ok {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_MATCH_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	keyId := driver.GetKeyIdByName(tblName, REGISTER_KEY_NAME)
	if keyId == 0 {
		return nil, &ErrNameNotFound{Msg: "Cannot find key id for table name", Entity: tblName}
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
						TableId: tblId,
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

	entities, err := driver.SendReadRequest(tblEntries)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v", entities)

	return nil, nil

}
