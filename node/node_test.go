package node

import (
	"slava0135/blockchan/blockgen"
	"testing"
	"time"
)

type testLink struct {
	startBlocks       []blockgen.Block
	wasAskedForBlocks bool
	receivedBlocks    []blockgen.Block
	chanToNode        chan blockgen.Block
}

func (link *testLink) GetAllBlocks() []blockgen.Block {
	link.wasAskedForBlocks = true
	return link.startBlocks
}

func (link *testLink) SendBlock(b blockgen.Block) {
	link.receivedBlocks = append(link.receivedBlocks, b)
}

func (link *testLink) GetReceiveChan() chan blockgen.Block {
	return link.chanToNode
}

func newTestLink() testLink {
	var link = testLink{}
	link.startBlocks = append(link.startBlocks, blockgen.GenerateGenesisBlock())
	for i := byte(0); i < 10; i += 1 {
		var newBlock = blockgen.GenerateNextFrom(link.startBlocks[i], blockgen.Data{i})
		link.startBlocks = append(link.startBlocks, newBlock)
	}
	link.chanToNode = make(chan blockgen.Block)
	return link
}

func TestNodeStart_GetBlocks(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	node.Start()
	node.Shutdown()
	if !link.wasAskedForBlocks {
		t.Fatalf("node did not ask for blocks")
	}
	if len(node.Blocks) < len(link.startBlocks) {
		t.Fatalf("node blocks amount = %d less than amount of start blocks = %d", len(node.Blocks), len(link.startBlocks))
	}
	for i := range link.startBlocks {
		if node.Blocks[i] != link.startBlocks[i] {
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
		t.Errorf("genesis block index is wrong")
	}
}

func TestNodeRun_Shutdown(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	node.Start()
	if !node.IsRunning {
		t.Errorf("node is not running after start")
	}
	node.Shutdown()
	if node.IsRunning {
		t.Errorf("node is running after shutdown")
	}
}

func TestNodeRun_AlreadyRunning(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	defer func() { _ = recover() }()
	node.Start()
	node.Start()
	t.Errorf("should have panicked because node was already running")
}

func TestNodeShutdown_AlreadyShutdown(t *testing.T) {
	var link = testLink{}
	var node = NewNode(&link)
	defer func() { _ = recover() }()
	node.Start()
	node.Shutdown()
	node.Shutdown()
	t.Errorf("should have panicked because node was already shutdown")
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
	var data blockgen.Data
	var text = []byte("marko zajc")
	copy(data[:], text)
	var last = link.startBlocks[len(link.startBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data)
	node.Start()
	link.chanToNode <- next
	node.Shutdown()
	if node.Blocks[next.Index].Data != data {
		t.Errorf("node did not accept valid received block")
	}
}

func TestNodeRun_RejectReceivedBlock(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	var data blockgen.Data
	var text = []byte("marko zajc")
	copy(data[:], text)
	var last = link.startBlocks[len(link.startBlocks)-1]
	var next = blockgen.GenerateNextFrom(last, data)
	next.Hash.Reset()
	node.Start()
	link.chanToNode <- next
	node.Shutdown()
	if len(node.Blocks) >= next.Index && node.Blocks[next.Index].Data == data {
		t.Errorf("node accepted invalid received block")
	}
}
