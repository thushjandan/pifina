package model

type NeoHostPerfCounterResult struct {
	Counters         []NeoHostPerfCounterItem  `json:"counters"`
	Analysis         []NeoHostPerfAnalysisItem `json:"analysis"`
	Metadata         NeoHostPerfGroups         `json:"metadata"`
	AnalysisMetadata NeoHostPerfAnalysisGroups `json:"analysisMetadata"`
}

type NeoHostPerfCounterItem struct {
	Counter NeoHostPerfCounterValue `json:"counter"`
}

type NeoHostPerfCounterValue struct {
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	Value                 float64 `json:"value"`
	Timestamp             int64   `json:"timestamp"`
	UtilizationPercentage int     `json:"UtilizationPercentage"`
	UtilizationReference  string  `json:"UtilizationReference"`
	Units                 string  `json:"units"`
}

type NeoHostPerfAnalysisItem struct {
	AnalysisAttribute NeoHostPerfAnalysisValue `json:"analysisAttribute"`
}

type NeoHostPerfAnalysisValue struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Units       string  `json:"units"`
	Timestamp   int64   `json:"timestamp"`
	Value       float64 `json:"value"`
	ValueType   string  `json:"valueType"`
}

type NeoHostPerfGroups struct {
	Groups []NeoHostPerfMetadataGroupItem `json:"groups"`
}

type NeoHostPerfMetadataGroupItem struct {
	MetadataGroup NeoHostPerfMetadataGroupValue `json:"metadataGroup"`
}
type NeoHostPerfMetadataGroupValue struct {
	Unit     string   `json:"unit"`
	Counters []string `json:"counters"`
}

type NeoHostPerfAnalysisGroups struct {
	Groups []NeoHostPerfAnalysisMetadataGroupItem `json:"groups"`
}

type NeoHostPerfAnalysisMetadataGroupItem struct {
	AnalysisMetadataGroup NeoHostPerfAnalysisMetadataGroupValue `json:"analysisMetadataGroup"`
}

type NeoHostPerfAnalysisMetadataGroupValue struct {
	Group              string   `json:"group"`
	AnalysisAttributes []string `json:"analysisAttributes"`
}
