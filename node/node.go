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
		if len(n.blocks) != 0 {
			n.Verified = n.blocks[len(n.blocks)-1].Index
			log.Infof("node %s verified chain (last verified: %d)", n.Name, n.Verified)
		} else {
			log.Infof("node %s did not get blocks", n.Name)
		}
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
		case fb := <-n.Mesh.RecvChan(n):
			var b = fb.Block
			log.Infof("node %s received block %s", n.Name, b)
			if fb.Drop {
				log.Warnf("node %s was asked to drop unverified blocks (last verified: %d)", n.Name, n.Verified)
				n.blocks = n.blocks[:n.Verified+1]
			}
			if len(n.blocks) == 0 {
				if b.Index == 0 {
					log.Infof("node %s accepted genesis block", n.Name)
					n.blocks = []blockgen.Block{b}
				} else {
					log.Infof("node %s still does not have any blocks", n.Name)
					n.blocks = n.Mesh.RequestBlocks(0)
					if len(n.blocks) != 0 {
						n.Verified = n.blocks[len(n.blocks)-1].Index
						log.Infof("node %s verified chain (last verified: %d)", n.Name, n.Verified)
					}
				}
				return
			}
			var lastIndex = n.blocks[len(n.blocks)-1].Index
			if b.Index == lastIndex+1 {
				var chain []blockgen.Block
				chain = append(chain, n.blocks[n.Verified:]...)
				chain = append(chain, b)
				if validate.IsValidChain(chain) {
					log.Infof("node %s verified block with index %d", n.Name, b.Index)
					n.blocks = append(n.blocks, b)
					n.Verified = b.Index
					return
				}
			}
			if b.Index > lastIndex+1 {
				var index = n.Verified+1
				log.Infof("node %s requesting blocks from network from index %d", n.Name, index)
				var received = n.Mesh.RequestBlocks(index)
				n.blocks = n.blocks[:index]
				n.blocks = append(n.blocks, received...)
				n.Verified = n.blocks[len(n.blocks)-1].Index
				log.Infof("node %s verified chain (last verified: %d)", n.Name, n.Verified)
				return
			}
			if b.Index <= n.Verified {
				if !b.Equal(n.blocks[n.Verified]) {
					log.Infof("node %s asks sender to drop unverified blocks because it verified other chain", n.Name)
					n.Mesh.DropUnverifiedBlocks(fb.From, n.blocks[n.Verified])
				}
			}
			log.Infof("node %s ignores old block with hash %x", n.Name, b.Hash)
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
