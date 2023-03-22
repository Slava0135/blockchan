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
	SendBlock(from *Node, b blockgen.Block)
	ReceiveChan(*Node) chan blockgen.Block
	Connect(*Node)
	Disconnect(*Node)
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
	n.Blocks = n.Mesh.AllExistingBlocks()
	if len(n.Blocks) == 0 {
		n.Blocks = append(n.Blocks, blockgen.GenerateGenesisBlock())
		n.Mesh.SendBlock(n, n.Blocks[0])
	}
	n.IsRunning = true
	n.Mesh.Connect(n)
	go n.run()
}

func (n *Node) run() {
	var cancel = make(chan struct{})
	var nextBlock = make(chan blockgen.Block)
	go generateNextFrom(n.Blocks[len(n.Blocks)-1], nextBlock, cancel)
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.ReceiveChan(n):
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
				n.Mesh.SendBlock(n, b)
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
	n.Mesh.Disconnect(n)
	n.shutdown <- struct{}{}
	n.IsRunning = false
}
