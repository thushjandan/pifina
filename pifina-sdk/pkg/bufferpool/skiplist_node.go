package bufferpool

import "github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"

type SkiplistNode struct {
	key   string
	value map[string]driver.MetricItem
	next  []*SkiplistNode
}

func (node *SkiplistNode) Key() string {
	return node.key
}

func (node *SkiplistNode) Next() *SkiplistNode {
	return node.next[0]
}
