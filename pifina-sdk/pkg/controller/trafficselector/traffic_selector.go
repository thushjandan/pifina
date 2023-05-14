package trafficselector

import (
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/model"
)

type TrafficSelector struct {
	logger                  hclog.Logger
	driver                  *driver.TofinoDriver
	matchSelectorEntryCache []*model.MatchSelectorEntry
}

func NewTrafficSelector(logger hclog.Logger, d *driver.TofinoDriver) *TrafficSelector {
	return &TrafficSelector{
		logger: logger.Named("traffic-sel"),
		driver: d,
	}
}

func (t *TrafficSelector) AddTrafficSelectorRule() {
	panic("not implemented")
}

// Retrieve the match selector entries and extract the session IDs.
func (t *TrafficSelector) LoadSessionsFromDevice() error {
	matchSelectorEntries, err := t.driver.GetKeysFromMatchSelectors()
	if err != nil {
		return err
	}
	t.matchSelectorEntryCache = matchSelectorEntries

	return nil
}

func (t *TrafficSelector) GetTrafficSelectorCache() []*model.MatchSelectorEntry {
	// If sessionId cache is empty, then refresh the cache
	if t.matchSelectorEntryCache == nil {
		err := t.LoadSessionsFromDevice()
		if err != nil {
			t.logger.Error("Error occured during collection. Cannot retrieve sessionIds from Ingress Start Match table", "err", err)
			return nil
		}
	}
	return t.matchSelectorEntryCache
}

// Returns just a list of sessionIds from the cache
func (t *TrafficSelector) GetSessionIdCache() []uint32 {
	sessionIds := make([]uint32, 0, len(t.matchSelectorEntryCache))

	for i := range t.matchSelectorEntryCache {
		sessionIds = append(sessionIds, t.matchSelectorEntryCache[i].SessionId)
	}

	return sessionIds
}
