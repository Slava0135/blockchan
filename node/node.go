package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"
)

type Node struct {
	Mesh      Mesh
	Enabled   bool
	blocks    []blockgen.Block
	shutdown  chan struct{}
	inProcess *bool
}

type Mesh interface {
	AllExistingBlocks(from int) []blockgen.Block
	SendBlock(from Fork, b blockgen.Block) bool
	ReceiveChan(Fork) chan blockgen.Block
	Connect(Fork)
	Disconnect(Fork)
}

type Fork interface {
	Blocks() []blockgen.Block
}

func (n *Node) Blocks() []blockgen.Block {
	return n.blocks
}

func NewNode(mesh Mesh) *Node {
	var node = &Node{}
	node.Mesh = mesh
	node.shutdown = make(chan struct{})
	node.inProcess = new(bool)
	return node
}

func (n *Node) Enable() {
	if n.Enabled {
		panic("node was already enabled!")
	}
	n.blocks = n.Mesh.AllExistingBlocks(0)
	if len(n.blocks) == 0 {
		n.blocks = append(n.blocks, blockgen.GenerateGenesisBlock())
		n.Mesh.SendBlock(n, n.blocks[0])
	}
	n.Enabled = true
	n.Mesh.Connect(n)
}

func (n *Node) ProcessNextBlock() {
	if *n.inProcess {
		panic("node was already processing next block!")
	}
	*n.inProcess = true
	defer func() { *n.inProcess = false }()
	var cancel = false
	defer func() { cancel = true }()
	var nextBlock = make(chan blockgen.Block, 1)
	go generateNextFrom(n.blocks[len(n.blocks)-1], nextBlock, &cancel)
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.ReceiveChan(n):
			if len(n.blocks) > b.Index {
				continue
			}
			if len(n.blocks) < b.Index {
				n.blocks = n.Mesh.AllExistingBlocks(0)
				return
			}
			var chain []blockgen.Block
			chain = append(chain, n.blocks...)
			chain = append(chain, b)
			if validate.IsValidChain(chain) {
				n.blocks = append(n.blocks, b)
				return
			}
		case b := <-nextBlock:
			n.blocks = append(n.blocks, b)
			n.Mesh.SendBlock(n, b)
			return
		}
	}
}

func generateNextFrom(block blockgen.Block, nextBlock chan blockgen.Block, cancel *bool) {
	var b = blockgen.GenerateNextFrom(block, blockgen.Data{}, cancel)
	nextBlock <- b
}

func (n *Node) Disable() {
	if !n.Enabled {
		panic("node was not enabled!")
	}
	n.Mesh.Disconnect(n)
	if *n.inProcess {
		n.shutdown <- struct{}{}
	}
	n.Enabled = false
}
