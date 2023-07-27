package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/cheynewallace/tabby"
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/sink"
	"github.com/thushjandan/pifina/pkg/endpoint/dataplane/neohost"
	"github.com/thushjandan/pifina/pkg/model"
)

type EndpointCollector struct {
	logger                  hclog.Logger
	neohost                 *neohost.NeoHostDriver
	neoHostCounterNameCache map[string]empty
	sink                    *sink.Sink
}

type EndpointCollectorOptions struct {
	Logger            hclog.Logger
	SDKPath           string
	NEOMode           string
	NEOPort           int
	TelemetryEndpoint string
}

type empty struct{}

func NewEndpointCollector(options *EndpointCollectorOptions) *EndpointCollector {
	neohost := neohost.NewNeoHostDriver(&neohost.NeoHostDriverOptions{
		Logger:  options.Logger.Named("neohost"),
		SDKPath: options.SDKPath,
		NEOMode: options.NEOMode,
		NEOPort: options.NEOPort,
	})

	// Create a cache of interested counter names for fast lookup
	counterNameCache := make(map[string]empty)
	for _, counterName := range model.NEOHOST_COUNTERS {
		counterNameCache[counterName] = empty{}
	}
	// Init sink
	sink := sink.NewSink(options.Logger, options.TelemetryEndpoint)
	return &EndpointCollector{
		logger:                  options.Logger.Named("endpoint-collector"),
		neohost:                 neohost,
		neoHostCounterNameCache: counterNameCache,
		sink:                    sink,
	}
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

func (c *EndpointCollector) GetMlxPerformanceCounters(ctx context.Context, targetDevices []string, sampleInterval int) error {
	// Retrieve information about Mellanox interfaces
	result, err := c.neohost.ListMlxNetworkCards()
	if err != nil {
		return err
	}

	// Use hostname as source
	hostname, err := os.Hostname()
	if err != nil {
		c.logger.Error("Cannot retrieve system hostname. setting system name to unknown")
		hostname = "unknown"
	}

	// Create a cache userdefined name <-> dev-uid
	devUids := make(map[string]string)

	for _, targetDevice := range targetDevices {
		uid, ok := c.findDevUid(result, targetDevice)
		if !ok {
			c.logger.Error("NIC with the given dev-uid has not been found.", "dev-uid", targetDevice)
			return &model.ErrNameNotFound{Entity: targetDevice, Msg: "Device not found"}
		}
		devUids[targetDevice] = uid
	}

	// Initialize ticker
	ticker := time.NewTicker(time.Duration(sampleInterval) * time.Second)
	defer ticker.Stop()

	c.logger.Info("Starting to collect performance counters from NEO-SDK...")

	for {
		select {
		case <-ticker.C:
			for targetDevice, uid := range devUids {
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
						c.logger.Debug("Transformed performance counters from NEO Host", "metrics", jsonMetrics)
					} else {
						c.logger.Debug("Transformed performance counters from NEO Host", "metrics", metrics)
					}
				}
				// Send metrics
				err = c.sink.ChunkAndEmitWithSource(metrics, fmt.Sprintf("%s_%s", hostname, targetDevice))
				if err != nil {
					c.logger.Warn("Sending metrics to telemetry server has failed", "err", err)
				}
				c.logger.Debug("Time duration of the collection", "duration", time.Since(timeNow))
			}
		case <-ctx.Done():
			c.logger.Info("Stopping collector...")
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
