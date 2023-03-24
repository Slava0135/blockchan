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

func (mesh *testMesh) AllExistingBlocks(from blockgen.Index) []blockgen.Block {
	mesh.timesAskedForBlocks += 1
	return mesh.existingBlocks[from:]
}

func (mesh *testMesh) SendBlock(f Fork, b blockgen.Block) bool {
	mesh.receivedBlocks = append(mesh.receivedBlocks, b)
	return true
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
	node.ProcessNextBlock(blockgen.Data{})
	node.ProcessNextBlock(blockgen.Data{})
	node.Disable()
	if mesh.timesAskedForBlocks == 0 {
		t.Fatalf("node did not ask for blocks")
	}
	if len(node.Blocks(0)) < len(mesh.existingBlocks) {
		t.Fatalf("node blocks amount = %d less than amount of start blocks = %d", len(node.Blocks(0)), len(mesh.existingBlocks))
	}
	for i := range mesh.existingBlocks {
		if !node.Blocks(0)[i].Equal(mesh.existingBlocks[i]) {
			t.Fatalf("node block and start block did not match")
		}
	}
}

func TestNodeStart_Genesis(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable()
	if len(node.Blocks(0)) == 0 {
		t.Fatalf("node did not generate genesis block")
	}
	if node.Blocks(0)[0].Index != 0 {
		t.Fatalf("genesis block index is wrong")
	}
}

func TestNodeDisable(t *testing.T) {
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

func TestNodeEnable_AlreadyEnabled(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	defer func() { _ = recover() }()
	node.Enable()
	node.Enable()
	t.Fatalf("should have panicked because node was already processing next block")
}

func TestNodeDisable_AlreadyDisabled(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	defer func() { _ = recover() }()
	node.Enable()
	node.Disable()
	node.Disable()
	t.Fatalf("should have panicked because node was already disabled")
}

func TestNodeRun_SendBlocks(t *testing.T) {
	var mesh = testMesh{}
	var node = NewNode(&mesh)
	node.Enable()
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
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	node.Enable()
	go node.ProcessNextBlock(blockgen.Data{})
	mesh.chanToNode <- next
	node.Disable()
	if node.Blocks(0)[next.Index].Data != data {
		t.Fatalf("node did not accept valid received block")
	}
}

func TestNodeProcessNextBlock_RejectReceivedBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	next.Hash = []byte{}
	node.Enable()
	go node.ProcessNextBlock(blockgen.Data{})
	mesh.chanToNode <- next
	node.Disable()
	if blockgen.Index(len(node.Blocks(0))) > next.Index && node.Blocks(0)[next.Index].Data == data {
		t.Fatalf("node accepted invalid received block")
	}
}

func TestNodeProcessNextBlock_AcceptMissedBlock(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var last = mesh.existingBlocks[len(mesh.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	var nextnext = blockgen.GenerateNextFrom(next, data, nil)
	node.Enable()
	go node.ProcessNextBlock(blockgen.Data{})
	mesh.existingBlocks = append(mesh.existingBlocks, next, nextnext)
	mesh.chanToNode <- nextnext
	node.Disable()
	if mesh.timesAskedForBlocks < 2 {
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
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	var data = testData()
	var old = blockgen.GenerateNextFrom(mesh.existingBlocks[0], data, nil)
	node.Enable()
	go node.ProcessNextBlock(blockgen.Data{})
	mesh.chanToNode <- old
	node.Disable()
	if node.Blocks(0)[old.Index].Data == data {
		t.Fatalf("node accepted received old block")
	}
}

func TestNode_Connection(t *testing.T) {
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

func TestNodeProcessNextBlock_DoubleProcess(t *testing.T) {
	var mesh = newTestMesh()
	var node = NewNode(&mesh)
	node.Enable()
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
