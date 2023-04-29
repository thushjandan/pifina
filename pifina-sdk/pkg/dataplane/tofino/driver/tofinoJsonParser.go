package driver

import "encoding/json"

type ForwardPipelineConfig struct {
	SchemaVersion string  `json:"schema_version"`
	Tables        []Table `json:"tables"`
}

type Table struct {
	Name                  string       `json:"name"`
	Id                    uint32       `json:"id"`
	TableType             string       `json:"table_type"`
	Size                  uint32       `json:"size"`
	Annotations           []Annotation `json:"annotations"`
	DependsOn             []uint32     `json:"depends_on"`
	HasConstDefaultAction bool         `json:"has_const_default_action"`
	Key                   []Field      `json:"key"`
	ActionSpecs           []ActionSpec `json:"action_specs"`
	Data                  []Field      `json:"data"`
	Attributes            []string     `json:"attributes"`
	SupportedOperations   []string     `json:"supported_operations"`
}

type Field struct {
	Id          uint32         `json:"id"`
	Name        string         `json:"name"`
	Repeated    bool           `json:"repeated"`
	Annotations []Annotation   `json:"annotations"`
	Mandatory   bool           `json:"mandatory"`
	MatchType   string         `json:"match_type"`
	ReadOnly    bool           `json:"read_only"`
	Singleton   SingletonField `json:"singleton"`
}

type SingletonField struct {
	Id          uint32       `json:"id"`
	Name        string       `json:"name"`
	Repeated    bool         `json:"repeated"`
	Annotations []Annotation `json:"annotations"`
	Type        Type         `json:"type"`
}

type Type struct {
	Type    string   `json:"type"`
	Width   uint32   `json:"width"`
	Choices []string `json:"choices"`
}

type ActionSpec struct {
	Id          uint32       `json:"id"`
	Name        string       `json:"name"`
	ActionScope string       `json:"action_scope"`
	Annotations []Annotation `json:"annotations"`
	Data        []Field      `json:"data"`
}

type Annotation struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func UnmarshalBfruntimeInfoJson(input []byte) ([]Table, error) {
	var bfrtInfo ForwardPipelineConfig

	err := json.Unmarshal(input, &bfrtInfo)

	if err != nil {
		return nil, err
	}

	return bfrtInfo.Tables, nil
}
