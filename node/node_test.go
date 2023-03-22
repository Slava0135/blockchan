package node

import (
	"slava0135/blockchan/blockgen"
	"testing"
)

type testMesh struct {
	existingBlocks      []blockgen.Block
	timesAskedForBlocks int
	receivedBlocks      []blockgen.Block
	chanToNode          chan blockgen.Block
	connected           bool
}

func (mesh *testMesh) AllExistingBlocks() []blockgen.Block {
	mesh.timesAskedForBlocks += 1
	return mesh.existingBlocks
}

func (mesh *testMesh) SendBlock(f Fork, b blockgen.Block) {
	mesh.receivedBlocks = append(mesh.receivedBlocks, b)
}

func (mesh *testMesh) ReceiveChan(f Fork) chan blockgen.Block {
	return mesh.chanToNode
}

func (mesh *testMesh) Connect(f Fork) {
	mesh.connected = true
}

func (mesh *testMesh) Disconnect(f Fork) {
	mesh.connected = false
}

func newTestMesh() testMesh {
	var mesh = testMesh{}
	mesh.existingBlocks = append(mesh.existingBlocks, blockgen.GenerateGenesisBlock())
	for i := byte(0); i < 3; i += 1 {
		var newBlock = blockgen.GenerateNextFrom(mesh.existingBlocks[i], blockgen.Data{i}, nil)
		mesh.existingBlocks = append(mesh.existingBlocks, newBlock)
	}
	mesh.chanToNode = make(chan blockgen.Block)
	return mesh
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
	node.Enable()
	node.ProcessNextBlock()
	node.ProcessNextBlock()
	node.Disable()
	if mesh.timesAskedForBlocks == 0 {
		t.Fatalf("node did not ask for blocks")
	}
	if len(node.Blocks()) < len(mesh.existingBlocks) {
		t.Fatalf("node blocks amount = %d less than amount of start blocks = %d", len(node.Blocks()), len(mesh.existingBlocks))
	}
	for i := range mesh.existingBlocks {
		if node.Blocks()[i] != mesh.existingBlocks[i] {
			t.Fatalf("node block and start block did not match")
		}
	}
}

func TestNodeStart_Genesis(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable()
	if len(node.Blocks()) == 0 {
		t.Fatalf("node did not generate genesis block")
	}
	if node.Blocks()[0].Index != 0 {
		t.Fatalf("genesis block index is wrong")
	}
}

func TestNodeRun_Shutdown(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable()
	if !node.Enabled {
		t.Fatalf("node is not running after start")
	}
	node.Disable()
	if node.Enabled {
		t.Fatalf("node is running after shutdown")
	}
}

func TestNodeRun_AlreadyRunning(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	defer func() { _ = recover() }()
	node.Enable()
	node.Enable()
	t.Fatalf("should have panicked because node was already running")
}

func TestNodeShutdown_AlreadyShutdown(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	defer func() { _ = recover() }()
	node.Enable()
	node.Disable()
	node.Disable()
	t.Fatalf("should have panicked because node was already shutdown")
}

func TestNodeRun_SendBlocks(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable()
	node.ProcessNextBlock()
	node.ProcessNextBlock()
	node.ProcessNextBlock()
	node.Disable()
	if len(node.Blocks()) != len(mesh.receivedBlocks) {
		t.Fatalf("node blocks amount = %d not equals amount of sent blocks = %d", len(node.Blocks()), len(mesh.receivedBlocks))
	}
	if len(node.Blocks()) == 1 {
		t.Fatalf("node did not generate any blocks except genesis")
	}
	for i, v := range mesh.receivedBlocks {
		if node.Blocks()[i] != v {
			t.Fatalf("node did not send correct block")
		}
	}
}

func TestNodeRun_AcceptReceivedBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	node.Enable()
	go node.ProcessNextBlock()
	mesh.chanToNode <- next
	node.Disable()
	if node.Blocks()[next.Index].Data != data {
		t.Fatalf("node did not accept valid received block")
	}
}

func TestNodeRun_RejectReceivedBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	next.Hash.Reset()
	node.Enable()
	go node.ProcessNextBlock()
	mesh.chanToNode <- next
	node.Disable()
	if len(node.Blocks()) > next.Index && node.Blocks()[next.Index].Data == data {
		t.Fatalf("node accepted invalid received block")
	}
}

func TestNodeRun_AcceptMissedBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	var nextnext = blockgen.GenerateNextFrom(next, data, nil)
	node.Enable()
	go node.ProcessNextBlock()
	mesh.existingBlocks = append(mesh.existingBlocks, next, nextnext)
	mesh.chanToNode <- nextnext
	node.Disable()
	if mesh.timesAskedForBlocks < 2 {
		t.Fatalf("node did not ask for blocks when it got block ahead")
	}
	if node.Blocks()[next.Index].Data != data {
		t.Fatalf("node did not saved missing block")
	}
	if node.Blocks()[nextnext.Index].Data != data {
		t.Fatalf("node did not saved received block")
	}
}

func TestNodeRun_RejectMissedBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, blockgen.Data{}, nil)
	var nextnext = blockgen.GenerateNextFrom(next, blockgen.Data{}, nil)
	nextnext.Hash.Reset()
	node.Enable()
	go node.ProcessNextBlock()
	mesh.existingBlocks = append(mesh.existingBlocks, next, nextnext)
	mesh.chanToNode <- nextnext
	node.Disable()
	if mesh.timesAskedForBlocks > 1 {
		t.Fatalf("node asked for blocks when it got invalid block ahead")
	}
}

func TestNodeRun_IgnoreOldBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var old = blockgen.GenerateNextFrom(mesh.existingBlocks[0], data, nil)
	node.Enable()
	go node.ProcessNextBlock()
	mesh.chanToNode <- old
	node.Disable()
	if node.Blocks()[old.Index].Data == data {
		t.Fatalf("node accepted received old block")
	}
}

func TestNodeRun_Connection(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	node.Enable()
	if !mesh.connected {
		t.Fatalf("node did not connect to mesh when started")
	}
	node.Disable()
	if mesh.connected {
		t.Fatalf("node did not disconnect from mesh when shutdown")
	}
}
