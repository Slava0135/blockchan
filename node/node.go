package node

import (
	"slava0135/blockchan/blockgen"
)

type Node struct {
	Link      Link
	Blocks    []blockgen.Block
	IsRunning bool
	shutdown  chan struct{}
}

type Link interface {
	GetAllBlocks() []blockgen.Block
}

func NewNode(link Link) Node {
	var node = Node{}
	node.Link = link
	node.shutdown = make(chan struct{})
	return node
}

func (n *Node) Start() {
	if (n.IsRunning) {
		panic("node was already running!")
	}
	n.Blocks = n.Link.GetAllBlocks()
	if len(n.Blocks) == 0 {
		n.Blocks = append(n.Blocks, blockgen.GenerateGenesisBlock())
	}
	n.IsRunning = true
	go n.Run()
}

func (n *Node) Run() {
	for {
		select {
		case <-n.shutdown:
			return
		}
	}
}

func (n *Node) Shutdown() {
	if (!n.IsRunning) {
		panic("node was not running!")
	}
	n.shutdown <- struct{}{}
	n.IsRunning = false
}
