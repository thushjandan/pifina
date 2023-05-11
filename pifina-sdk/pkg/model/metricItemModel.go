package model

import "time"

type MetricItem struct {
	SessionId   uint32    `json:"sessionId"`
	Type        string    `json:"type"`
	Value       uint64    `json:"value"`
	MetricName  string    `json:"metricName"`
	LastUpdated time.Time `json:"timestamp"`
}

type TelemetryMessage struct {
	Source     string
	MetricList []*MetricItem
}
