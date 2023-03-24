package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"slava0135/blockchan/validate"
)

type ForkMesh struct {
	receiveChannels map[node.Fork]chan blockgen.Block
	Mentor          node.Fork
}

func (m *ForkMesh) AllExistingBlocks(from blockgen.Index) []blockgen.Block {
	var longest []blockgen.Block
	var chains = make(map[node.Fork][]blockgen.Block)
	for fork := range m.receiveChannels {
		var chain = fork.Blocks(from)
		if !validate.IsValidChain(chain) {
			continue
		}
		if len(chain) > len(longest) {
			longest = chain
		}
		chains[fork] = chain
	}
	var count = 0
	var dupForks = make(map[node.Fork]int)
	for fork, chain := range chains {
		if len(chain) == len(longest) {
			count += 1
			var processed = false
			for otherFork := range dupForks {
				if validate.AreEqualChains(chain, chains[otherFork]) {
					dupForks[otherFork] += 1
					processed = true
					break
				}
			}
			if !processed {
				dupForks[fork] = 1
			}
		}
	}
	var majorFork node.Fork
	var majorNumber = 0
	for fork, count := range dupForks {
		if count > majorNumber {
			majorFork = fork
			majorNumber = count
		}
	}
	return chains[majorFork]
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

func (m *ForkMesh) MentorFork() node.Fork {
	return m.Mentor
}

func NewForkMesh() *ForkMesh {
	var mesh = &ForkMesh{}
	mesh.receiveChannels = make(map[node.Fork]chan blockgen.Block)
	return mesh
}
