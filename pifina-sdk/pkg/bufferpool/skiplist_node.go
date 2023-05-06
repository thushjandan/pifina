package bufferpool

import "github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"

type nodeHeader struct {
	next []*SkipListNode
}

type SkipListNode struct {
	nodeHeader
	key       string
	nodeType  string
	singleVal driver.MetricItem
	multiVal  map[uint32]driver.MetricItem
}

func (node *SkipListNode) Key() string {
	return node.key
}

func (node *SkipListNode) NodeType() string {
	return node.nodeType
}

func (node *SkipListNode) Next() *SkipListNode {
	return node.next[0]
}

const (
	SL_NODETYPE_MULTIVAL  = "SL_NODETYPE_MULTIVAL"
	SL_NODETYPE_SINGLEVAL = "SL_NODETYPE_SINGLEVAL"
)
