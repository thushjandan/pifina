package bufferpool

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Skiplist struct {
	next        []*SkiplistNode
	maxLevel    int
	Length      int
	randSource  rand.Source
	probability float64
	probTable   []float64
	mutex       sync.RWMutex
}

const (
	DEFAULT_PROBABILITY float64 = 1 / math.E
)

// NewSkiplistWithMaxBound creates a new skip list with a given upper bound.
// The width of the sessionId variable on the dataplane is the upper bound
// e.g. if the width of the sessionId register is 7-bit, then the upper bound is 7
// maxLevel has to be int(math.Ceil(math.Log(N))) for DefaultProbability (where N is an upper bound on the
// number of elements in a skip list) and derived from the upper bound.
// See http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
// Returns a pointer to the new list.
func NewSkiplistWithMaxBound(upperBound int) (*Skiplist, error) {
	if upperBound < 1 {
		return nil, errors.New("the upper bound needs to be set and must be bigger than 0!")
	}

	maxLevel := int(math.Ceil(math.Log(float64(upperBound))))
	levelsPointerList := make([]*SkiplistNode, maxLevel)

	return &Skiplist{
		next:        levelsPointerList,
		maxLevel:    maxLevel,
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		probability: DEFAULT_PROBABILITY,
		probTable:   createProbabilityTable(DEFAULT_PROBABILITY, maxLevel),
	}, nil

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
