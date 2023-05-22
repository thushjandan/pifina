package collector

import (
	"context"
	"sync"
	"time"

	"github.com/thushjandan/pifina/pkg/model"
)

func (c *MetricCollector) CollectAppRegisterValues(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	// Stop ticker before leaving
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			appRegistersToReq := c.ts.GetAppRegisterProbes()
			if len(appRegistersToReq) > 0 {
				metrics, err := c.driver.GetMetricFromRegister(appRegistersToReq, model.METRIC_EXT_VALUE)
				if err != nil {
					c.logger.Error("Error occured during collection of application owned registers", "err", err)
				} else {
					c.logger.Trace("Collection of application owned registers has succeeded.")
					for i := range metrics {
						metricSink <- metrics[i]
					}
				}
			}
		case <-ctx.Done():
			c.logger.Info("Stopping application owned register collector...")
			return
		}
	}
}
