package model

import "time"

type MetricItem struct {
	SessionId   uint32
	Type        string
	Value       uint64
	MetricName  string
	LastUpdated time.Time
}
