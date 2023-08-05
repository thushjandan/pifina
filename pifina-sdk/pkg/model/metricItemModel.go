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
	HostType   string
	GroupId    uint32
	MetricList []*MetricItem
}

const (
	METRIC_BYTES     = "METRIC_BYTES"
	METRIC_PKTS      = "METRIC_PKTS"
	METRIC_EXT_VALUE = "METRIC_EXT_VALUE"
	HOSTTYPE_TOFINO  = "HOSTTYPE_TOFINO"
	HOSTTYPE_NIC     = "HOSTTYPE_NIC"
)
