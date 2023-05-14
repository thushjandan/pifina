package controller

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/api"
	"github.com/thushjandan/pifina/pkg/controller/bufferpool"
	"github.com/thushjandan/pifina/pkg/controller/collector"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/controller/sink"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
	"github.com/thushjandan/pifina/pkg/model"
)

type TofinoController struct {
	ctx           context.Context
	logger        hclog.Logger
	endpoint      string
	p4name        string
	driver        *driver.TofinoDriver
	collector     *collector.MetricCollector
	ts            *trafficselector.TrafficSelector
	sink          *sink.Sink
	metricStorage *bufferpool.SkipList
	api           *api.ControllerApiServer
}

type TofinoControllerOptions struct {
	Logger                  hclog.Logger
	Endpoint                string
	P4name                  string
	CollectorServerEndpoint string
	SampleInterval          int
	APIPort                 string
}

func NewTofinoController(options *TofinoControllerOptions) *TofinoController {
	if options.Logger == nil {
		return nil
	}
	driver := driver.NewTofinoDriver(options.Logger)
	ts := trafficselector.NewTrafficSelector(options.Logger, driver)
	collector := collector.NewMetricCollector(options.Logger, driver, options.SampleInterval, ts)
	apiServer := api.NewControllerApiServer(options.Logger, options.APIPort, ts)
	sink := sink.NewSink(options.Logger, options.CollectorServerEndpoint)
	return &TofinoController{
		logger:    options.Logger.Named("controller"),
		driver:    driver,
		collector: collector,
		endpoint:  options.Endpoint,
		p4name:    options.P4name,
		sink:      sink,
		ts:        ts,
		api:       apiServer,
	}
}

func (controller *TofinoController) StartController(ctx context.Context, wg *sync.WaitGroup, connectTimeout int) error {
	controller.ctx = ctx
	// Connect to switch
	err := controller.driver.Connect(ctx, controller.endpoint, controller.p4name, connectTimeout)
	if err != nil {
		return err
	}
	// Disconnect from switch after terminating the controller
	defer controller.driver.Disconnect()
	if err != nil {
		return err
	}

	controller.EnableSyncOperationOnTables()
	metricDataChannel := make(chan *model.MetricItem, 10)
	controller.collector.StartMetricCollection(ctx, wg, metricDataChannel)
	metricsSinkChannel := make(chan []*model.MetricItem)
	wg.Add(3)
	go controller.sink.StartSink(ctx, wg, metricsSinkChannel)
	go controller.StartBufferpoolManager(ctx, wg, metricDataChannel)
	go controller.StartSampleMetrics(ctx, wg, metricsSinkChannel)
	// Start API server in a thread. No need for waitgroup
	go controller.api.StartWebServer(ctx)
	// Block until a kill signal
	<-ctx.Done()
	// Shutdown API server
	controller.api.Shutdown()
	// Close all channels
	close(metricDataChannel)
	close(metricsSinkChannel)

	return nil
}

func (controller *TofinoController) EnableSyncOperationOnTables() {
	for _, tbl := range driver.PROBE_TABLES {
		tblName := controller.driver.FindTableNameByShortName(tbl)
		if tblName == "" {
			controller.logger.Error("Cannot find full table name", "table", tbl)
			continue
		}
		err := controller.driver.EnableSyncOperationOnRegister(tblName)
		if err != nil {
			controller.logger.Error("Error occured when enabling sync operation on table", "table", tbl, "err", err)
		} else {
			controller.logger.Info("Sync Table operation has been enabled on table", "table", tbl)
		}
	}
}
