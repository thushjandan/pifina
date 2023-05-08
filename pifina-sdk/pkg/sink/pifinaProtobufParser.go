package sink

import (
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/sink/protos/pifina/pifina"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertMetricsToProtobuf(metrics []*model.MetricItem) []*pifina.PifinaMetric {
	// Check precondition
	if metrics == nil || len(metrics) == 0 {
		return nil
	}
	protoResp := make([]*pifina.PifinaMetric, 0, len(metrics))

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
