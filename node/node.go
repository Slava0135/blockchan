package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/validate"

	log "github.com/sirupsen/logrus"
)

type Node struct {
	Mesh      mesh.Mesh
	Enabled   bool
	Verified  blockgen.Index
	Name      string
	blocks    []blockgen.Block
	shutdown  chan struct{}
	inProcess *bool
}

func (n *Node) Blocks(from blockgen.Index) []blockgen.Block {
	if int(from) < len(n.blocks) {
		return n.blocks[from:]
	}
	return nil
}

func NewNode(mesh mesh.Mesh) *Node {
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
		n.blocks = n.Mesh.RequestBlocks(0)
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
	if len(n.blocks) > 0 {
		go generateNextFrom(n.blocks[len(n.blocks)-1], data, nextBlock, cancel)
	} else {
		log.Warnf("node %s does not have any blocks!", n.Name)
	}
	for {
		select {
		case <-n.shutdown:
			return
		case b := <-n.Mesh.RecvChan(n):
			log.Infof("node %s received block %s", n.Name, b)
			if len(n.blocks) == 0 {
				if b.Index == 0 {
					log.Infof("node %s accepted genesis block", n.Name)
					n.blocks = []blockgen.Block{b}
				} else {
					log.Infof("node %s still does not have any blocks", n.Name)
					n.blocks = n.Mesh.RequestBlocks(0)
				}
				return
			}
			var chain []blockgen.Block
			chain = append(chain, n.blocks[n.Verified:]...)
			chain = append(chain, b)
			if validate.IsValidChain(chain) {
				log.Infof("node %s verified block with index %d", n.Name, b.Index)
				n.blocks = append(n.blocks, b)
				n.Verified = b.Index
				return
			}
		case b := <-nextBlock:
			log.Infof("node %s generated next block %s", n.Name, b)
			n.blocks = append(n.blocks, b)
			n.Mesh.SendBlockBroadcast(n, b)
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
