package model

type MatchSelectorSchema struct {
	FieldId   uint32
	Name      string
	MatchType string
	Type      string
	Width     uint32
}

type MatchSelectorEntry struct {
	SessionId uint32
	Keys      []*MatchSelectorKey
}

type MatchSelectorKey struct {
	FieldId   uint32
	Value     []byte
	MatchType string
	ValueMask []byte
	ValueLpm  int32
}
