package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"
)

type Node struct {
	Mesh      Mesh
	IsRunning bool
	blocks    []blockgen.Block
	shutdown  chan struct{}
}

type Mesh interface {
	AllExistingBlocks() []blockgen.Block
	SendBlock(from Fork, b blockgen.Block)
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
	return node
}

func (n *Node) Start() {
	if n.IsRunning {
		panic("node was already running!")
	}
	n.blocks = n.Mesh.AllExistingBlocks()
	if len(n.blocks) == 0 {
		n.blocks = append(n.blocks, blockgen.GenerateGenesisBlock())
		n.Mesh.SendBlock(n, n.blocks[0])
	}
	n.IsRunning = true
	n.Mesh.Connect(n)
	go n.run()
}

func (n *Node) run() {
	var cancel bool
	var nextBlock = make(chan blockgen.Block)
	go generateNextFrom(n.blocks[len(n.blocks)-1], nextBlock, &cancel)
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.ReceiveChan(n):
			if !b.HasValidHash() {
				continue
			}
			if len(n.blocks) < b.Index {
				n.blocks = n.Mesh.AllExistingBlocks()
				continue
			}
			if len(n.blocks) > b.Index {
				continue
			}
			var chain []blockgen.Block
			chain = append(chain, n.blocks...)
			chain = append(chain, b)
			if validate.IsValidChain(chain) {
				n.blocks = append(n.blocks, b)
				cancel = true
			}
		case b := <-nextBlock:
			var chain []blockgen.Block
			chain = append(chain, n.blocks...)
			chain = append(chain, b)
			if validate.IsValidChain(chain) {
				n.blocks = append(n.blocks, b)
				n.Mesh.SendBlock(n, b)
			}
			cancel = false
			go generateNextFrom(n.blocks[len(n.blocks)-1], nextBlock, &cancel)
		}
	}
}

func generateNextFrom(block blockgen.Block, nextBlock chan blockgen.Block, cancel *bool) {
	nextBlock <- blockgen.GenerateNextFrom(block, blockgen.Data{}, cancel)
}

func (n *Node) Shutdown() {
	if !n.IsRunning {
		panic("node was not running!")
	}
	n.Mesh.Disconnect(n)
	n.shutdown <- struct{}{}
	n.IsRunning = false
}
