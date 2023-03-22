package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"slava0135/blockchan/validate"
)

type ForkMesh struct {
	receiveChannels map[node.Fork]chan blockgen.Block
}

func (m *ForkMesh) AllExistingBlocks(from int) []blockgen.Block {
	var longest []blockgen.Block
	for fork := range m.receiveChannels {
		if !validate.IsValidChain(fork.Blocks()) {
			continue
		}
		if len(fork.Blocks()) > len(longest) {
			longest = fork.Blocks()
		}
	}
	return longest
}

func (m *ForkMesh) SendBlock(from node.Fork, b blockgen.Block) bool {
	if !b.HasValidHash() {
		return false
	}
	for fork, ch := range m.receiveChannels {
		if fork != from {
			ch <- b
		}
	}
	return true
}

func (m *ForkMesh) ReceiveChan(f node.Fork) chan blockgen.Block {
	for fork, ch := range m.receiveChannels {
		if fork == f {
			return ch
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
