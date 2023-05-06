package bufferpool

import "github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"

type nodeHeader struct {
	next []*SkipListNode
}

type SkipListNode struct {
	nodeHeader
	key   string
	value *driver.MetricItem
}

func (node *SkipListNode) Key() string {
	return node.key
}

func (node *SkipListNode) Value() *driver.MetricItem {
	return node.value
}

func (node *SkipListNode) Next() *SkipListNode {
	return node.next[0]
}
