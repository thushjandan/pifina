package model

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
