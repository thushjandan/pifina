package controller

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/collector"
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
)

type TofinoController struct {
	ctx       context.Context
	logger    hclog.Logger
	endpoint  string
	p4name    string
	driver    *driver.TofinoDriver
	collector *collector.MetricCollector
}

func NewTofinoController(logger hclog.Logger, endpoint string, p4name string) *TofinoController {
	driver := driver.NewTofinoDriver(logger)
	collector := collector.NewMetricCollector(logger, driver)
	return &TofinoController{
		logger:    logger.Named("controller"),
		driver:    driver,
		collector: collector,
		endpoint:  endpoint,
		p4name:    p4name,
	}
}

func (controller *TofinoController) StartController(ctx context.Context, connectTimeout int) error {
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
	controller.collector.TriggerMetricCollection()

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
