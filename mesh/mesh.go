package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/validate"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Mesh interface {
	RequestBlocks(from blockgen.Index, caller Fork) []blockgen.Block
	SendBlockBroadcast(from Fork, b blockgen.Block) bool
	SendBlockTo(to Fork, b ForkBlock) bool
	RecvChan(Fork) chan ForkBlock
	Connect(Fork)
	Disconnect(Fork)
	DropUnverifiedBlocks(Fork, blockgen.Block)
}

type Fork interface {
	Blocks(from blockgen.Index) []blockgen.Block
}

type ForkBlock struct {
	Block blockgen.Block
	From  Fork
	Drop  bool
}

type ForkMesh struct {
	receiveChannels map[Fork]chan ForkBlock
	mu              sync.Mutex
}

func (m *ForkMesh) RequestBlocks(from blockgen.Index, caller Fork) []blockgen.Block {
	var longest []blockgen.Block
	for fork := range m.receiveChannels {
		if fork == caller {
			continue
		}
		var chain = fork.Blocks(from)
		if !validate.IsValidChain(chain) {
			continue
		}
		if len(chain) > len(longest) {
			longest = chain
		}
	}
	return longest
}

func (m *ForkMesh) SendBlockBroadcast(from Fork, b blockgen.Block) bool {
	if !b.HasValidHash() {
		return false
	}
	for fork, ch := range m.receiveChannels {
		if fork != from {
			ch <- ForkBlock{Block: b, From: from}
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
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receiveChannels[f] = make(chan ForkBlock, 13)
}

func (m *ForkMesh) Disconnect(f Fork) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.receiveChannels, f)
}

func (m *ForkMesh) DropUnverifiedBlocks(f Fork, b blockgen.Block) {
	m.RecvChan(f) <- ForkBlock{Block: b, Drop: true}
}

func NewForkMesh() *ForkMesh {
	var mesh = &ForkMesh{}
	mesh.receiveChannels = make(map[Fork]chan ForkBlock)
	return mesh
}
