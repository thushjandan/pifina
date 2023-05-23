package bufferpool

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/controller/skiplist"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
	"github.com/thushjandan/pifina/pkg/model"
)

type Bufferpool struct {
	logger        hclog.Logger
	metricStorage *skiplist.SkipList
	driver        *driver.TofinoDriver
	ts            *trafficselector.TrafficSelector
}

func NewBufferpool(logger hclog.Logger, driver *driver.TofinoDriver, ts *trafficselector.TrafficSelector) *Bufferpool {
	return &Bufferpool{
		logger: logger,
		driver: driver,
		ts:     ts,
	}
}

// Creates a buffer pool and listens on data channel for any metrics to add to buffer pool
// This thread is the only one, which adds metrics to bufferpool
func (bp *Bufferpool) StartBufferpoolManager(ctx context.Context, wg *sync.WaitGroup, c chan *model.MetricItem) {
	defer wg.Done()
	sessionIdWidth, err := bp.driver.GetSessionIdBitWidth()
	notReady := false
	bp.logger.Debug("Bit-length of sessionId variable for buffer pool", "size", sessionIdWidth)
	if err != nil {
		bp.logger.Error("Error occured during bufferpool initialization", "error", err)
		notReady = true
	}

	// Amount of static probes * variable length of sessionId = upper bound
	upperBound := int(math.Pow(2, float64(sessionIdWidth)))
	if len(bp.ts.GetTrafficSelectorCache()) > 0 {
		upperBound = upperBound * len(bp.ts.GetTrafficSelectorCache())
	}

	bp.logger.Debug("Creating bufferpool", "upperBound", upperBound)
	bp.metricStorage, err = skiplist.NewSkiplistWithMaxBound(upperBound)
	if err != nil {
		bp.logger.Error("Error occured during bufferpool initialization", "error", err)
		notReady = true
	}
	bp.logger.Debug("Bufferpool is starting to listen for new metrics")

	for {
		select {
		case newMetric := <-c:
			// Check if buffer pool is ready
			if !notReady {
				bp.logger.Trace("Adding a new metric to buffer pool", "metricName", newMetric.MetricName, "sessionId", newMetric.SessionId)
				bp.metricStorage.Set(newMetric.MetricName, newMetric.SessionId, newMetric)
			}
		case <-ctx.Done():
			bp.logger.Info("Stopping skiplist...")
			return
		}
	}
}

// Remove a metric from skiplist
// primary key is the name of the metric, subkey is the register index
func (bp *Bufferpool) RemoveMetric(key string, subKey uint32, metricType string) {
	bp.metricStorage.Remove(key, subKey, metricType)
	bp.logger.Info("Metric has been removed from bufferpool", "metric", key, "index", subKey, "type", metricType)
}

// Sample every second all the metrics
// It will retrive wait-free all available metrics from the bufferpool and reset them if needed
func (bp *Bufferpool) StartSampleMetrics(ctx context.Context, wg *sync.WaitGroup, c chan []*model.MetricItem) {
	defer wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			allItems := bp.metricStorage.GetAllAndReset()
			bp.logger.Debug("Sampled metrics", "metrics", allItems)
			c <- allItems
		case <-ctx.Done():
			bp.logger.Info("Stopping metric sampler...")
			return
		}
	}
}
