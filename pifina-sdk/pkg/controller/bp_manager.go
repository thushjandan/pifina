package controller

import (
	"context"
	"math"
	"sync"

	"github.com/thushjandan/pifina/pkg/bufferpool"
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
)

func (ctrl *TofinoController) StartBufferpoolManager(ctx context.Context, wg *sync.WaitGroup, c chan *driver.MetricItem) {
	defer wg.Done()
	sessionIdWidth, err := ctrl.driver.GetSessionIdBitWidth()
	notReady := false
	ctrl.logger.Debug("Bit-length of sessionId variable for buffer pool", "size", sessionIdWidth)
	if err != nil {
		ctrl.logger.Error("Error occured during bufferpool initialization", "error", err)
		notReady = true
	}

	// Amount of static probes * variable length of sessionId = upper bound
	upperBound := len(ctrl.collector.GetSessionIdCache()) * int(math.Pow(2, float64(sessionIdWidth)))
	ctrl.logger.Debug("Creating bufferpool", "upperBound", upperBound)
	ctrl.metricStorage, err = bufferpool.NewSkiplistWithMaxBound(upperBound)
	if err != nil {
		ctrl.logger.Error("Error occured during bufferpool initialization", "error", err)
		notReady = true
	}
	ctrl.logger.Debug("Bufferpool is starting to listen for new metrics")

	for {
		select {
		case newMetric := <-c:
			// Check if buffer pool is ready
			if !notReady {
				ctrl.logger.Debug("Adding a new metric to buffer pool", "metricName", newMetric.MetricName, "sessionId", newMetric.SessionId)
				ctrl.addMetricToStorage(ctx, newMetric)
			}
		case <-ctx.Done():
			ctrl.logger.Info("Stopping Bufferpool...")
			return
		}
	}
}

func (ctrl *TofinoController) addMetricToStorage(ctx context.Context, newMetricList *driver.MetricItem) {
	if ctrl.driver.IsInProbeTable(newMetricList.MetricName) {
		ctrl.metricStorage.Set(newMetricList.MetricName, newMetricList.SessionId, newMetricList)
	} else {
		ctrl.metricStorage.Set(newMetricList.MetricName, uint32(0), newMetricList)
	}
}
