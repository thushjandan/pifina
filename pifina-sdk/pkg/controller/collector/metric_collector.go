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
)

type MetricCollector struct {
	logger         hclog.Logger
	driver         *driver.TofinoDriver
	sampleInterval time.Duration
	ts             *trafficselector.TrafficSelector
	lpfTimeConst   float32
}

func NewMetricCollector(logger hclog.Logger, driver *driver.TofinoDriver, sampleInterval int, ts *trafficselector.TrafficSelector) *MetricCollector {
	return &MetricCollector{
		logger:         logger.Named("collector"),
		driver:         driver,
		sampleInterval: time.Duration(sampleInterval) * time.Millisecond,
		ts:             ts,
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
	/*go collector.CollectIngressStartMatchCounter(ctx, wg, metricSink)
	go collector.CollectIngressHdrStartCounter(ctx, wg, metricSink)
	go collector.CollectIngressHdrEndCounter(ctx, wg, metricSink)
	go collector.CollectEgressStartCounter(ctx, wg, metricSink)
	go collector.CollectEgressEndCounter(ctx, wg, metricSink)
	go collector.CollectIngressJitter(ctx, wg, metricSink)
	go collector.CollectAppRegisterValues(ctx, wg, metricSink)
	go collector.CollectTrafficManagerCounters(ctx, wg, metricSink)

	// Start collector for each extra probe
	extraProbes := collector.driver.GetExtraProbes()
	for i := range extraProbes {
		wg.Add(1)
		go collector.CollectExtraHeaderSizeCounter(ctx, wg, metricSink, extraProbes[i])
	}*/
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
			collector.logger.Debug("Time collection after sending read", "time", time.Since(start))
			if err != nil {
				collector.logger.Error("Error occured during collection", "err", err)
			}
			// Reset counters
			collector.ResetCounters(sessionIds)
			// Process metrics
			metrics, err := collector.driver.ProcessMetricResponse(bfResponse)
			if err != nil {
				collector.logger.Error("Error occured during processing raw metric values", "err", err)
			} else {
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
			collector.logger.Debug("Time Collection end", "time", time.Since(start))
		// Terminate the for loop.
		case <-ctx.Done():
			collector.logger.Info("Stopping collector...")
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
		collector.logger.Error("Resetting counters failed!", "err", err)
	}
}
