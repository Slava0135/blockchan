package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"slava0135/blockchan/validate"
	log "github.com/sirupsen/logrus"
)

type ForkMesh struct {
	receiveChannels map[node.Fork]chan blockgen.Block
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

func (m *ForkMesh) SendBlockBroadcast(from node.Fork, b blockgen.Block) bool {
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

func (m *ForkMesh) SendBlockTo(to node.Fork, b blockgen.Block) bool {
	if !b.HasValidHash() {
		return false
	}
	m.ReceiveChan(to) <- b
	return true
}

func (m *ForkMesh) ReceiveChan(f node.Fork) chan blockgen.Block {
	if ch := m.receiveChannels[f]; ch != nil {
		return ch
	}
	log.Panic("node not connected to mesh tried to get receive channel")
	panic("")
}

func (m *ForkMesh) Connect(f node.Fork) {
	m.receiveChannels[f] = make(chan blockgen.Block)
}

func (m *ForkMesh) Disconnect(f node.Fork) {
	delete(m.receiveChannels, f)
}

func NewForkMesh() *ForkMesh {
	var mesh = &ForkMesh{}
	mesh.receiveChannels = make(map[node.Fork]chan blockgen.Block)
	return mesh
}
