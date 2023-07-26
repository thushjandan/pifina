package endpoint

import (
	"github.com/cheynewallace/tabby"
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/endpoint/dataplane/neohost"
)

type EndpointCollector struct {
	logger  hclog.Logger
	neohost *neohost.NeoHostDriver
}

type EndpointCollectorOptions struct {
	Logger  hclog.Logger
	SDKPath string
	NEOMode string
	NEOPort int
}

func NewEndpointCollector(options *EndpointCollectorOptions) *EndpointCollector {
	neohost := neohost.NewNeoHostDriver(&neohost.NeoHostDriverOptions{
		Logger:  options.Logger,
		SDKPath: options.SDKPath,
		NEOMode: options.NEOMode,
		NEOPort: options.NEOPort,
	})
	return &EndpointCollector{
		logger:  options.Logger,
		neohost: neohost,
	}
}

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
			t.AddLine(port.UID, result.Results[i].Name, port.IbDevice, port.PhysicalFunctions[0].NetworkInterfaces[0])
		}
	}
	t.Print()

	perfCounters, err := c.neohost.GetPerformanceCounters("0000:b3:00.0")
	if err != nil {
		return err
	}

	c.logger.Debug("Raw performance counters from NEO Host", "result", perfCounters.Counters)

	return nil
}
