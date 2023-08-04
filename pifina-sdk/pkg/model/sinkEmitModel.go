package model

type SinkEmitCommand struct {
	SourceSuffix string
	Metrics      []*MetricItem
}
