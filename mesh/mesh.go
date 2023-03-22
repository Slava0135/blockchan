package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
)

type NodeMesh struct {
	receiveChannels map[*node.Node]chan blockgen.Block
}

func (m *NodeMesh) AllExistingBlocks() []blockgen.Block {
	return nil
}

func (m *NodeMesh) SendBlock(from *node.Node, b blockgen.Block) {
	for k, v := range m.receiveChannels {
		if k != from {
			v <- b
		}
	}
}

func (m *NodeMesh) ReceiveChan(n *node.Node) chan blockgen.Block {
	for k, v := range m.receiveChannels {
		if k == n {
			return v
		}
	}
	var new = make(chan blockgen.Block)
	m.receiveChannels[n] = new
	return new
}

func NewNodeMesh() *NodeMesh {
	var mesh = &NodeMesh{}
	mesh.receiveChannels = make(map[*node.Node]chan blockgen.Block)
	return mesh
}
