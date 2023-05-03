package sink

import (
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/sink/protos/pifina/pifina"
)

func ConvertMetricsToProtobuf(metrics []*driver.MetricItem) []*pifina.PifinaMetric {
	// Check precondition
	if metrics == nil || len(metrics) == 0 {
		return nil
	}
	protoResp := make([]*pifina.PifinaMetric, 0, len(metrics))

	for i := range metrics {
		protoResp = append(protoResp, &pifina.PifinaMetric{
			SessionId:  metrics[i].SessionId,
			Value:      metrics[i].Value,
			ValueType:  metrics[i].Type,
			MetricName: metrics[i].MetricName,
		})
	}

	return protoResp
}
