package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"
)

type Node struct {
	Mesh      Mesh
	Blocks    []blockgen.Block
	IsRunning bool
	shutdown  chan struct{}
}

type Mesh interface {
	AllExistingBlocks() []blockgen.Block
	SendBlock(blockgen.Block)
	ReceiveChan() chan blockgen.Block
}

func NewNode(mesh Mesh) Node {
	var node = Node{}
	node.Mesh = mesh
	node.shutdown = make(chan struct{})
	return node
}

func (n *Node) Start() {
	if n.IsRunning {
		panic("node was already running!")
	}
	n.Blocks = n.Mesh.AllExistingBlocks()
	if len(n.Blocks) == 0 {
		n.Blocks = append(n.Blocks, blockgen.GenerateGenesisBlock())
		n.Mesh.SendBlock(n.Blocks[0])
	}
	n.IsRunning = true
	go n.Run()
}

func (n *Node) Run() {
	var cancel = make(chan struct{})
	var nextBlock = make(chan blockgen.Block)
	go generateNextFrom(n.Blocks[len(n.Blocks)-1], nextBlock, cancel)
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.ReceiveChan():
			if !b.HasValidHash() {
				continue
			}
			if len(n.Blocks) < b.Index {
				n.Blocks = n.Mesh.AllExistingBlocks()
				continue
			}
			if len(n.Blocks) > b.Index {
				continue
			}
			var chain []blockgen.Block
			chain = append(chain, n.Blocks...)
			chain = append(chain, b)
			if validate.IsValidChain(chain) {
				n.Blocks = append(n.Blocks, b)
				cancel <- struct{}{}
			}
		case b := <-nextBlock:
			if !b.HasValidHash() {
				continue
			}
			var chain []blockgen.Block
			chain = append(chain, n.Blocks...)
			chain = append(chain, b)
			if validate.IsValidChain(chain) {
				n.Blocks = append(n.Blocks, b)
				n.Mesh.SendBlock(b)
			}
			go generateNextFrom(n.Blocks[len(n.Blocks)-1], nextBlock, cancel)
		}
	}
}

func generateNextFrom(block blockgen.Block, nextBlock chan blockgen.Block, cancel chan struct{}) {
	nextBlock <- blockgen.GenerateNextFrom(block, blockgen.Data{}, cancel)
}

func (n *Node) Shutdown() {
	if !n.IsRunning {
		panic("node was not running!")
	}
	n.shutdown <- struct{}{}
	n.IsRunning = false
}
