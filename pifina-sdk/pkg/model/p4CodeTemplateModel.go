package model

type P4CodeTemplate struct {
	SessionIdWidth    uint
	MatchKeys         []*P4CodeTemplateKey
	IngressHeaderType string
	EgressHeaderType  string
	ExtraProbeList    []ExtraProbeTemplate
}

type P4CodeTemplateKey struct {
	Name      string
	MatchType string
}

type ExtraProbeTemplate struct {
	Name string
	Type string
}

const (
	EXTRA_PROBE_TYPE_IG = "INGRESS"
	EXTRA_PROBE_TYPE_EG = "EGRESS"
)
