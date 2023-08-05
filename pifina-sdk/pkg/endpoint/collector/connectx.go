package collector

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/cheynewallace/tabby"
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
)

func (c *EndpointCollector) IsNeoSDKExists() bool {
	return c.neohost.IsNeoSDKExists()
}

// List all available Mellanox network interface cards
func (c *EndpointCollector) ListMlxNetworkCards() error {
	result, err := c.neohost.ListMlxNetworkCards()
	if err != nil {
		return err
	}

	c.logger.Debug("Get System device command output", "result", result.Results)

	t := tabby.New()
	t.AddHeader("Device UID", "Type", "Infiniband Devicename", "Interface name")
	for i := range result.Results {
		for _, port := range result.Results[i].Ports {
			if len(port.PhysicalFunctions) > 0 && len(port.PhysicalFunctions[0].NetworkInterfaces) > 0 {
				t.AddLine(port.UID, result.Results[i].Name, port.IbDevice, port.PhysicalFunctions[0].NetworkInterfaces[0])
			} else {
				t.AddLine(port.UID, result.Results[i].Name, port.IbDevice, "")
			}
		}
	}
	t.Print()

	return nil
}

// Starts collection of performance counter from ConnectX card.
// In addition it starts collection of ethtool stats too.
func (c *EndpointCollector) CollectMlxPerfCounters(ctx context.Context, wg *sync.WaitGroup, targetDevices []string) error {
	// Retrieve information about Mellanox interfaces
	result, err := c.neohost.ListMlxNetworkCards()
	if err != nil {
		return err
	}

	// Create a cache userdefined name <-> dev-uid
	devUids := make(map[string]string)
	ethNames := make([]string, 0)

	for _, targetDevice := range targetDevices {
		uid, ok := c.findDevUid(result, targetDevice)
		if !ok {
			c.logger.Error("NIC with the given name has not been found.", "name", targetDevice)
			return &model.ErrNameNotFound{Entity: targetDevice, Msg: "Device not found"}
		}
		devUids[targetDevice] = uid
		ethName, ok := c.findEthNameFromMlxDev(result, targetDevice)
		if !ok {
			c.logger.Error("NIC with the given name has not been found.", "name", targetDevice)
			return &model.ErrNameNotFound{Entity: targetDevice, Msg: "Device not found"}
		}
		ethNames = append(ethNames, ethName)
	}

	for device, uid := range devUids {
		go c.GetMlxPerformanceCounters(ctx, wg, device, uid)
		wg.Add(1)
	}

	// Start Ethtool counter
	c.CollectEthCounter(ctx, wg, ethNames)

	return nil
}

func (c *EndpointCollector) GetMlxPerformanceCounters(ctx context.Context, wg *sync.WaitGroup, targetDevice string, uid string) error {
	defer wg.Done()

	// Initialize ticker
	ticker := time.NewTicker(time.Duration(c.sampleInterval) * time.Second)
	defer ticker.Stop()

	c.logger.Info("Collecting performance counters from NEO-SDK in background", "dev", targetDevice)

	for {
		select {
		case <-ticker.C:
			timeNow := time.Now()
			// Get counters from NEO-Host
			perfCounters, err := c.neohost.GetPerformanceCounters(uid)
			if err != nil {
				c.logger.Warn("Error occured during performance counter collection", "device", targetDevice, "err", err)
				continue
			}
			// Transform metrics to MetricItem object
			metrics := c.transformNeoHostMetrics(perfCounters)
			if c.logger.GetLevel() == hclog.Debug {
				if jsonMetrics, err := json.Marshal(metrics); err != nil {
					c.logger.Debug("Transformed performance counters from NEO Host", "dev", targetDevice, "metrics", jsonMetrics)
				} else {
					c.logger.Debug("Transformed performance counters from NEO Host", "dev", targetDevice, "metrics", metrics)
				}
			}
			// Send metrics
			c.metricSinkChan <- &model.SinkEmitCommand{SourceSuffix: targetDevice, Metrics: metrics}
			c.logger.Debug("Time duration of the collection", "dev", targetDevice, "duration", time.Since(timeNow))
		case <-ctx.Done():
			c.logger.Info("Stopping neohost collector...", "dev", targetDevice)
			return nil
		}
	}
}

// Transform NEO-Host result to a slice of MetricItem
// Only selected
func (c *EndpointCollector) transformNeoHostMetrics(perfCounters *model.NeoHostPerfCounterResult) []*model.MetricItem {
	metrics := make([]*model.MetricItem, 0)
	timeNow := time.Now()
	for i := range perfCounters.Counters {
		counterName := perfCounters.Counters[i].Counter.Name
		if _, ok := c.neoHostCounterNameCache[counterName]; ok {
			metrics = append(metrics, &model.MetricItem{
				MetricName:  counterName,
				Value:       uint64(math.Round(perfCounters.Counters[i].Counter.Value)),
				Type:        model.METRIC_EXT_VALUE,
				SessionId:   0,
				LastUpdated: timeNow,
			})
		}
	}
	for i := range perfCounters.Analysis {
		counterName := perfCounters.Analysis[i].AnalysisAttribute.Name
		if _, ok := c.neoHostCounterNameCache[counterName]; ok {
			metrics = append(metrics, &model.MetricItem{
				MetricName:  counterName,
				Value:       uint64(math.Round(perfCounters.Analysis[i].AnalysisAttribute.Value)),
				Type:        model.METRIC_EXT_VALUE,
				SessionId:   0,
				LastUpdated: timeNow,
			})
		}
	}

	return metrics
}

// Find dev-uid given a search string, which can be dev-uid, ibDevice name or eth name
func (c *EndpointCollector) findDevUid(devices *model.NeoHostDeviceList, targetDevice string) (string, bool) {
	for i := range devices.Results {
		for _, port := range devices.Results[i].Ports {
			if port.UID == targetDevice || port.IbDevice == targetDevice {
				return port.UID, true
			}
			if len(port.PhysicalFunctions) > 0 && len(port.PhysicalFunctions[0].NetworkInterfaces) > 0 {
				if port.PhysicalFunctions[0].NetworkInterfaces[0] == targetDevice {
					return port.UID, true
				}
			}
		}
	}

	return "", false
}

func (c *EndpointCollector) findEthNameFromMlxDev(devices *model.NeoHostDeviceList, targetDevice string) (string, bool) {
	for i := range devices.Results {
		for _, port := range devices.Results[i].Ports {
			if len(port.PhysicalFunctions) > 0 && len(port.PhysicalFunctions[0].NetworkInterfaces) > 0 {
				if port.PhysicalFunctions[0].NetworkInterfaces[0] == targetDevice || port.UID == targetDevice || port.IbDevice == targetDevice {
					return port.PhysicalFunctions[0].NetworkInterfaces[0], true
				}
			}
		}
	}

	return "", false
}
