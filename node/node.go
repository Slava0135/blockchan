package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"

	log "github.com/sirupsen/logrus"
)

type Node struct {
	Mesh      Mesh
	Enabled   bool
	Verified  blockgen.Index
	Name      string
	blocks    []blockgen.Block
	shutdown  chan struct{}
	inProcess *bool
}

type Mesh interface {
	NeighbourBlocks(from blockgen.Index) []blockgen.Block
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
	if int(from) < len(n.blocks) {
		return n.blocks[from:]
	}
	return nil
}

func NewNode(mesh Mesh) *Node {
	var node = &Node{}
	node.Mesh = mesh
	node.shutdown = make(chan struct{})
	node.inProcess = new(bool)
	mesh.Connect(node)
	return node
}

func (n *Node) Enable(genesis bool) {
	if n.Enabled {
		log.Panicf("node %s was already enabled!", n.Name)
	}
	if genesis {
		log.Infof("node %s generated genesis block", n.Name)
		n.blocks = append(n.blocks, blockgen.GenerateGenesisBlock())
		n.Mesh.SendBlockBroadcast(n, n.blocks[0])
	} else {
		log.Infof("node %s asks for neighbours blocks", n.Name)
		n.blocks = n.Mesh.NeighbourBlocks(0)
	}
	n.Enabled = true
}

func (n *Node) ProcessNextBlock(data blockgen.Data) {
	if *n.inProcess {
		log.Panicf("node %s was already processing next block!", n.Name)
	}
	*n.inProcess = true
	defer func() { *n.inProcess = false }()
	var cancel = make(chan struct{}, 1)
	defer func() { cancel <- struct{}{} }()
	var nextBlock = make(chan blockgen.Block, 1)
	go generateNextFrom(n.blocks[len(n.blocks)-1], data, nextBlock, cancel)
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.ReceiveChan(n):
			log.Infof("node %s received block %s", n.Name, b)
			var lastThis = n.blocks[len(n.blocks)-1].Index
			if n.Verified >= b.Index {
				log.Infof("node %s ignores old block", n.Name)
				continue
			}
			var chain []blockgen.Block
			if lastThis+1 == b.Index {
				log.Infof("node %s tries to append block", n.Name)
				chain = append(chain, n.blocks...)
				chain = append(chain, b)
				if validate.IsValidChain(chain) {
					n.blocks = append(n.blocks, b)
					log.Infof("node %s verified block with index %d", n.Name, b.Index)
					n.Verified = b.Index
					return
				} else {
					log.Warnf("node %s rejected block with %s", n.Name, b)
				}
			} else {
				log.Infof("node %s asks for missing neighbours blocks", n.Name)
				chain = append(chain, n.blocks[:n.Verified+1]...)
				chain = append(chain, n.Mesh.NeighbourBlocks(n.Verified+1)...)
				if validate.IsValidChain(chain) {
					log.Infof("node %s accepted new chain", n.Name)
					n.blocks = chain
					return
				} else {
					log.Warnf("node %s rejected new chain, last verified block: %d", n.Name, n.Verified)
				}
			}
		case b := <-nextBlock:
			log.Infof("node %s generated next block %s", n.Name, b)
			n.blocks = append(n.blocks, b)
			go n.Mesh.SendBlockBroadcast(n, b)
			return
		}
	}
}

func generateNextFrom(block blockgen.Block, data blockgen.Data, nextBlock chan blockgen.Block, cancel chan struct{}) {
	var b = blockgen.GenerateNextFrom(block, data, cancel)
	nextBlock <- b
}

func (n *Node) Disable() {
	if !n.Enabled {
		log.Panicf("node %s was not enabled!", n.Name)
	}
	if *n.inProcess {
		n.shutdown <- struct{}{}
	}
	n.Enabled = false
}
