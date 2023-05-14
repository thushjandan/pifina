package collector

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
	"github.com/thushjandan/pifina/pkg/model"
)

type MetricCollector struct {
	logger         hclog.Logger
	driver         *driver.TofinoDriver
	sampleInterval time.Duration
	ts             *trafficselector.TrafficSelector
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

	wg.Add(1)
	go collector.CollectIngressStartMatchCounter(ctx, wg, metricSink)

	wg.Add(1)
	go collector.CollectIngressHdrStartCounter(ctx, wg, metricSink)

	wg.Add(1)
	go collector.CollectIngressHdrEndCounter(ctx, wg, metricSink)

	wg.Add(1)
	go collector.CollectEgressStartCounter(ctx, wg, metricSink)

	wg.Add(1)
	go collector.CollectEgressEndCounter(ctx, wg, metricSink)
}

func (collector *MetricCollector) CollectIngressStartMatchCounter(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	// Mark the context as done after exiting the routine.
	defer wg.Done()

	ticker := time.NewTicker(collector.sampleInterval)
	// Stop the ticker before leaving
	defer ticker.Stop()

	for {
		select {
		// Got a tick from the ticker.
		case <-ticker.C:
			metrics, err := collector.driver.GetIngressStartMatchSelectorCounter()
			if err != nil {
				collector.logger.Error("Error occured during collection of Ingress Start Match table counter", "err", err)
			} else {
				collector.logger.Trace("Collection of Ingress Start Match table counter has succeeded.")
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
		// Terminate the for loop.
		case <-ctx.Done():
			collector.logger.Info("Stopping Ingress Start Match table counter collector...")
			return
		}
	}
}

func (collector *MetricCollector) CollectIngressHdrStartCounter(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(collector.sampleInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sessionIds := collector.ts.GetSessionIdCache()
			metrics, err := collector.driver.GetIngressHdrStartCounter(sessionIds)
			if err != nil {
				collector.logger.Error("Error occured during collection of Ingress header start size counter", "err", err)
			} else {
				collector.logger.Trace("Collection of Ingress header start size counter has succeeded.")
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Ingress header start size counter collector...")
			return
		}
	}
}

func (collector *MetricCollector) CollectIngressHdrEndCounter(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(collector.sampleInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sessionIds := collector.ts.GetSessionIdCache()
			metrics, err := collector.driver.GetIngressHdrEndCounter(sessionIds)
			if err != nil {
				collector.logger.Error("Error occured during collection of Ingress header end size counter", "err", err)
			} else {
				collector.logger.Trace("Collection of Ingress header end size counter has succeeded.")
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Ingress header end size counter collector...")
			return
		}
	}
}

func (collector *MetricCollector) CollectEgressStartCounter(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(collector.sampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sessionIds := collector.ts.GetSessionIdCache()
			metrics, err := collector.driver.GetEgressStartCounter(sessionIds)
			if err != nil {
				collector.logger.Error("Error occured during collection of Egress Start counter", "err", err)
			} else {
				collector.logger.Trace("Collection of Egress start counter has succeeded.")
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Egress start counter collector...")
			return
		}
	}

}

func (collector *MetricCollector) CollectEgressEndCounter(ctx context.Context, wg *sync.WaitGroup, metricSink chan *model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(collector.sampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sessionIds := collector.ts.GetSessionIdCache()
			metrics, err := collector.driver.GetEgressEndCounter(sessionIds)
			if err != nil {
				collector.logger.Error("Error occured during collection of Egress End counter", "err", err)
			} else {
				collector.logger.Trace("Collection of Egress end counter has succeeded.")
				for i := range metrics {
					metricSink <- metrics[i]
				}
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Egress end counter collector...")
			return
		}
	}

}
