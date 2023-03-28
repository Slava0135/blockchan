package node

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/validate"
	"testing"
)

type testMesh struct {
	networkBlocks       []blockgen.Block
	timesAskedForBlocks int
	receivedBlocks      []blockgen.Block
	chanToNode          chan mesh.ForkBlock
	connected           bool
	askedToDropBlocks   bool
}

func (mesh *testMesh) RequestBlocks(from blockgen.Index) []blockgen.Block {
	mesh.timesAskedForBlocks += 1
	return mesh.networkBlocks[from:]
}

func (mesh *testMesh) SendBlockBroadcast(f mesh.Fork, b blockgen.Block) bool {
	mesh.receivedBlocks = append(mesh.receivedBlocks, b)
	return true
}

func (mesh *testMesh) SendBlockTo(f mesh.Fork, b mesh.ForkBlock) bool {
	return true
}

func (mesh *testMesh) RecvChan(f mesh.Fork) chan mesh.ForkBlock {
	return mesh.chanToNode
}

func (mesh *testMesh) Connect(f mesh.Fork) {
	mesh.connected = true
}

func (mesh *testMesh) Disconnect(f mesh.Fork) {
	mesh.connected = false
}

func (mesh *testMesh) DropUnverifiedBlocks(f mesh.Fork, b blockgen.Block) {
	mesh.askedToDropBlocks = true
}

func newTestMesh() testMesh {
	var m = testMesh{}
	m.networkBlocks = append(m.networkBlocks, blockgen.GenerateGenesisBlock())
	for i := byte(0); i < 3; i += 1 {
		var newBlock = blockgen.GenerateNextFrom(m.networkBlocks[i], blockgen.Data{i}, nil)
		m.networkBlocks = append(m.networkBlocks, newBlock)
	}
	m.chanToNode = make(chan mesh.ForkBlock)
	return m
}

func testData() blockgen.Data {
	var data blockgen.Data
	var text = []byte("marko zajc")
	copy(data[:], text)
	return data
}

func TestNodeStart_GetBlocks(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	node.Enable(false)
	node.ProcessNextBlock(blockgen.Data{})
	node.ProcessNextBlock(blockgen.Data{})
	node.Disable()
	if mesh.timesAskedForBlocks == 0 {
		t.Fatalf("node did not ask for blocks")
	}
	if len(node.Blocks(0)) < len(mesh.networkBlocks) {
		t.Fatalf("node blocks amount = %d less than amount of start blocks = %d", len(node.Blocks(0)), len(mesh.networkBlocks))
	}
	for i := range mesh.networkBlocks {
		if !node.Blocks(0)[i].Equal(mesh.networkBlocks[i]) {
			t.Fatalf("node block and start block did not match")
		}
	}
}

func TestNodeStart_Genesis(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable(true)
	if len(node.Blocks(0)) == 0 {
		t.Fatalf("node did not generate genesis block")
	}
	if node.Blocks(0)[0].Index != 0 {
		t.Fatalf("genesis block index is wrong")
	}
}

func TestNodeDisable(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	node.Enable(true)
	if !node.Enabled {
		t.Fatalf("node is not running after start")
	}
	go node.ProcessNextBlock(blockgen.Data{})
	m.RecvChan(node) <- mesh.ForkBlock{}
	node.Disable()
	if node.Enabled {
		t.Fatalf("node is running after shutdown")
	}
}

func TestNodeEnable_AlreadyEnabled(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	defer func() { _ = recover() }()
	node.Enable(true)
	node.Enable(true)
	t.Fatalf("should have panicked because node was already processing next block")
}

func TestNodeDisable_AlreadyDisabled(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	defer func() { _ = recover() }()
	node.Enable(true)
	node.Disable()
	node.Disable()
	t.Fatalf("should have panicked because node was already disabled")
}

func TestNodeRun_SendBlocks(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable(true)
	node.ProcessNextBlock(blockgen.Data{})
	node.ProcessNextBlock(blockgen.Data{})
	node.ProcessNextBlock(blockgen.Data{})
	node.Disable()
	if len(node.Blocks(0)) != len(mesh.receivedBlocks) {
		t.Fatalf("node blocks amount = %d not equals amount of sent blocks = %d", len(node.Blocks(0)), len(mesh.receivedBlocks))
	}
	if len(node.Blocks(0)) == 1 {
		t.Fatalf("node did not generate any blocks except genesis")
	}
	for i, v := range mesh.receivedBlocks {
		if !node.Blocks(0)[i].Equal(v) {
			t.Fatalf("node did not send correct block")
		}
	}
}

func TestNodeProcessNextBlock_AcceptReceivedBlock(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	var data = testData()
	var last = m.networkBlocks[len(m.networkBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	node.Enable(false)
	go func() { m.chanToNode <- mesh.ForkBlock{Block: next} }()
	node.ProcessNextBlock(blockgen.Data{})
	if node.Verified != next.Index || node.Blocks(0)[next.Index].Data != data {
		t.Fatalf("node did not accept valid received block")
	}
}

func TestNodeProcessNextBlock_RejectReceivedBlock(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	var data = testData()
	var last = m.networkBlocks[len(m.networkBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	next.Hash = []byte{}
	node.Enable(true)
	go func() { m.chanToNode <- mesh.ForkBlock{Block: next} }()
	node.ProcessNextBlock(blockgen.Data{})
	if node.Verified == next.Index {
		t.Fatalf("node accepted invalid received block")
	}
}

func TestNodeProcessNextBlock_AcceptMissedBlock(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	var data = testData()
	var last = m.networkBlocks[len(m.networkBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	var nextnext = blockgen.GenerateNextFrom(next, data, nil)
	node.Enable(false)
	m.networkBlocks = append(m.networkBlocks, next, nextnext)
	go func() { m.chanToNode <- mesh.ForkBlock{Block: nextnext} }()
	node.ProcessNextBlock(blockgen.Data{})
	if m.timesAskedForBlocks < 2 {
		t.Fatalf("node did not ask for blocks when it got block ahead")
	}
	if node.Blocks(0)[next.Index].Data != data {
		t.Fatalf("node did not saved missing block")
	}
	if node.Blocks(0)[nextnext.Index].Data != data {
		t.Fatalf("node did not saved received block")
	}
}

func TestNodeProcessNextBlock_IgnoreOldBlock(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	var data = testData()
	var old = blockgen.GenerateNextFrom(m.networkBlocks[0], data, nil)
	node.Enable(false)
	go func() { m.chanToNode <- mesh.ForkBlock{Block: old} }()
	node.ProcessNextBlock(blockgen.Data{})
	if node.Blocks(0)[old.Index].Data == data {
		t.Fatalf("node accepted old block")
	}
}

func TestNode_Connection(t *testing.T) {
	var mesh = newTestMesh()
	var _ = NewNode(&mesh)
	if !mesh.connected {
		t.Fatalf("node did not connect to mesh when created")
	}
}

func TestNodeProcessNextBlock_DoubleProcess(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	node.Enable(true)
	var success = new(bool)
	go func() {
		defer func() { _ = recover() }()
		node.ProcessNextBlock(blockgen.Data{})
		*success = true
	}()
	node.ProcessNextBlock(blockgen.Data{})
	if *success {
		t.Fatalf("node did not panic because of double processing")
	}
}

func TestNodeBlocks_OutOfRange(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	node.Blocks(42)
}

func TestNodeEnable_NoRequestBlocks(t *testing.T) {
	var m = newTestMesh()
	m.networkBlocks = nil
	var node = NewNode(&m)
	node.Enable(false)
	var genesis = blockgen.GenerateGenesisBlock()
	go func() {
		m.RecvChan(node) <- mesh.ForkBlock{Block: genesis}
	}()
	node.ProcessNextBlock(blockgen.Data{})
	if !validate.AreEqualChains(node.blocks, []blockgen.Block{genesis}) {
		t.Fatalf("node did not accept genesis block")
	}
}

func TestNodeProcessNextBlock_NoBlocks(t *testing.T) {
	var m = newTestMesh()
	m.networkBlocks = nil
	var node = NewNode(&m)
	node.Enable(false)
	var genesis = blockgen.GenerateGenesisBlock()
	var next = blockgen.GenerateNextFrom(genesis, blockgen.Data{}, nil)
	go func() {
		m.RecvChan(node) <- mesh.ForkBlock{Block: next}
	}()
	m.networkBlocks = []blockgen.Block{genesis, next}
	node.ProcessNextBlock(blockgen.Data{})
	if !validate.AreEqualChains(m.networkBlocks, node.blocks) {
		t.Fatalf("empty node did not ask mesh for blocks when got non genesis block")
	}
}

func TestNodeProcessNextBlock_AskToDropUnverified(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	node.Enable(false)
	var prev = m.networkBlocks[len(m.networkBlocks)-2]
	var other = blockgen.GenerateNextFrom(prev, blockgen.Data{42}, nil)
	go func() {
		m.RecvChan(node) <- mesh.ForkBlock{Block: other}
	}()
	node.ProcessNextBlock(blockgen.Data{})
	if !m.askedToDropBlocks {
		t.Fatalf("node did not ask other fork to drop their block")
	}
}

func TestNodeProcessNextBlock_DropBlocks(t *testing.T) {
	var m = newTestMesh()
	var node = NewNode(&m)
	node.Enable(false)
	node.ProcessNextBlock(blockgen.Data{})
	node.ProcessNextBlock(blockgen.Data{})
	var next = blockgen.GenerateNextFrom(m.networkBlocks[len(m.networkBlocks)-1], blockgen.Data{}, nil)
	m.networkBlocks = append(m.networkBlocks, next)
	go func() {
		m.RecvChan(node) <- mesh.ForkBlock{Block: next, Drop: true}
	}()
	node.ProcessNextBlock(blockgen.Data{})
	if !validate.AreEqualChains(m.networkBlocks, node.blocks) {
		t.Fatalf("node did not drop blocks when asked")
	}
}
