package skiplist

import "github.com/thushjandan/pifina/pkg/model"

type nodeHeader struct {
	next []*SkipListNode
}

type SkipListNode struct {
	nodeHeader
	key   string
	value *model.MetricItem
}

func (node *SkipListNode) Key() string {
	return node.key
}

func (node *SkipListNode) Value() *model.MetricItem {
	return node.value
}

func (node *SkipListNode) Next() *SkipListNode {
	return node.next[0]
}
