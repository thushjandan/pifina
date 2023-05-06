package bufferpool

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
)

func TestCRUD(t *testing.T) {
	sessionIdWidth := 7
	sl, err := NewSkiplistWithMaxBound(sessionIdWidth)
	if err != nil {
		t.Fatal("Cannot initialize skip list with a specific upper bound", sessionIdWidth)
	}
	keys := []string{"REGISTER1", "REGISTER5", "REGISTER9", "REGISTER3", "REGISTER12", "REGISTER34", "REGISTER7"}
	for _, key := range keys {
		sl.Set(key, 0, &driver.MetricItem{Value: 100, MetricName: key, Type: driver.METRIC_BYTES})
	}

	for _, key := range keys {
		node := sl.Get(key, 0)
		if node.value.Value != 100 {
			t.Fatal("Wrong value. Expected 100", node.value.Value)
		}
	}

	INDEX_TO_REMOVE := 2
	sl.Remove(keys[INDEX_TO_REMOVE], 0)
	for i, key := range keys {
		node := sl.Get(key, 0)
		if i == INDEX_TO_REMOVE {
			if node != nil {
				t.Fatal("Node has not been deleted")
			}
			continue
		}
		if sl.Get(key, 0) == nil {
			t.Fatal("Skiplist is corrupt as other keys cannot been found")
		}
	}

}

func BenchmarkRandomSet(b *testing.B) {
	b.ReportAllocs()
	max := int(math.Pow(2, float64(12)))
	sl, err := NewSkiplistWithMaxBound(b.N)
	if err != nil {
		b.Fatal("Cannot initialize skip list with a specific upper bound")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomSessionId := rand.Intn(max-1) + 1
		sl.Set("TABLE", uint32(randomSessionId), &driver.MetricItem{SessionId: uint32(randomSessionId), Value: uint64(randomSessionId)})
	}

	b.SetBytes(int64(b.N))
}

func BenchmarkIncSet(b *testing.B) {
	b.ReportAllocs()
	sl, err := NewSkiplistWithMaxBound(b.N)
	if err != nil {
		b.Fatal("Cannot initialize skip list with a specific upper bound")
	}

	for i := 1; i < b.N; i++ {
		tmpInt := uint32(i)
		sl.Set(fmt.Sprintf("TABLE%d", i), 0, &driver.MetricItem{SessionId: tmpInt, Value: uint64(i)})
	}

	b.SetBytes(int64(b.N))
}
