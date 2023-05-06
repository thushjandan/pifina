package bufferpool

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
)

func TestSingleValCRUD(t *testing.T) {
	sessionIdWidth := 7
	sl, err := NewSkiplistWithMaxBound(sessionIdWidth)
	if err != nil {
		t.Fatal("Cannot initialize skip list with a specific upper bound", sessionIdWidth)
	}
	key1 := "REGISTER1"
	key2 := "REGISTER5"
	key3 := "REGISTER9"
	key4 := "REGISTER3"
	sl.SetSingleVal(key1, driver.MetricItem{Value: 100, MetricName: key1, Type: driver.METRIC_BYTES})
	sl.SetSingleVal(key2, driver.MetricItem{Value: 100, MetricName: key2, Type: driver.METRIC_BYTES})
	sl.SetSingleVal(key3, driver.MetricItem{Value: 100, MetricName: key3, Type: driver.METRIC_BYTES})
	sl.SetSingleVal(key4, driver.MetricItem{Value: 100, MetricName: key4, Type: driver.METRIC_BYTES})
	node := sl.Get(key2)
	if node.nodeType != SL_NODETYPE_SINGLEVAL {
		t.Fatal("Node type is not singleval", node.nodeType)
	}
	if node.singleVal.Value != 100 {
		t.Fatal("Wrong value. Expected 100", node.singleVal.Value)
	}

	node = sl.Get(key1)
	if node.nodeType != SL_NODETYPE_SINGLEVAL || node.singleVal.Value != 100 {
		t.Fatal("Node type is not singleval or wrong value. Expected 100", node.nodeType, node.singleVal.Value)
	}

	node = sl.Get(key2)
	if node.nodeType != SL_NODETYPE_SINGLEVAL || node.singleVal.Value != 100 {
		t.Fatal("Node type is not singleval or wrong value. Expected 100", node.nodeType, node.singleVal.Value)
	}

	node = sl.Get(key3)
	if node.nodeType != SL_NODETYPE_SINGLEVAL || node.singleVal.Value != 100 {
		t.Fatal("Node type is not singleval or wrong value. Expected 100", node.nodeType, node.singleVal.Value)
	}

	node = sl.Get(key4)
	if node.nodeType != SL_NODETYPE_SINGLEVAL || node.singleVal.Value != 100 {
		t.Fatal("Node type is not singleval or wrong value. Expected 100", node.nodeType, node.singleVal.Value)
	}

	sl.Remove(key2)
	node = sl.Get(key2)
	if node != nil {
		t.Fatal("Node has not been deleted")
	}

	if sl.Get(key1) == nil || sl.Get(key3) == nil || sl.Get(key4) == nil {
		t.Fatal("Skiplist is corrupt as other keys cannot been found")
	}

}

func TestIncrementSingleVal(t *testing.T) {
	sessionIdWidth := 7
	sl, err := NewSkiplistWithMaxBound(sessionIdWidth)
	if err != nil {
		t.Fatal("Cannot initialize skip list with a specific upper bound", sessionIdWidth)
	}
	key1 := "REGISTER1"
	key2 := "REGISTER5"
	key3 := "REGISTER9"
	sl.SetSingleVal(key1, driver.MetricItem{Value: 100, MetricName: key1, Type: driver.METRIC_BYTES})
	sl.SetSingleVal(key2, driver.MetricItem{Value: 100, MetricName: key2, Type: driver.METRIC_BYTES})
	sl.SetSingleVal(key3, driver.MetricItem{Value: 100, MetricName: key3, Type: driver.METRIC_BYTES})
	sl.SetSingleVal(key2, driver.MetricItem{Value: 100, MetricName: key2, Type: driver.METRIC_BYTES})
	node := sl.Get(key2)
	if node.nodeType != SL_NODETYPE_SINGLEVAL {
		t.Fatal("Node type is not singleval", node.nodeType)
	}
	if node.singleVal.Value != 200 {
		t.Fatal("Wrong value. Expected 200", node.singleVal.Value)
	}

	node = sl.Get(key1)
	if node.nodeType != SL_NODETYPE_SINGLEVAL || node.singleVal.Value != 100 {
		t.Fatal("Node type is not singleval or wrong value. Expected 100", node.nodeType, node.singleVal.Value)
	}

	node = sl.Get(key3)
	if node.nodeType != SL_NODETYPE_SINGLEVAL || node.singleVal.Value != 100 {
		t.Fatal("Node type is not singleval or wrong value. Expected 100", node.nodeType, node.singleVal.Value)
	}

}

func TestMultiValCRUD(t *testing.T) {
	sessionIdWidth := 7
	sl, err := NewSkiplistWithMaxBound(sessionIdWidth)
	if err != nil {
		t.Fatal("Cannot initialize skip list with a specific upper bound", sessionIdWidth)
	}
	key1 := "REGISTER1"
	key2 := "REGISTER5"
	key3 := "REGISTER9"
	key4 := "REGISTER3"
	sessionId1 := uint32(4)
	sessionId2 := uint32(6)
	sl.SetMultiVal(key1)
	sl.SetMultiVal(key2)
	sl.SetMultiVal(key3)
	sl.SetMultiVal(key4)
	node := sl.Get(key2)
	if node.nodeType != SL_NODETYPE_MULTIVAL {
		t.Fatal("Node type is not singleval", node.nodeType)
	}

	node.multiVal[sessionId1] = driver.MetricItem{SessionId: sessionId1, Value: 100, MetricName: key1, Type: driver.METRIC_BYTES}
	node.multiVal[sessionId2] = driver.MetricItem{SessionId: sessionId2, Value: 100, MetricName: key1, Type: driver.METRIC_BYTES}

	node = sl.Get(key1)
	node.multiVal[sessionId1] = driver.MetricItem{SessionId: sessionId1, Value: 100, MetricName: key1, Type: driver.METRIC_BYTES}
	// Test if there is no overwrite happening
	sl.SetMultiVal(key1)

	node = sl.Get(key2)
	if node.nodeType != SL_NODETYPE_MULTIVAL || node.multiVal[sessionId1].Value != 100 || node.multiVal[sessionId2].Value != 100 {
		t.Fatal("Node type is not multiVal or wrong value. Expected 100", node.nodeType, node.multiVal[sessionId1].Value, node.multiVal[sessionId2].Value)
	}

	node = sl.Get(key1)
	if node.nodeType != SL_NODETYPE_MULTIVAL || node.multiVal[sessionId1].Value != 100 {
		t.Fatal("Node type is not multiVal or wrong value. Expected 100", node.nodeType, node.multiVal[sessionId1].Value)
	}

	sl.Remove(key2)
	node = sl.Get(key2)
	if node != nil {
		t.Fatal("Node has not been deleted")
	}

	if sl.Get(key1) == nil || sl.Get(key3) == nil || sl.Get(key4) == nil {
		t.Fatal("Skiplist is corrupt as other keys cannot been found")
	}
}

func BenchmarkIncSet(b *testing.B) {
	b.ReportAllocs()
	sessionIdWidth := 7
	max := int(math.Pow(2, float64(sessionIdWidth)))
	sl, err := NewSkiplistWithMaxBound(sessionIdWidth)
	if err != nil {
		b.Fatal("Cannot initialize skip list with a specific upper bound", sessionIdWidth)
	}

	for i := 0; i < b.N; i++ {
		randomSessionId := rand.Intn(max-1) + 1
		sl.SetSingleVal(fmt.Sprintf("TABLE%d", randomSessionId), driver.MetricItem{SessionId: uint32(randomSessionId), Value: uint64(randomSessionId)})
	}

	b.SetBytes(int64(b.N))
}
