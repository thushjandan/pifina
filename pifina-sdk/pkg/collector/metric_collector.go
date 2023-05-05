package collector

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
)

type MetricCollector struct {
	logger         hclog.Logger
	driver         *driver.TofinoDriver
	sessionIdCache []uint32
}

func NewMetricCollector(logger hclog.Logger, driver *driver.TofinoDriver) *MetricCollector {
	return &MetricCollector{
		logger: logger.Named("collector"),
		driver: driver,
	}
}

// Retrieve the match selector entries and extract the session IDs.
func (collector *MetricCollector) LoadSessionsFromDevice() error {
	sessions, err := collector.driver.GetSessionsFromMatchSelectors()
	if err != nil {
		return err
	}
	collector.sessionIdCache = sessions
	return nil

}

func (collector *MetricCollector) GetSessionIdCache() []uint32 {
	return collector.sessionIdCache
}

func (collector *MetricCollector) StartMetricCollection(ctx context.Context, wg *sync.WaitGroup, metricSink chan driver.MetricItem) {
	// If sessionId cache is empty, then refresh the cache
	if collector.sessionIdCache == nil {
		err := collector.LoadSessionsFromDevice()
		if err != nil {
			collector.logger.Error("Error occured during collection. Cannot retrieve sessionIds from Ingress Start Match table", "err", err)
			return
		}
	}

	wg.Add(1)
	go collector.CollectIngressStartMatchCounter(ctx, wg)

	wg.Add(1)
	go collector.CollectIngressHdrStartCounter(ctx, wg)

	wg.Add(1)
	go collector.CollectIngressHdrEndCounter(ctx, wg)

	wg.Add(1)
	go collector.CollectEgressStartCounter(ctx, wg)

	wg.Add(1)
	go collector.CollectEgressEndCounter(ctx, wg)
}

func (collector *MetricCollector) CollectIngressStartMatchCounter(ctx context.Context, wg *sync.WaitGroup) {
	// Mark the context as done after exiting the routine.
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	// Stop the ticker before leaving
	defer ticker.Stop()

	for {
		select {
		// Got a tick from the ticker.
		case <-ticker.C:
			_, err := collector.driver.GetIngressStartMatchSelectorCounter()
			if err != nil {
				collector.logger.Error("Error occured during collection of Ingress Start Match table counter", "err", err)
			} else {
				collector.logger.Debug("Collection of Ingress Start Match table counter has succeeded.")
			}
		// Terminate the for loop.
		case <-ctx.Done():
			collector.logger.Info("Stopping Ingress Start Match table counter collector...")
			return
		}
	}
}

func (collector *MetricCollector) CollectIngressHdrStartCounter(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, err := collector.driver.GetIngressHdrStartCounter(collector.sessionIdCache)
			if err != nil {
				collector.logger.Error("Error occured during collection of Ingress header start size counter", "err", err)
			} else {
				collector.logger.Debug("Collection of Ingress header start size counter has succeeded.")
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Ingress header start size counter collector...")
			return
		}
	}
}

func (collector *MetricCollector) CollectIngressHdrEndCounter(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, err := collector.driver.GetIngressHdrEndCounter(collector.sessionIdCache)
			if err != nil {
				collector.logger.Error("Error occured during collection of Ingress header end size counter", "err", err)
			} else {
				collector.logger.Debug("Collection of Ingress header end size counter has succeeded.")
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Ingress header end size counter collector...")
			return
		}
	}
}

func (collector *MetricCollector) CollectEgressStartCounter(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := collector.driver.GetEgressStartCounter(collector.sessionIdCache)
			if err != nil {
				collector.logger.Error("Error occured during collection of Egress Start counter", "err", err)
			} else {
				collector.logger.Debug("Collection of Egress start counter has succeeded.")
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Egress start counter collector...")
			return
		}
	}

}

func (collector *MetricCollector) CollectEgressEndCounter(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := collector.driver.GetEgressEndCounter(collector.sessionIdCache)
			if err != nil {
				collector.logger.Error("Error occured during collection of Egress End counter", "err", err)
			} else {
				collector.logger.Debug("Collection of Egress end counter has succeeded.")
			}
		case <-ctx.Done():
			collector.logger.Info("Stopping Egress end counter collector...")
			return
		}
	}

}