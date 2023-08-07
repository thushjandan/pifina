package collector

import (
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/console/nic/dataplane/neohost"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/sink"
)

type EndpointCollector struct {
	logger                  hclog.Logger
	sampleInterval          int
	metricSinkChan          chan *model.SinkEmitCommand
	neohost                 *neohost.NeoHostDriver
	neoHostCounterNameCache map[string]empty
	ethNameCache            map[string]string
	sink                    *sink.Sink
}

type EndpointCollectorOptions struct {
	Logger            hclog.Logger
	SampleInterval    int
	MetricSinkChan    chan *model.SinkEmitCommand
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
	return &EndpointCollector{
		logger:                  options.Logger.Named("endpoint-collector"),
		sampleInterval:          options.SampleInterval,
		neohost:                 neohost,
		neoHostCounterNameCache: counterNameCache,
		metricSinkChan:          options.MetricSinkChan,
		ethNameCache:            make(map[string]string), // EthName <-> user defined Name
	}
}
