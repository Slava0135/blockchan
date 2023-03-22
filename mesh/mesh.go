package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
)

type NodeMesh struct {
	receiveChan chan blockgen.Block
}

func (m *NodeMesh) AllExistingBlocks() []blockgen.Block {
	return nil
}

func (m *NodeMesh) SendBlock(n *node.Node, b blockgen.Block) {
	m.receiveChan <- b
}

func (m *NodeMesh) ReceiveChan(n *node.Node) chan blockgen.Block {
	return m.receiveChan
}

func NewNodeMesh() *NodeMesh {
	var mesh = &NodeMesh{}
	mesh.receiveChan = make(chan blockgen.Block)
	return mesh
}
