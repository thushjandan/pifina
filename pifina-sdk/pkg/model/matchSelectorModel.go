package model

import (
	"encoding/hex"
	"encoding/json"
)

type MatchSelectorSchema struct {
	FieldId   uint32
	Name      string
	MatchType string
	Type      string
	Width     uint32
}

type MatchSelectorEntry struct {
	SessionId uint32              `json:"sessionId"`
	Keys      []*MatchSelectorKey `json:"keys"`
}

type MatchSelectorKey struct {
	FieldId      uint32 `json:"fieldId"`
	Value        []byte `json:"value"`
	MatchType    string `json:"matchType"`
	ValueMask    []byte `json:"valueMask,omitempty"`
	PrefixLength int32  `json:"valueLpm,omitempty"`
}

const (
	MATCH_TYPE_EXACT   = "Exact"
	MATCH_TYPE_TERNARY = "Ternary"
	MATCH_TYPE_LPM     = "LPM"
)

func (key *MatchSelectorKey) MarshalJSON() ([]byte, error) {
	type Alias MatchSelectorKey
	return json.Marshal(&struct {
		Value     string `json:"value"`
		ValueMask string `json:"valueMask,omitempty"`
		*Alias
	}{
		Value:     hex.EncodeToString(key.Value),
		ValueMask: hex.EncodeToString(key.ValueMask),
		Alias:     (*Alias)(key),
	})
}

func (key *MatchSelectorKey) UnmarshalJSON(data []byte) error {
	type Alias MatchSelectorKey
	aux := &struct {
		Value     string `json:"value"`
		ValueMask string `json:"valueMask,omitempty"`
		*Alias
	}{
		Value:     hex.EncodeToString(key.Value),
		ValueMask: hex.EncodeToString(key.ValueMask),
		Alias:     (*Alias)(key),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	var err error
	if key.Value, err = hex.DecodeString(aux.Value); err != nil {
		return err
	}

	if key.ValueMask, err = hex.DecodeString(aux.ValueMask); err != nil {
		return err
	}

	return nil
}
