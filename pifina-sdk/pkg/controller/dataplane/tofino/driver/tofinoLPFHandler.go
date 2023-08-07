// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package driver

import (
	"encoding/binary"

	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/model"
)

// Configure all LPF instances with the correct parameter (time constants)
func (driver *TofinoDriver) ConfigureLPF(sessionIds []uint32) error {
	tblName, ok := driver.probeTableMap[PROBE_INGRESS_JITTER_LPF]
	if !ok {
		return &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: PROBE_INGRESS_JITTER_LPF}
	}

	tblId := driver.GetTableIdByName(tblName)
	if tblId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find table name for the probe", Entity: tblName}
	}

	keyId := driver.GetKeyIdByName(tblName, LPF_INDEX_KEY_NAME)
	if keyId == 0 {
		return &model.ErrNameNotFound{Msg: "Cannot find key id for table name", Entity: tblName}
	}

	// Get key Ids
	lpfSampleKeyId := driver.GetSingletonDataIdByName(tblName, LPF_SPEC_TYPE)
	lpfGainKeyId := driver.GetSingletonDataIdByName(tblName, LPF_GAIN_TIME)
	lpfDecayKeyId := driver.GetSingletonDataIdByName(tblName, LPF_DECAY_TIME)
	lpfScaleDownKeyId := driver.GetSingletonDataIdByName(tblName, LPF_SCALE_DOWN_FACTOR)
	lpfTimeConst := float32(80)

	byteScaleDown := make([]byte, 4)
	binary.BigEndian.PutUint32(byteScaleDown, 0)
	// Set the LPF parameter
	dataFields := []*bfruntime.DataField{
		{
			FieldId: lpfSampleKeyId,
			Value: &bfruntime.DataField_StrVal{
				StrVal: "SAMPLE",
			},
		},
		{
			FieldId: lpfGainKeyId,
			Value: &bfruntime.DataField_FloatVal{
				FloatVal: lpfTimeConst,
			},
		},
		{
			FieldId: lpfDecayKeyId,
			Value: &bfruntime.DataField_FloatVal{
				FloatVal: lpfTimeConst,
			},
		},
		{
			FieldId: lpfScaleDownKeyId,
			Value: &bfruntime.DataField_Stream{
				Stream: byteScaleDown,
			},
		},
	}

	updateReq := []*bfruntime.Update{}

	for _, sessionId := range sessionIds {
		// Convert to byte slice
		byteEntryId := make([]byte, 4)
		binary.BigEndian.PutUint32(byteEntryId, sessionId)

		keyFields := []*bfruntime.KeyField{
			{
				FieldId: keyId,
				MatchType: &bfruntime.KeyField_Exact_{
					Exact: &bfruntime.KeyField_Exact{
						Value: byteEntryId,
					},
				},
			},
		}

		updateReq = append(updateReq, &bfruntime.Update{
			Type: bfruntime.Update_MODIFY,
			Entity: &bfruntime.Entity{
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
			},
		})
	}

	err := driver.SendWriteRequest(updateReq)
	if err != nil {
		return err
	}
	driver.logger.Info("LPF has been configured", "sessionIds", sessionIds)

	return nil
}
