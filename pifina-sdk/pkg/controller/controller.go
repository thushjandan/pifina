// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package controller

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/api"
	"github.com/thushjandan/pifina/pkg/controller/bufferpool"
	"github.com/thushjandan/pifina/pkg/controller/collector"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/sink"
)

type TofinoController struct {
	ctx            context.Context
	logger         hclog.Logger
	endpoint       string
	p4name         string
	connectTimeout int
	driver         *driver.TofinoDriver
	collector      *collector.MetricCollector
	ts             *trafficselector.TrafficSelector
	sink           *sink.Sink
	bp             *bufferpool.Bufferpool
	api            *api.ControllerApiServer
}

type TofinoControllerOptions struct {
	Logger                  hclog.Logger
	Endpoint                string
	GroupId                 uint
	ConnectTimeout          int
	P4name                  string
	CollectorServerEndpoint string
	SampleInterval          int
	APIPort                 string
	LpfTimeConst            float32
	PipelineCount           int
}

func NewTofinoController(options *TofinoControllerOptions) *TofinoController {
	if options.Logger == nil {
		return nil
	}
	driver := driver.NewTofinoDriver(options.Logger, options.P4name)
	ts := trafficselector.NewTrafficSelector(options.Logger, driver, options.LpfTimeConst)
	collector := collector.NewMetricCollector(options.Logger, driver, options.SampleInterval, ts, options.PipelineCount)
	bp := bufferpool.NewBufferpool(options.Logger, driver, ts)
	apiServer := api.NewControllerApiServer(options.Logger, options.APIPort, ts, bp)
	sink := sink.NewSink(options.Logger, model.HOSTTYPE_TOFINO, options.CollectorServerEndpoint, uint32(options.GroupId))
	return &TofinoController{
		logger:         options.Logger.Named("controller"),
		driver:         driver,
		collector:      collector,
		endpoint:       options.Endpoint,
		p4name:         options.P4name,
		connectTimeout: options.ConnectTimeout,
		sink:           sink,
		ts:             ts,
		bp:             bp,
		api:            apiServer,
	}
}

func (controller *TofinoController) StartController(ctx context.Context, wg *sync.WaitGroup) error {
	controller.ctx = ctx
	// Connect to switch
	err := controller.driver.Connect(ctx, controller.endpoint, controller.connectTimeout)
	if err != nil {
		return err
	}
	// Disconnect from switch after terminating the controller
	defer controller.driver.Disconnect()
	err = controller.driver.LoadPortNameCache()
	if err != nil {
		return err
	}

	metricDataChannel := make(chan *model.MetricItem, 10)
	metricsSinkChannel := make(chan *model.SinkEmitCommand)
	wg.Add(2)
	go controller.sink.StartSink(ctx, wg, metricsSinkChannel)
	// Start Bufferpool and Sampler
	go controller.bp.StartBufferpoolManager(ctx, wg, metricDataChannel, metricsSinkChannel)
	// Start collector threads
	controller.collector.StartMetricCollection(ctx, wg, metricDataChannel)
	// Start API server in a thread. No need for waitgroup
	go controller.api.StartWebServer(ctx)
	// Block until a kill signal
	<-ctx.Done()
	// Shutdown API server
	controller.api.Shutdown()

	return nil
}
