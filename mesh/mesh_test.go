package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"testing"
	"time"
)

type testFork struct {
	blocks []blockgen.Block
}

func (f *testFork) Blocks(from blockgen.Index) []blockgen.Block {
	return f.blocks[from:]
}

func newTestFork(mesh node.Mesh) *testFork {
	var fork = &testFork{}
	mesh.Connect(fork)
	return fork
}

func TestForkMesh_Interface(t *testing.T) {
	var _ node.Mesh = &ForkMesh{}
}

func TestForkMesh_SendAndReceive(t *testing.T) {
	var mesh = NewForkMesh()
	var forkFrom = newTestFork(mesh)
	var forkTo = newTestFork(mesh)
	var sent = blockgen.GenerateGenesisBlock()
	go mesh.SendBlockBroadcast(forkFrom, sent)
	var received = <-mesh.ReceiveChan(forkTo)
	if !sent.Equal(received) {
		t.Fatalf("block was not sent")
	}
}

func TestForkMeshSendBlockBroadcast_Loopback(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlockBroadcast(fork, block)
	time.Sleep(time.Second)
	select {
	case <-mesh.ReceiveChan(fork):
		t.Fatalf("mesh tried to send block back to sender")
	default:
	}
}

func TestForkMeshSendBlockBroadcast_ThreeForks(t *testing.T) {
	var mesh = NewForkMesh()
	var forkFrom = newTestFork(mesh)
	var forkTo1 = newTestFork(mesh)
	var forkTo2 = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlockBroadcast(forkFrom, block)
	var block1 blockgen.Block
	var block2 blockgen.Block
	go func() {
		block1 = <-mesh.ReceiveChan(forkTo1)
	}()
	go func() {
		block2 = <-mesh.ReceiveChan(forkTo2)
	}()
	time.Sleep(time.Second)
	if !block.Equal(block1) {
		t.Fatalf("block was not sent to first fork")
	}
	if !block.Equal(block2) {
		t.Fatalf("block was not sent to second fork")
	}
}

func TestForkMeshConnection_EarlyReceive(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	mesh.Disconnect(fork)
	defer func() { _ = recover() }()
	mesh.ReceiveChan(fork)
	t.Fatalf("fork got receive channel without connecting to mesh")
}

func TestForkMeshAllExistingBlocks_ThreeForks(t *testing.T) {
	var mesh = NewForkMesh()
	var fork1 = newTestFork(mesh)
	var fork2 = newTestFork(mesh)
	var fork3 = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	if len(mesh.AllExistingBlocks(0)) != 0 {
		t.Fatalf("mesh found non existant blocks")
	}
	fork1.blocks = chain[0:2]
	var got = len(mesh.AllExistingBlocks(0))
	var want = len(fork1.Blocks(0))
	if got != want {
		t.Fatalf("got %d blocks; want %d blocks", got, want)
	}
	fork2.blocks = chain[0:4]
	got = len(mesh.AllExistingBlocks(0))
	want = len(fork2.Blocks(0))
	if got != want {
		t.Fatalf("got %d blocks; want %d blocks", got, want)
	}
	fork3.blocks = chain[0:3]
	got = len(mesh.AllExistingBlocks(0))
	if got != want {
		t.Fatalf("got %d blocks; want %d blocks", got, want)
	}
}

func TestForkMeshSendBlock_DontSendInvalidBlock(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	block.Nonce += 1
	if mesh.SendBlockBroadcast(fork, block) {
		t.Fatalf("sent invalid block")
	}
}

func TestForkMeshAllExistingBlocks_IgnoreInvalidChains(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	chain[2].Nonce += 1
	fork.blocks = chain
	if len(mesh.AllExistingBlocks(0)) != 0 {
		t.Fatalf("mesh accepted invalid chain")
	}
}

func TestForkMeshAllExistingBlocks_CheckIndex(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	fork.blocks = chain
	var from = blockgen.Index(2)
	if mesh.AllExistingBlocks(from)[0].Index != from {
		t.Fatalf("mesh accepted chain with different index")
	}
}

func TestForkMeshAllExistingBlocks_SameIndex(t *testing.T) {
	var mesh = NewForkMesh()
	var fork1 = newTestFork(mesh)
	var fork2 = newTestFork(mesh)
	var fork3 = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var nextMinor = blockgen.GenerateNextFrom(chain[len(chain)-1], blockgen.Data{1}, nil)
	var chainMinor = []blockgen.Block(chain[:])
	chainMinor = append(chainMinor, nextMinor)
	fork1.blocks = chainMinor
	var nextMajor = blockgen.GenerateNextFrom(chain[len(chain)-1], blockgen.Data{2}, nil)
	var chainMajor = []blockgen.Block(chain[:])
	chainMajor = append(chainMajor, nextMajor)
	fork2.blocks = chainMajor
	fork3.blocks = fork2.blocks
	var got = mesh.AllExistingBlocks(0)
	if !got[len(got)-1].Equal(nextMajor) {
		t.Fatalf("mesh did not prefer major chain over minor")
	}
}

func TestFrokMeshSendBlockTo(t *testing.T) {
	var mesh = NewForkMesh()
	var forkFrom = newTestFork(mesh)
	var forkTo = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlockTo(forkTo, block)
	var blockTo blockgen.Block
	var blockFrom blockgen.Block
	go func() {
		blockTo = <-mesh.ReceiveChan(forkTo)
	}()
	go func() {
		blockFrom = <-mesh.ReceiveChan(forkFrom)
	}()
	time.Sleep(time.Second)
	if !block.Equal(blockTo) {
		t.Fatalf("block was not sent to first fork")
	}
	if block.Equal(blockFrom) {
		t.Fatalf("block was sent to second fork")
	}
}
