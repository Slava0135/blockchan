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
}

func (link *testLink) GetAllBlocks() []blockgen.Block {
	link.wasAskedForBlocks = true
	return link.startBlocks
}

func (link *testLink) SendBlock(b blockgen.Block) {
	link.receivedBlocks = append(link.receivedBlocks, b)
}

func newTestLink() testLink {
	var link = testLink{}
	link.startBlocks = append(link.startBlocks, blockgen.GenerateGenesisBlock())
	for i := byte(0); i < 10; i += 1 {
		var newBlock = blockgen.GenerateNextFrom(link.startBlocks[i], blockgen.Data{i})
		link.startBlocks = append(link.startBlocks, newBlock)
	}
	return link
}

func TestNodeStart_GetBlocks(t *testing.T) {
	var link = newTestLink()
	var node = NewNode(&link)
	node.Start()
	if !link.wasAskedForBlocks {
		t.Errorf("node did not ask for blocks")
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
