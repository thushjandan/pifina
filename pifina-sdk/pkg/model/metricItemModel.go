// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
	Source     string        `json:"source"`
	HostType   string        `json:"type"`
	GroupId    uint32        `json:"groupId"`
	MetricList []*MetricItem `json:"metrics"`
}

const (
	METRIC_BYTES     = "METRIC_BYTES"
	METRIC_PKTS      = "METRIC_PKTS"
	METRIC_EXT_VALUE = "METRIC_EXT_VALUE"
	HOSTTYPE_TOFINO  = "HOSTTYPE_TOFINO"
	HOSTTYPE_NIC     = "HOSTTYPE_NIC"
)
