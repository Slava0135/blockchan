package node

import (
	"slava0135/blockchan/blockgen"
)

type Node struct {
	Link   Link
	Blocks []blockgen.Block
}

type Link interface {
	GetAllBlocks() []blockgen.Block
}

func NewNode(link Link) Node {
	var node = Node{}
	node.Link = link
	return node
}

func (n *Node) Start() {
	n.Blocks = n.Link.GetAllBlocks()
	if len(n.Blocks) == 0 {
		n.Blocks = append(n.Blocks, blockgen.GenerateGenesisBlock())
	}
}
