package collector

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/safchain/ethtool"
	"github.com/thushjandan/pifina/pkg/model"
)

func (c *EndpointCollector) CollectEthCounter(ctx context.Context, wg *sync.WaitGroup, targetDevices []string) error {
	ethtoolHandle, err := ethtool.NewEthtool()
	if err != nil {
		return err
	}

	for i := range targetDevices {
		go c.GetEthtoolStats(ctx, wg, ethtoolHandle, targetDevices[i])
		wg.Add(1)
	}

	return nil
}

// Get stats from ethtool
func (c *EndpointCollector) GetEthtoolStats(ctx context.Context, wg *sync.WaitGroup, ethtoolHandle *ethtool.Ethtool, deviceName string) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(c.sampleInterval) * time.Second)
	defer ticker.Stop()

	c.logger.Info("Collecting stats from ethtool background", "dev", deviceName)

	for {
		select {
		case <-ticker.C:
			stats, err := ethtoolHandle.Stats(deviceName)
			if err != nil {
				c.logger.Warn("Cannot retrieve ethtool stats from NIC", "dev", deviceName, "err", err)
				continue
			}
			metrics := c.transformEthtoolMetrics(stats)
			c.metricSinkChan <- &model.SinkEmitCommand{SourceSuffix: deviceName, Metrics: metrics}
		case <-ctx.Done():
			c.logger.Info("Stopping ethtool collector", "dev", deviceName)
			return
		}
	}
}

// Transform ethtool stats to MetricItem objects
func (c *EndpointCollector) transformEthtoolMetrics(ethtoolStats map[string]uint64) []*model.MetricItem {
	timeNow := time.Now()
	metrics := make([]*model.MetricItem, 0)
	for i := range model.ETHTOOL_COUNTERS {
		if statVal, ok := ethtoolStats[model.ETHTOOL_COUNTERS[i]]; ok {
			metrics = append(metrics, &model.MetricItem{
				MetricName:  model.ETHTOOL_COUNTERS[i],
				Value:       statVal,
				LastUpdated: timeNow,
				Type:        model.METRIC_EXT_VALUE,
				SessionId:   0,
			})
		}
	}
	return metrics
}

// Check if given name exists as ethernet interface
func (c *EndpointCollector) IsEthInterfaceExists(deviceName string) (bool, error) {
	allInterfaces, err := net.Interfaces()
	if err != nil {
		return false, err
	}

	for i := range allInterfaces {
		if allInterfaces[i].Name == deviceName {
			return true, nil
		}
	}

	return false, nil
}
