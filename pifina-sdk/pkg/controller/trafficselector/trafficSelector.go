package trafficselector

import (
	"math"
	"math/rand"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/model"
)

type TrafficSelector struct {
	logger                  hclog.Logger
	driver                  *driver.TofinoDriver
	matchSelectorEntryCache []*model.MatchSelectorEntry
	appRegisterProbes       []*model.AppRegister
	appRegisterProbesLock   sync.RWMutex
	monitoredDevPorts       []string
	monitoredDevPortsLock   sync.RWMutex
}

func NewTrafficSelector(logger hclog.Logger, d *driver.TofinoDriver) *TrafficSelector {
	return &TrafficSelector{
		logger:            logger.Named("traffic-sel"),
		driver:            d,
		appRegisterProbes: make([]*model.AppRegister, 0),
	}
}

// Add a new selector rule in the dataplane
// It will generate a new sessionId for the rule.
func (t *TrafficSelector) AddTrafficSelectorRule(newSelectorRule *model.MatchSelectorEntry) error {
	var randomSessionId uint32
	sessionBitWidth, err := t.driver.GetSessionIdBitWidth()
	if err != nil {
		return err
	}
	sessionIdsMap := make(map[uint32]struct{})
	for i := range t.matchSelectorEntryCache {
		sessionIdsMap[t.matchSelectorEntryCache[i].SessionId] = struct{}{}
	}

	// Get the upperbound for a sessionId
	max := int(math.Pow(2, float64(sessionBitWidth)))
	// Find a unique sessionId
	randomSessionId = uint32(rand.Intn(max-1) + 1)
	// Check if new sessionId is unique
	_, ok := sessionIdsMap[randomSessionId]
	// Find until new sessionId is unique
	for ok {
		t.logger.Debug("Overlap in sessionId has been found. Generating a new sessionId")
		randomSessionId = uint32(rand.Intn(max-1) + 1)
		_, ok = sessionIdsMap[randomSessionId]
	}
	newSelectorRule.SessionId = randomSessionId
	// Create rule in dataplane
	err = t.driver.AddSelectorEntry(newSelectorRule)
	if err != nil {
		return err
	}
	t.logger.Info("A new entry has been added in the dataplane", "sessionId", newSelectorRule)
	// Refresh match selector cache
	err = t.LoadSessionsFromDevice()
	return err
}

// Remove an existing selector rule from dataplane
func (t *TrafficSelector) RemoveTrafficSelectorRule(selectorRule *model.MatchSelectorEntry) error {
	// Remove rule from dataplane
	err := t.driver.RemoveSelectorEntry(selectorRule)
	if err != nil {
		return err
	}
	t.logger.Info("Selector rule has been successfully removed", "rule", selectorRule)
	// Refresh match selector cache
	err = t.LoadSessionsFromDevice()
	return err
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

// Retrieves schema of keys from the P4 schema cache.
func (t *TrafficSelector) GetTrafficSelectorSchema() ([]*model.MatchSelectorSchema, error) {
	return t.driver.GetIngressStartMatchSelectorSchema()
}
