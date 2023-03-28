package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"

	log "github.com/sirupsen/logrus"
)

type Mesh interface {
	RequestBlocks(from blockgen.Index) []blockgen.Block
	SendBlockBroadcast(from Fork, b blockgen.Block) bool
	SendBlockTo(to Fork, b ForkBlock) bool
	RecvChan(Fork) chan ForkBlock
	Connect(Fork)
	Disconnect(Fork)
	DropUnverifiedBlocks(Fork)
}

type Fork interface {
	Blocks(from blockgen.Index) []blockgen.Block
}

type ForkBlock struct {
	Block blockgen.Block
	From  Fork
}

type ForkMesh struct {
	receiveChannels map[Fork]chan ForkBlock
}

func (m *ForkMesh) RequestBlocks(from blockgen.Index) []blockgen.Block {
	var longest []blockgen.Block
	var chains = make(map[Fork][]blockgen.Block)
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
	var dupForks = make(map[Fork]int)
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
	var majorFork Fork
	var majorNumber = 0
	for fork, count := range dupForks {
		if count > majorNumber {
			majorFork = fork
			majorNumber = count
		}
	}
	return chains[majorFork]
}

func (m *ForkMesh) SendBlockBroadcast(from Fork, b blockgen.Block) bool {
	if !b.HasValidHash() {
		return false
	}
	for fork, ch := range m.receiveChannels {
		if fork != from {
			ch := ch
			from := from
			go func() {
				ch <- ForkBlock{Block: b, From: from}
			}()
		}
	}
	return true
}

func (m *ForkMesh) SendBlockTo(to Fork, b ForkBlock) bool {
	if !b.Block.HasValidHash() {
		return false
	}
	m.RecvChan(to) <- b
	return true
}

func (m *ForkMesh) RecvChan(f Fork) chan ForkBlock {
	if ch, ok := m.receiveChannels[f]; ok {
		return ch
	}
	log.Panic("node not connected to mesh tried to get receive channel")
	panic("")
}

func (m *ForkMesh) Connect(f Fork) {
	m.receiveChannels[f] = make(chan ForkBlock)
}

func (m *ForkMesh) Disconnect(f Fork) {
	delete(m.receiveChannels, f)
}

func (m *ForkMesh) DropUnverifiedBlocks(f Fork) {
}

func NewForkMesh() *ForkMesh {
	var mesh = &ForkMesh{}
	mesh.receiveChannels = make(map[Fork]chan ForkBlock)
	return mesh
}
