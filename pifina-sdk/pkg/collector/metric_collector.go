package collector

import (
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

func (collector *MetricCollector) TriggerMetricCollection() []*driver.MetricItem {
	// If sessionId cache is empty, then refresh the cache
	if collector.sessionIdCache == nil {
		err := collector.LoadSessionsFromDevice()
		if err != nil {
			collector.logger.Error("Error occured during collection. Cannot retrieve sessionIds from Ingress Start Match table", "err", err)
			return nil
		}
	}

	metrics, err := collector.CollectIngressStartMatchCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Ingress Start Match table counter", "err", err)
		metrics = make([]*driver.MetricItem, 0)
	}

	tmpMetrics, err := collector.CollectIngressHdrStartCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Ingress header start size counter", "err", err)
	} else {
		metrics = append(metrics, tmpMetrics...)
	}

	tmpMetrics, err = collector.CollectIngressHdrEndCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Ingress header end size counter", "err", err)
	} else {
		metrics = append(metrics, tmpMetrics...)
	}

	tmpMetrics, err = collector.CollectEgressStartCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Egress Start counter", "err", err)
	} else {
		metrics = append(metrics, tmpMetrics...)
	}

	tmpMetrics, err = collector.CollectEgressEndCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Egress End counter", "err", err)
	} else {
		metrics = append(metrics, tmpMetrics...)
	}

	return metrics
}

func (collector *MetricCollector) CollectIngressStartMatchCounter() ([]*driver.MetricItem, error) {
	return collector.driver.GetIngressStartMatchSelectorCounter()
}

func (collector *MetricCollector) CollectIngressHdrStartCounter() ([]*driver.MetricItem, error) {
	return collector.driver.GetIngressHdrStartCounter(collector.sessionIdCache)
}

func (collector *MetricCollector) CollectIngressHdrEndCounter() ([]*driver.MetricItem, error) {
	return collector.driver.GetIngressHdrEndCounter(collector.sessionIdCache)
}

func (collector *MetricCollector) CollectEgressStartCounter() ([]*driver.MetricItem, error) {
	return collector.driver.GetEgressStartCounter(collector.sessionIdCache)
}

func (collector *MetricCollector) CollectEgressEndCounter() ([]*driver.MetricItem, error) {
	return collector.driver.GetEgressEndCounter(collector.sessionIdCache)
}
