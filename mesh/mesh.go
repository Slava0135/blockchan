package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
)

type ForkMesh struct {
	receiveChannels map[node.Fork]chan blockgen.Block
}

func (m *ForkMesh) AllExistingBlocks() []blockgen.Block {
	var longest []blockgen.Block
	for k := range m.receiveChannels {
		if len(k.Blocks()) > len(longest) {
			longest = k.Blocks() 
		}
	}
	return longest
}

func (m *ForkMesh) SendBlock(from node.Fork, b blockgen.Block) {
	for k, v := range m.receiveChannels {
		if k != from {
			v <- b
		}
	}
}

func (m *ForkMesh) ReceiveChan(f node.Fork) chan blockgen.Block {
	for k, v := range m.receiveChannels {
		if k == f {
			return v
		}
	}
	panic("node not connected to mesh tried to get receive channel")
}

func (m *ForkMesh) Connect(f node.Fork) {
	m.receiveChannels[f] = make(chan blockgen.Block)
}

func (m *ForkMesh) Disconnect(f node.Fork) {
	close(m.receiveChannels[f])
	delete(m.receiveChannels, f)
}

func NewForkMesh() *ForkMesh {
	var mesh = &ForkMesh{}
	mesh.receiveChannels = make(map[node.Fork]chan blockgen.Block)
	return mesh
}
