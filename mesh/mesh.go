package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
)

type ForkMesh struct {
	receiveChannels map[node.Fork]chan blockgen.Block
}

func (m *ForkMesh) AllExistingBlocks() []blockgen.Block {
	for k := range m.receiveChannels {
		return k.Blocks()
	}
	return nil
}

func (m *ForkMesh) SendBlock(from node.Fork, b blockgen.Block) {
	for k, v := range m.receiveChannels {
		if k != from {
			v <- b
		}
	}
}

func (m *ForkMesh) ReceiveChan(n node.Fork) chan blockgen.Block {
	for k, v := range m.receiveChannels {
		if k == n {
			return v
		}
	}
	panic("node not connected to mesh tried to get receive channel")
}

func (m *ForkMesh) Connect(n node.Fork) {
	m.receiveChannels[n] = make(chan blockgen.Block)
}

func (m *ForkMesh) Disconnect(n node.Fork) {
	close(m.receiveChannels[n])
	m.receiveChannels[n] = nil
}

func NewNodeMesh() *ForkMesh {
	var mesh = &ForkMesh{}
	mesh.receiveChannels = make(map[node.Fork]chan blockgen.Block)
	return mesh
}
