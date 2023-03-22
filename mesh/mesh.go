package mesh

import "slava0135/blockchan/blockgen"

type NodeMesh struct {
}

func (m *NodeMesh) AllExistingBlocks() []blockgen.Block {
	return nil
}

func (m *NodeMesh) SendBlock(b blockgen.Block) {
}

func (m *NodeMesh) ReceiveChan() chan blockgen.Block {
	return nil
}
