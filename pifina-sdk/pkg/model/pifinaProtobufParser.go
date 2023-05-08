package model

import (
	"github.com/thushjandan/pifina/pkg/sink/protos/pifina/pifina"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertMetricsToProtobuf(metrics []*MetricItem) []*pifina.PifinaMetric {
	protoResp := make([]*pifina.PifinaMetric, 0, len(metrics))
	// Check precondition
	if metrics == nil || len(metrics) == 0 {
		return protoResp
	}

	for i := range metrics {
		protoResp = append(protoResp, &pifina.PifinaMetric{
			SessionId:   metrics[i].SessionId,
			Value:       metrics[i].Value,
			ValueType:   metrics[i].Type,
			MetricName:  metrics[i].MetricName,
			LastUpdated: timestamppb.New(metrics[i].LastUpdated),
		})
	}

	return protoResp
}

func ConvertProtobufToMetrics(rawMetrics []*pifina.PifinaMetric) []*MetricItem {
	data := make([]*MetricItem, 0, len(rawMetrics))

	// Check precondition
	if rawMetrics == nil || len(rawMetrics) == 0 {
		return data
	}

	for i := range rawMetrics {
		data = append(data, &MetricItem{
			SessionId:   rawMetrics[i].SessionId,
			Value:       rawMetrics[i].Value,
			Type:        rawMetrics[i].ValueType,
			MetricName:  rawMetrics[i].MetricName,
			LastUpdated: rawMetrics[i].LastUpdated.AsTime(),
		})
	}

	return data
}
