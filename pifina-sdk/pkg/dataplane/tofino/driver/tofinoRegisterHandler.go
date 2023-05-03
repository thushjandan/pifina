package driver

import (
	"encoding/binary"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
)

func (driver *TofinoDriver) GetEgressEndCounter(sessionIds []uint32) ([]*MetricItem, error) {
	if len(sessionIds) == 0 {
		driver.logger.Debug("Given list of session ids is empty. Skipping collecting egress end counter.")
		return nil, nil
	}

	tblName := driver.FindTableNameByShortName(PROBE_EGRESS_END_CNT)

	if tblName == "" {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_EGRESS_END_CNT}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return nil, &ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	keyId := driver.GetKeyIdByName(tblName, REGISTER_INDEX_KEY_NAME)
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
						TableId:        tblId,
						IsDefaultEntry: false,
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
	driver.logger.Debug("Requesting Egress start counter", "sessionIds", sessionIds)

	// Send read request to switch.
	entities, err := driver.SendReadRequest(tblEntries)
	if err != nil {
		return nil, err
	}

	// Transform response
	transformedMetrics := make([]*MetricItem, 0, len(entities))
	for i := range entities {
		// Get sessionId from key field.
		sessionId := binary.BigEndian.Uint32(entities[i].GetTableEntry().GetKey().GetFields()[0].GetExact().GetValue())
		dataEntries := entities[i].GetTableEntry().GetData().GetFields()
		for data_i := range dataEntries {
			decodedValue := binary.BigEndian.Uint32(dataEntries[data_i].GetStream())
			// Skip loop if value is 0
			if decodedValue == 0 {
				continue
			}
			transformedMetrics = append(transformedMetrics, &MetricItem{
				SessionId: sessionId,
				Value:     uint64(decodedValue),
				Type:      METRIC_BYTES,
			})
		}
	}

	return transformedMetrics, nil
}
