package node

import (
	"slava0135/blockchan/blockgen"
	"testing"
	"time"
)

type testLink struct {
	existingBlocks      []blockgen.Block
	timesAskedForBlocks int
	receivedBlocks      []blockgen.Block
	chanToNode          chan blockgen.Block
}

func (link *testLink) AllExistingBlocks() []blockgen.Block {
	link.timesAskedForBlocks += 1
	return link.existingBlocks
}

func (link *testLink) SendBlock(b blockgen.Block) {
	link.receivedBlocks = append(link.receivedBlocks, b)
}

func (link *testLink) ReceiveChan() chan blockgen.Block {
	return link.chanToNode
}

func newTestLink() testLink {
	var link = testLink{}
	link.existingBlocks = append(link.existingBlocks, blockgen.GenerateGenesisBlock())
	for i := byte(0); i < 3; i += 1 {
		var newBlock = blockgen.GenerateNextFrom(link.existingBlocks[i], blockgen.Data{i}, nil)
		link.existingBlocks = append(link.existingBlocks, newBlock)
	}
	link.chanToNode = make(chan blockgen.Block)
	return link
}

func testData() blockgen.Data {
	var data blockgen.Data
	var text = []byte("marko zajc")
	copy(data[:], text)
	return data
}

func TestNodeStart_GetBlocks(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	node.Start()
	node.Shutdown()
	if link.timesAskedForBlocks == 0 {
		t.Fatalf("node did not ask for blocks")
	}
	if len(node.Blocks) < len(link.existingBlocks) {
		t.Fatalf("node blocks amount = %d less than amount of start blocks = %d", len(node.Blocks), len(link.existingBlocks))
	}
	for i := range link.existingBlocks {
		if node.Blocks[i] != link.existingBlocks[i] {
			t.Fatalf("node block and start block did not match")
		}
	}
}

func TestNodeStart_Genesis(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	node.Start()
	if len(node.Blocks) == 0 {
		t.Fatalf("node did not generate genesis block")
	}
	if node.Blocks[0].Index != 0 {
		t.Fatalf("genesis block index is wrong")
	}
}

func TestNodeRun_Shutdown(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	node.Start()
	if !node.IsRunning {
		t.Fatalf("node is not running after start")
	}
	node.Shutdown()
	if node.IsRunning {
		t.Fatalf("node is running after shutdown")
	}
}

func TestNodeRun_AlreadyRunning(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	defer func() { _ = recover() }()
	node.Start()
	node.Start()
	t.Fatalf("should have panicked because node was already running")
}

func TestNodeShutdown_AlreadyShutdown(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	defer func() { _ = recover() }()
	node.Start()
	node.Shutdown()
	node.Shutdown()
	t.Fatalf("should have panicked because node was already shutdown")
}

func TestNodeRun_SendBlocks(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	node.Start()
	time.Sleep(time.Second)
	node.Shutdown()
	if len(node.Blocks) != len(link.receivedBlocks) {
		t.Fatalf("node blocks amount = %d not equals amount of sent blocks = %d", len(node.Blocks), len(link.receivedBlocks))
	}
	if len(node.Blocks) == 1 {
		t.Fatalf("node did not generate any blocks except genesis")
	}
	for i, v := range link.receivedBlocks {
		if node.Blocks[i] != v {
			t.Fatalf("node did not send correct block")
		}
	}
}

func TestNodeRun_AcceptReceivedBlock(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	var data = testData()
	var last = link.existingBlocks[len(link.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	node.Start()
	time.Sleep(time.Millisecond) // wait until node start generate next block
	link.chanToNode <- next
	node.Shutdown()
	if node.Blocks[next.Index].Data != data {
		t.Fatalf("node did not accept valid received block")
	}
}

func TestNodeRun_RejectReceivedBlock(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	var data = testData()
	var last = link.existingBlocks[len(link.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	next.Hash.Reset()
	node.Start()
	link.chanToNode <- next
	node.Shutdown()
	if len(node.Blocks) > next.Index && node.Blocks[next.Index].Data == data {
		t.Fatalf("node accepted invalid received block")
	}
}

func TestNodeRun_AcceptMissedBlock(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	var data = testData()
	var last = link.existingBlocks[len(link.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data, nil)
	var nextnext = blockgen.GenerateNextFrom(next, data, nil)
	node.Start()
	time.Sleep(time.Millisecond)
	link.existingBlocks = append(link.existingBlocks, next, nextnext)
	link.chanToNode <- nextnext
	node.Shutdown()
	if link.timesAskedForBlocks < 2 {
		t.Fatalf("node did not ask for blocks when it got block ahead")
	}
	if node.Blocks[next.Index].Data != data {
		t.Fatalf("node did not saved missing block")
	}
	if node.Blocks[nextnext.Index].Data != data {
		t.Fatalf("node did not saved received block")
	}
}

func TestNodeRun_RejectMissedBlock(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	var last = link.existingBlocks[len(link.existingBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, blockgen.Data{}, nil)
	var nextnext = blockgen.GenerateNextFrom(next, blockgen.Data{}, nil)
	nextnext.Hash.Reset()
	node.Start()
	link.existingBlocks = append(link.existingBlocks, next, nextnext)
	link.chanToNode <- nextnext
	node.Shutdown()
	if link.timesAskedForBlocks > 1 {
		t.Fatalf("node asked for blocks when it got invalid block ahead")
	}
}

func TestNodeRun_IgnoreOldBlock(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	var data = testData()
	var old = blockgen.GenerateNextFrom(link.existingBlocks[0], data, nil)
	node.Start()
	link.chanToNode <- old
	node.Shutdown()
	if node.Blocks[old.Index].Data == data {
		t.Fatalf("node accepted received old block")
	}
}
