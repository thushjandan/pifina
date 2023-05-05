package controller

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/collector"
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/sink"
)

type TofinoController struct {
	ctx       context.Context
	logger    hclog.Logger
	endpoint  string
	p4name    string
	driver    *driver.TofinoDriver
	collector *collector.MetricCollector
	sink      *sink.Sink
}

func NewTofinoController(logger hclog.Logger, endpoint string, p4name string, collectorServerEndpoint string) *TofinoController {
	driver := driver.NewTofinoDriver(logger)
	collector := collector.NewMetricCollector(logger, driver)
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
	metricDataChannel := make(chan driver.MetricItem)
	metrics := controller.collector.TriggerMetricCollection(ctx, *wg, metricDataChannel)
	err = controller.sink.Emit(metrics)
	if err != nil {
		return err
	}

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
