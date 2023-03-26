package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"

	log "github.com/sirupsen/logrus"
)

type Node struct {
	Mesh      Mesh
	Enabled   bool
	blocks    []blockgen.Block
	shutdown  chan struct{}
	inProcess *bool
}

type Mesh interface {
	AllExistingBlocks(from blockgen.Index) []blockgen.Block
	SendBlockBroadcast(from Fork, b blockgen.Block) bool
	SendBlockTo(to Fork, b blockgen.Block) bool
	ReceiveChan(Fork) chan blockgen.Block
	Connect(Fork)
	Disconnect(Fork)
}

type Fork interface {
	Blocks(from blockgen.Index) []blockgen.Block
}

func (n *Node) Blocks(from blockgen.Index) []blockgen.Block {
	return n.blocks[from:]
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
		log.Panic("node was already enabled!")
	}
	n.blocks = n.Mesh.AllExistingBlocks(0)
	if len(n.blocks) == 0 {
		n.blocks = append(n.blocks, blockgen.GenerateGenesisBlock())
		n.Mesh.SendBlockBroadcast(n, n.blocks[0])
	}
	n.Enabled = true
	n.Mesh.Connect(n)
}

func (n *Node) ProcessNextBlock(data blockgen.Data) {
	if *n.inProcess {
		log.Panic("node was already processing next block!")
	}
	*n.inProcess = true
	defer func() { *n.inProcess = false }()
	var cancel = false
	defer func() { cancel = true }()
	var nextBlock = make(chan blockgen.Block, 1)
	go generateNextFrom(n.blocks[len(n.blocks)-1], data, nextBlock, &cancel)
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.ReceiveChan(n):
			log.Info("node received block ", b)
			var lastThis = n.blocks[len(n.blocks)-1].Index
			var lastOther = b.Index
			if lastThis > lastOther {
				log.Info("node ignores old block")
				continue
			}
			var chain []blockgen.Block
			chain = append(chain, n.blocks...)
			if lastThis+1 == lastOther {
				log.Info("node tries to append block")
				chain = append(chain, b)
			} else {
				log.Info("node asks for all existing blocks")
				chain = append(chain, n.Mesh.AllExistingBlocks(lastThis+1)...)
			}
			if validate.IsValidChain(chain) {
				log.Info("node accepted new chain")
				n.blocks = chain
				return
			}
			log.Info("node rejected new chain")
		case b := <-nextBlock:
			log.Info("generated next block ", b)
			n.blocks = append(n.blocks, b)
			n.Mesh.SendBlockBroadcast(n, b)
			return
		}
	}
}

func generateNextFrom(block blockgen.Block, data blockgen.Data, nextBlock chan blockgen.Block, cancel *bool) {
	var b = blockgen.GenerateNextFrom(block, data, cancel)
	nextBlock <- b
}

func (n *Node) Disable() {
	if !n.Enabled {
		log.Panic("node was not enabled!")
	}
	n.Mesh.Disconnect(n)
	if *n.inProcess {
		n.shutdown <- struct{}{}
	}
	n.Enabled = false
}
