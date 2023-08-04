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
func (bp *Bufferpool) StartBufferpoolManager(ctx context.Context, wg *sync.WaitGroup, newMetricChannel chan *model.MetricItem, sinkMetricChannel chan *model.SinkEmitCommand) {
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

	samplerTicker := time.NewTicker(1 * time.Second)
	defer samplerTicker.Stop()

	for {
		select {
		case newMetric := <-newMetricChannel:
			// Check if buffer pool is ready
			if !notReady {
				bp.logger.Trace("Adding a new metric to buffer pool", "metricName", newMetric.MetricName, "sessionId", newMetric.SessionId)
				bp.metricStorage.Set(newMetric.MetricName, newMetric.SessionId, newMetric)
			}
		case <-samplerTicker.C:
			if !notReady {
				allItems := bp.metricStorage.GetAllAndReset()
				bp.logger.Trace("Sampled metrics", "metrics", allItems)
				sinkMetricChannel <- &model.SinkEmitCommand{Metrics: allItems}
			}
		case <-ctx.Done():
			bp.logger.Info("Stopping bufferpool...")
			return
		}
	}
}
