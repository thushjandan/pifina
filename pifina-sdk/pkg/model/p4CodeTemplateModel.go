package model

type P4CodeTemplate struct {
	SessionIdWidth uint
	MatchKeys      []*P4CodeTemplateKey
}

type P4CodeTemplateKey struct {
	Name      string
	MatchType string
}
