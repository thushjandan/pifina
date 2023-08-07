package collector

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
	"github.com/thushjandan/pifina/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricCollector struct {
	logger         hclog.Logger
	driver         *driver.TofinoDriver
	sampleInterval time.Duration
	ts             *trafficselector.TrafficSelector
	lpfTimeConst   float32
	pipelineCount  int
}

func NewMetricCollector(logger hclog.Logger, driver *driver.TofinoDriver, sampleInterval int, ts *trafficselector.TrafficSelector, pipelineCount int) *MetricCollector {
	return &MetricCollector{
		logger:         logger.Named("collector"),
		driver:         driver,
		sampleInterval: time.Duration(sampleInterval) * time.Millisecond,
		ts:             ts,
		pipelineCount:  pipelineCount,
	}
}

func (collector *MetricCollector) StartMetricCollection(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	// If sessionId cache is empty, then refresh the cache
	if collector.ts.GetTrafficSelectorCache() == nil {
		collector.logger.Error("Cannot start collection! Cannot retrieve sessionIds from Ingress Start Match table. Exiting.")
		return
	}

	err := collector.ts.ConfigureLPF()
	if err != nil {
		collector.logger.Error("Error occured during LPF initialization", "err", err)
	}

	wg.Add(1)
	// Start collector threads
	go collector.CollectMetrics(ctx, wg, metricSink)
}

func (collector *MetricCollector) CollectMetrics(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	// Mark the context as done after exiting the routine.
	defer wg.Done()

	ticker := time.NewTicker(collector.sampleInterval)
	// Stop the ticker before leaving
	defer ticker.Stop()

	for {
		select {
		// Got a tick from the ticker.
		case <-ticker.C:
			start := time.Now()
			sessionIds := collector.ts.GetSessionIdCache()
			allMetricRequests := make([]*bfruntime.Entity, 0)
			metricRequests, err := collector.driver.GetMatchSelectorEntriesRequest()
			if err == nil {
				allMetricRequests = append(allMetricRequests, metricRequests...)
			}
			metricRequests, err = collector.driver.GetIngressHdrStartCounter(sessionIds)
			if err == nil {
				allMetricRequests = append(allMetricRequests, metricRequests...)
			}
			metricRequests, err = collector.driver.GetIngressHdrEndCounter(sessionIds)
			if err == nil {
				allMetricRequests = append(allMetricRequests, metricRequests...)
			}
			metricRequests, err = collector.driver.GetEgressStartCounter(sessionIds)
			if err == nil {
				allMetricRequests = append(allMetricRequests, metricRequests...)
			}
			metricRequests, err = collector.driver.GetEgressEndCounter(sessionIds)
			if err == nil {
				allMetricRequests = append(allMetricRequests, metricRequests...)
			}
			metricRequests, err = collector.driver.GetIngressJitter(sessionIds)
			if err == nil {
				allMetricRequests = append(allMetricRequests, metricRequests...)
			}
			// App registers
			appRegistersToReq := collector.ts.GetAppRegisterProbes()
			if len(appRegistersToReq) > 0 {
				metricRequests, err = collector.driver.GetMetricFromRegisterRequest(appRegistersToReq, model.METRIC_EXT_VALUE)
				if err == nil {
					allMetricRequests = append(allMetricRequests, metricRequests...)
				}
			}
			// Extra Probes
			extraProbes := collector.driver.GetExtraProbes()
			for i := range extraProbes {
				metricRequests, err = collector.driver.GetHdrSizeCounter(extraProbes[i], sessionIds)
				if err == nil {
					allMetricRequests = append(allMetricRequests, metricRequests...)
				}
			}
			bfResponse, err := collector.driver.SendReadRequest(allMetricRequests)
			if err != nil {
				// Check if GRPC request has been canceled
				// If true, then user stopped app. Skip processing and move to cleanup
				if collector.errorIsCanceled(err) {
					continue
				}
				collector.logger.Error("Error occured during collection", "err", err)
			}
			// Reset counters
			collector.ResetCounters(sessionIds)
			// Traffic manager requests per port
			monitoredPorts := collector.ts.GetMonitoredPorts()
			if len(monitoredPorts) > 0 {
				metricRequests = collector.driver.GetTMCountersByPortRequests(monitoredPorts)
				tmBfResponse, err := collector.driver.SendReadRequest(metricRequests)
				if err != nil {
					// Check if GRPC request has been canceled
					// If true, then user stopped app. Skip processing and move to cleanup
					if collector.errorIsCanceled(err) {
						continue
					}
					collector.logger.Warn("Error occured during collection of traffic manager metric", "ports", monitoredPorts, "err", err)
				}
				bfResponse = append(bfResponse, tmBfResponse...)
			}
			// Traffic manager requests per pipeline
			tmMetrics, err := collector.driver.GetTMPipelineCounter(collector.pipelineCount)
			if err != nil {
				// Check if GRPC request has been canceled
				// If true, then user stopped app. Skip processing and move to cleanup
				if collector.errorIsCanceled(err) {
					continue
				}
				collector.logger.Warn("Error occured during collection of traffic manager metrics per pipeline", "err", err)
			}
			// Process metrics
			metrics, err := collector.driver.ProcessMetricResponse(bfResponse)
			if err != nil {
				collector.logger.Error("Error occured during processing raw metric values", "err", err)
			} else {
				// Append TM metrics
				metrics = append(metrics, tmMetrics...)
				// Send to sink thread
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
			collector.logger.Debug("Time Collection end", "time", time.Since(start))
		// Terminate the for loop.
		case <-ctx.Done():
			collector.logger.Info("Stopping collector...")
			// Sender closes the channel
			close(metricSink)
			return
		}
	}

}

func (collector *MetricCollector) ResetCounters(sessionIds []uint32) {
	selectorEntries := collector.ts.GetTrafficSelectorCache()
	// Reset register values
	allResetRequests, err := collector.driver.GetResetTableSelectorRequests(selectorEntries)
	if err != nil {
		collector.logger.Warn("Cannot retrieve reset requests for match action table", "err", err)
		allResetRequests = make([]*bfruntime.Update, 0)
	}
	resetRequests := collector.driver.GetResetRegisterRequest(sessionIds)
	allResetRequests = append(allResetRequests, resetRequests...)
	// Counter Reset requests
	resetRequests = collector.driver.GetResetCounterRequests(sessionIds)
	allResetRequests = append(allResetRequests, resetRequests...)
	err = collector.driver.SendWriteRequest(allResetRequests)
	if err != nil {
		// Check if grpc request has been canceled
		// If true, then user stopped app. Skip processing and move to cleanup
		if collector.errorIsCanceled(err) {
			return
		}
		collector.logger.Error("Resetting counters failed!", "err", err)
	}
}

// Checks if given error is context canceled error
// Most probably initiated by the user.
func (collector *MetricCollector) errorIsCanceled(err error) bool {
	if e, ok := status.FromError(err); ok && e.Code() == codes.Canceled {
		return true
	}
	return false
}
