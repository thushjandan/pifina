// Copyright (c) 2023 Thushjandan Ponnudurai
//
// Credit for the skiplist implementation to github.com/sean-public/fast-skiplist
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package skiplist

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/thushjandan/pifina/pkg/model"
)

type SkipList struct {
	root           nodeHeader
	prevNodesStack []*nodeHeader
	maxLevel       int
	randSource     rand.Source
	probability    float64
	probTable      []float64
	length         int
}

const (
	DEFAULT_PROBABILITY float64 = 1 / math.E
)

func (sl *SkipList) getCompositeKey(key string, subKey uint32, metricType string) string {
	return fmt.Sprintf("%s%s%d", key, metricType, subKey)
}

// Inserts a new item in the skiplist
// If the key exists, it ignores the request
func (sl *SkipList) Set(key string, subKey uint32, value *model.MetricItem) {
	compositeKey := sl.getCompositeKey(key, subKey, value.Type)

	prevs := sl.getPrevElementNodes(compositeKey)
	currentNode := prevs[0].next[0]

	// Key already exists
	if currentNode != nil && currentNode.key <= compositeKey {
		if currentNode.value.Type == model.METRIC_EXT_VALUE {
			currentNode.value.Value = value.Value
		} else {
			currentNode.value.Value += value.Value
		}
		currentNode.value.LastUpdated = value.LastUpdated
		return
	}

	node := &SkipListNode{
		nodeHeader: nodeHeader{next: make([]*SkipListNode, sl.randLevel())},
		key:        compositeKey,
		value:      value,
	}

	for i := range node.next {
		node.next[i] = prevs[i].next[i]
		prevs[i].next[i] = node
	}
	// Increase skiplist length
	sl.length++
}

// Get finds an element by key. It returns element pointer if found, nil if not found.
// Get is wait-free. Stale reads are possible, but this is accepted in our environment.
func (sl *SkipList) Get(key string, subKey uint32, metricType string) *SkipListNode {
	compositeKey := sl.getCompositeKey(key, subKey, metricType)

	var nextNode *SkipListNode

	prev := &sl.root

	for i := sl.maxLevel - 1; i >= 0; i-- {
		nextNode = prev.next[i]

		for nextNode != nil && compositeKey > nextNode.key {
			prev = &nextNode.nodeHeader
			nextNode = nextNode.next[i]
		}
	}

	if nextNode != nil && nextNode.key <= compositeKey {
		return nextNode
	}

	return nil
}

func (sl *SkipList) GetAllAndReset() []*model.MetricItem {
	nextNode := sl.root.next[0]

	allItems := make([]*model.MetricItem, 0, sl.length)
	timeNow := time.Now()
	// Older than 5 sec
	agedTime := timeNow.Add(-5 * time.Second)

	for nextNode != nil {
		// Copy metric struct
		newItem := *(nextNode.value)
		// Cleanup if required
		if nextNode.value.LastUpdated.Before(agedTime) {
			sl.Remove(nextNode.value.MetricName, nextNode.value.SessionId, nextNode.value.Type)
		}
		// Reset values
		if nextNode.value.Type != model.METRIC_EXT_VALUE {
			nextNode.value.Value = 0
		}

		// Sampling time
		newItem.LastUpdated = timeNow
		allItems = append(allItems, &newItem)
		nextNode = nextNode.next[0]
	}

	return allItems
}

// Remove deletes an element from the list.
// Returns removed element pointer if found, nil if not found.
func (sl *SkipList) Remove(key string, subKey uint32, metricType string) {
	compositeKey := sl.getCompositeKey(key, subKey, metricType)

	prevs := sl.getPrevElementNodes(compositeKey)

	node := prevs[0].next[0]

	// found the element, remove it
	if node != nil && node.key <= compositeKey {
		for k, v := range node.next {
			prevs[k].next[k] = v
		}
		sl.length--
		return
	}

	return
}

// Finds the previous nodes on each level relative to the current Element.
func (sl *SkipList) getPrevElementNodes(key string) []*nodeHeader {
	var nextNode *SkipListNode

	prev := &sl.root
	prevNodes := sl.prevNodesStack

	for i := sl.maxLevel - 1; i >= 0; i-- {
		nextNode = prev.next[i]

		for nextNode != nil && key > nextNode.key {
			prev = &nextNode.nodeHeader
			nextNode = nextNode.next[i]
		}

		prevNodes[i] = prev
	}

	return prevNodes
}

func (sl *SkipList) randLevel() int {
	// The random number source only has Int63(), so we have to produce a float64 from it
	r := float64(sl.randSource.Int63()) / (1 << 63)

	level := 1
	for level < sl.maxLevel && r < sl.probTable[level] {
		level++
	}

	return level
}

// probabilityTable calculates in advance the probability of a new node having a given level.
// probability is in [0, 1], MaxLevel is (0, 64]
// Returns a table of floating point probabilities that each level should be included during an insert.
func createProbabilityTable(probability float64, maxLevel int) (table []float64) {
	for i := 1; i <= maxLevel; i++ {
		prob := math.Pow(probability, float64(i-1))
		table = append(table, prob)
	}
	return table
}

// NewSkiplistWithMaxBound creates a new skip list with a given upper bound.
// The width of the sessionId variable on the dataplane is the upper bound
// e.g. if the width of the sessionId register is 7-bit, then the upper bound is 7
// maxLevel has to be int(math.Ceil(math.Log(N))) for DefaultProbability (where N is an upper bound on the
// number of elements in a skip list) and derived from the upper bound.
// See http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
// Returns a pointer to the new list.
func NewSkiplistWithMaxBound(upperBound int) (*SkipList, error) {
	if upperBound < 1 {
		return nil, errors.New("the upper bound needs to be set and must be bigger than 0!")
	}

	maxLevel := int(math.Ceil(math.Log(float64(upperBound))))
	if maxLevel < 1 {
		maxLevel = 1
	}
	probTable := createProbabilityTable(DEFAULT_PROBABILITY, maxLevel)
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &SkipList{
		root:           nodeHeader{next: make([]*SkipListNode, maxLevel)},
		prevNodesStack: make([]*nodeHeader, maxLevel),
		length:         0,
		maxLevel:       maxLevel,
		randSource:     randSrc,
		probability:    DEFAULT_PROBABILITY,
		probTable:      probTable,
	}, nil

}
