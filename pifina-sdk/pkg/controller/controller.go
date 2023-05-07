package controller

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/bufferpool"
	"github.com/thushjandan/pifina/pkg/collector"
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/sink"
)

type TofinoController struct {
	ctx           context.Context
	logger        hclog.Logger
	endpoint      string
	p4name        string
	driver        *driver.TofinoDriver
	collector     *collector.MetricCollector
	sink          *sink.Sink
	metricStorage *bufferpool.SkipList
}

func NewTofinoController(logger hclog.Logger, endpoint string, p4name string, collectorServerEndpoint string, sampleInterval int) *TofinoController {
	driver := driver.NewTofinoDriver(logger)
	collector := collector.NewMetricCollector(logger, driver, sampleInterval)
	sink := sink.NewSink(logger, collectorServerEndpoint)
	return &TofinoController{
		logger:    logger.Named("controller"),
		driver:    driver,
		collector: collector,
		endpoint:  endpoint,
		p4name:    p4name,
		sink:      sink,
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
	metricDataChannel := make(chan *driver.MetricItem, 10)
	controller.collector.StartMetricCollection(ctx, wg, metricDataChannel)
	metricsSinkChannel := make(chan []*driver.MetricItem)
	wg.Add(3)
	go controller.sink.StartSink(ctx, wg, metricsSinkChannel)
	go controller.StartBufferpoolManager(ctx, wg, metricDataChannel)
	go controller.StartSampleMetrics(ctx, wg)
	// Block until a kill signal
	<-ctx.Done()
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
		}
	}
}
