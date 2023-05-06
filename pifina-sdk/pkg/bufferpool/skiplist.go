package bufferpool

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
)

type SkipList struct {
	root           nodeHeader
	prevNodesStack []*nodeHeader
	maxLevel       int
	randSource     rand.Source
	probability    float64
	probTable      []float64
	mutex          sync.RWMutex
}

const (
	DEFAULT_PROBABILITY float64 = 1 / math.E
)

func (sl *SkipList) getCompositeKey(key string, subKey uint32) string {
	return fmt.Sprintf("%s%d", key, subKey)
}

// Inserts a new item in the skiplist
// If the key exists, it ignores the request
// Locking is optimistic and happens only after searching.
func (sl *SkipList) Set(key string, subKey uint32, value *driver.MetricItem) {
	compositeKey := sl.getCompositeKey(key, subKey)

	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	prevs := sl.getPrevElementNodes(compositeKey)
	currentNode := prevs[0].next[0]

	// Key already exists
	if currentNode != nil && currentNode.key <= compositeKey {
		currentNode.value.Value += value.Value
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

}

// Get finds an element by key. It returns element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check for deletion after locking.
func (sl *SkipList) Get(key string, subKey uint32) *SkipListNode {
	compositeKey := sl.getCompositeKey(key, subKey)
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

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

// Remove deletes an element from the list.
// Returns removed element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check on adjacent nodes after locking.
func (sl *SkipList) Remove(key string, subKey uint32) {
	compositeKey := sl.getCompositeKey(key, subKey)
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	prevs := sl.getPrevElementNodes(compositeKey)

	node := prevs[0].next[0]

	// found the element, remove it
	if node != nil && node.key <= compositeKey {
		for k, v := range node.next {
			prevs[k].next[k] = v
		}

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
		maxLevel:       maxLevel,
		randSource:     randSrc,
		probability:    DEFAULT_PROBABILITY,
		probTable:      probTable,
	}, nil

}
