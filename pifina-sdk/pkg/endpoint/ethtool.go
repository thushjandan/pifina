package endpoint

import (
	"context"
	"fmt"
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

func (c *EndpointCollector) GetEthtoolStats(ctx context.Context, wg *sync.WaitGroup, ethtoolHandle *ethtool.Ethtool, deviceName string) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(c.sampleInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := ethtoolHandle.Stats(deviceName)
			if err != nil {
				c.logger.Warn("Cannot retrieve ethtool stats from NIC", "dev", deviceName, "err", err)
				continue
			}
			c.transformEthtoolMetrics(stats)
		case <-ctx.Done():
			c.logger.Info("Stopping ethtool collector", "dev", deviceName)
			return
		}
	}
}

func (c *EndpointCollector) transformEthtoolMetrics(ethtoolStats map[string]uint64) []*model.MetricItem {
	for statName, statValue := range ethtoolStats {
		fmt.Printf("%s => %d", statName, statValue)
	}
	return nil
}

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
