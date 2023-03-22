package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
)

type NodeMesh struct {
	receiveChannels map[*node.Node]chan blockgen.Block
}

func (m *NodeMesh) AllExistingBlocks() []blockgen.Block {
	for k := range m.receiveChannels {
		return k.Blocks
	}
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
	panic("node not connected to mesh tried to get receive channel")
}

func (m *NodeMesh) Connect(n *node.Node) {
	m.receiveChannels[n] = make(chan blockgen.Block)
}

func NewNodeMesh() *NodeMesh {
	var mesh = &NodeMesh{}
	mesh.receiveChannels = make(map[*node.Node]chan blockgen.Block)
	return mesh
}
