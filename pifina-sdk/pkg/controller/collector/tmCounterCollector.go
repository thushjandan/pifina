package collector

import (
	"context"
	"sync"
	"time"

	"github.com/thushjandan/pifina/pkg/model"
)

func (c *MetricCollector) CollectTrafficManagerCounters(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	// Stop ticker before leaving
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			monitoredPorts := c.ts.GetMonitoredPorts()
			if len(monitoredPorts) > 0 {
				metrics, err := c.driver.GetTMCountersByPort(monitoredPorts)
				if err != nil {
					c.logger.Error("Error occured during collection of traffic manager port based counters", "err", err)
				} else {
					c.logger.Trace("Collection of traffic manager counters has succeeded.")
					for i := range metrics {
						metricSink <- metrics[i]
					}
				}
			}
			metrics, err := c.driver.GetTMPipelineCounter(c.pipelineCount)
			if err != nil {
				c.logger.Error("Error occured during collection of traffic manager pipeline based counters", "err", err)
			} else {
				c.logger.Trace("Collection of traffic manager counters has succeeded.")
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
		case <-ctx.Done():
			c.logger.Info("Stopping traffic manager counter collector...")
			return
		}
	}

}
