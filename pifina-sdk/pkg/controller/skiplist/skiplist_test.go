package skiplist

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/thushjandan/pifina/pkg/model"
)

func TestCRUD(t *testing.T) {
	sessionIdWidth := 7
	sl, err := NewSkiplistWithMaxBound(sessionIdWidth)
	if err != nil {
		t.Fatal("Cannot initialize skip list with a specific upper bound", sessionIdWidth)
	}
	keys := []string{"REGISTER1", "REGISTER5", "REGISTER9", "REGISTER3", "REGISTER12", "REGISTER34", "REGISTER7"}
	for _, key := range keys {
		sl.Set(key, 0, &model.MetricItem{Value: 100, MetricName: key, Type: model.METRIC_BYTES})
	}

	for _, key := range keys {
		node := sl.Get(key, 0, model.METRIC_BYTES)
		if node.value.Value != 100 {
			t.Fatal("Wrong value. Expected 100", node.value.Value)
		}
	}

	INDEX_TO_REMOVE := 2
	sl.Remove(keys[INDEX_TO_REMOVE], 0, model.METRIC_BYTES)
	for i, key := range keys {
		node := sl.Get(key, 0, model.METRIC_BYTES)
		if i == INDEX_TO_REMOVE {
			if node != nil {
				t.Fatal("Node has not been deleted")
			}
			continue
		}
		if sl.Get(key, 0, model.METRIC_BYTES) == nil {
			t.Fatal("Skiplist is corrupt as other keys cannot been found")
		}
	}

}

func BenchmarkRandomSet(b *testing.B) {
	b.ReportAllocs()
	// Using 2^12
	max := int(math.Pow(2, float64(12)))
	sl, err := NewSkiplistWithMaxBound(b.N)
	if err != nil {
		b.Fatal("Cannot initialize skip list with a specific upper bound")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomSessionId := rand.Intn(max-1) + 1
		sl.Set("TABLE", uint32(randomSessionId), &model.MetricItem{SessionId: uint32(randomSessionId), Value: uint64(randomSessionId), Type: model.METRIC_BYTES})
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
		sl.Set(fmt.Sprintf("TABLE%d", i), 0, &model.MetricItem{SessionId: tmpInt, Value: uint64(i), Type: model.METRIC_BYTES})
	}

	b.SetBytes(int64(b.N))
}

func BenchmarkGetRandomSet(b *testing.B) {
	b.ReportAllocs()
	max := int(math.Pow(2, float64(12)))
	sl, err := NewSkiplistWithMaxBound(b.N)
	if err != nil {
		b.Fatal("Cannot initialize skip list with a specific upper bound")
	}

	randomIds := make([]int, 0, b.N)

	for i := 1; i < b.N; i++ {
		randomSessionId := rand.Intn(max-1) + 1
		randomIds = append(randomIds, randomSessionId)
		sl.Set("TABLE", uint32(randomSessionId), &model.MetricItem{SessionId: uint32(randomSessionId), Value: uint64(randomSessionId), Type: model.METRIC_BYTES})
	}

	b.ResetTimer()

	for _, randomId := range randomIds {
		sl.Get("TABLE", uint32(randomId), model.METRIC_BYTES)
	}

	b.SetBytes(int64(b.N))
}

func BenchmarkGetIncSet(b *testing.B) {
	b.ReportAllocs()
	sl, err := NewSkiplistWithMaxBound(b.N)
	if err != nil {
		b.Fatal("Cannot initialize skip list with a specific upper bound")
	}

	for i := 1; i < b.N; i++ {
		tmpInt := uint32(i)
		sl.Set(fmt.Sprintf("TABLE%d", i), 0, &model.MetricItem{SessionId: tmpInt, Value: uint64(i), Type: model.METRIC_BYTES})
	}

	b.ResetTimer()

	for i := 1; i < b.N; i++ {
		sl.Get(fmt.Sprintf("TABLE%d", i), 0, model.METRIC_BYTES)
	}

	b.SetBytes(int64(b.N))
}
