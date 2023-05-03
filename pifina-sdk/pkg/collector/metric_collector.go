package collector

import (
	"fmt"

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

func (collector *MetricCollector) TriggerMetricCollection() {
	// If sessionId cache is empty, then refresh the cache
	if collector.sessionIdCache == nil {
		err := collector.LoadSessionsFromDevice()
		if err != nil {
			collector.logger.Error("Error occured during collection. Cannot retrieve sessionIds from Ingress Start Match table", "err", err)
			return
		}
	}

	err := collector.CollectIngressStartMatchCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Ingress Start Match table counter", "err", err)
		return
	}
	err = collector.CollectIngressHdrStartCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Ingress header start size counter", "err", err)
		return
	}
	err = collector.CollectIngressHdrEndCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Ingress header end size counter", "err", err)
		return
	}
	err = collector.CollectEgressStartCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Egress Start counter", "err", err)
		return
	}
	err = collector.CollectEgressEndCounter()
	if err != nil {
		collector.logger.Error("Error occured during collection of Egress End counter", "err", err)
		return
	}

}

func (collector *MetricCollector) CollectIngressStartMatchCounter() error {
	metrics, err := collector.driver.GetIngressStartMatchSelectorCounter()
	if err != nil {
		return err
	}
	for _, item := range metrics {
		fmt.Printf("%+v\n", *item)
	}
	return nil
}

func (collector *MetricCollector) CollectIngressHdrStartCounter() error {
	metrics, err := collector.driver.GetIngressHdrStartCounter(collector.sessionIdCache)
	if err != nil {
		return err
	}
	for _, item := range metrics {
		fmt.Printf("%+v\n", *item)
	}
	return nil
}

func (collector *MetricCollector) CollectIngressHdrEndCounter() error {
	metrics, err := collector.driver.GetIngressHdrEndCounter(collector.sessionIdCache)
	if err != nil {
		return err
	}
	for _, item := range metrics {
		fmt.Printf("%+v\n", *item)
	}
	return nil
}

func (collector *MetricCollector) CollectEgressStartCounter() error {
	metrics, err := collector.driver.GetEgressStartCounter(collector.sessionIdCache)
	if err != nil {
		return err
	}
	for _, item := range metrics {
		fmt.Printf("%+v\n", *item)
	}

	return nil
}

func (collector *MetricCollector) CollectEgressEndCounter() error {
	metrics, err := collector.driver.GetEgressEndCounter(collector.sessionIdCache)
	if err != nil {
		return err
	}
	for _, item := range metrics {
		fmt.Printf("%+v\n", *item)
	}
	return nil
}
