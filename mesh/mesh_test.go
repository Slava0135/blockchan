package mesh

import (
	"slava0135/blockchan/blockgen"
	"testing"
	"time"
)

type testFork struct {
	blocks []blockgen.Block
}

func (f *testFork) Blocks(from blockgen.Index) []blockgen.Block {
	return f.blocks[from:]
}

func newTestFork(mesh Mesh) *testFork {
	var fork = &testFork{}
	mesh.Connect(fork)
	return fork
}

func TestForkMesh_Interface(t *testing.T) {
	var _ Mesh = &ForkMesh{}
}

func TestForkMesh_SendAndReceive(t *testing.T) {
	var mesh = NewForkMesh()
	var forkFrom = newTestFork(mesh)
	var forkTo = newTestFork(mesh)
	var sent = blockgen.GenerateGenesisBlock()
	go mesh.SendBlockBroadcast(forkFrom, sent)
	var received = <-mesh.RecvChan(forkTo)
	if !sent.Equal(received.Block) {
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
	case <-mesh.RecvChan(fork):
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
		block1 = (<-mesh.RecvChan(forkTo1)).Block
	}()
	go func() {
		block2 = (<-mesh.RecvChan(forkTo2)).Block
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
	mesh.RecvChan(fork)
	t.Fatalf("fork got receive channel without connecting to mesh")
}

func TestForkMeshRequestBlocks_ThreeForks(t *testing.T) {
	var mesh = NewForkMesh()
	var fork1 = newTestFork(mesh)
	var fork2 = newTestFork(mesh)
	var fork3 = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	if len(mesh.RequestBlocks(0, nil)) != 0 {
		t.Fatalf("mesh found non existant blocks")
	}
	fork1.blocks = chain[0:2]
	var got = len(mesh.RequestBlocks(0, nil))
	var want = len(fork1.Blocks(0))
	if got != want {
		t.Fatalf("got %d blocks; want %d blocks", got, want)
	}
	fork2.blocks = chain[0:4]
	got = len(mesh.RequestBlocks(0, nil))
	want = len(fork2.Blocks(0))
	if got != want {
		t.Fatalf("got %d blocks; want %d blocks", got, want)
	}
	fork3.blocks = chain[0:3]
	got = len(mesh.RequestBlocks(0, nil))
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
	if mesh.SendBlockTo(fork, ForkBlock{Block: block}) {
		t.Fatalf("sent invalid block")
	}
}

func TestForkMeshRequestBlocks_IgnoreInvalidChains(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	chain[2].Nonce += 1
	fork.blocks = chain
	if len(mesh.RequestBlocks(0, nil)) != 0 {
		t.Fatalf("mesh accepted invalid chain")
	}
}

func TestForkMeshRequestBlocks_CheckIndex(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	fork.blocks = chain
	var from = blockgen.Index(2)
	if mesh.RequestBlocks(from, nil)[0].Index != from {
		t.Fatalf("mesh accepted chain with different index")
	}
}

func TestForkMeshRequestBlocks_NoLoopBack(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := 0; i < 3; i++ {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	fork.blocks = chain
	var from = blockgen.Index(2)
	if mesh.RequestBlocks(from, fork) != nil {
		t.Fatalf("mesh returned fork its own chain")
	}
}

func TestForkMeshSendBlockTo(t *testing.T) {
	var mesh = NewForkMesh()
	var forkFrom = newTestFork(mesh)
	var forkTo = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlockTo(forkTo, ForkBlock{Block: block})
	var blockTo blockgen.Block
	var blockFrom blockgen.Block
	go func() {
		blockTo = (<-mesh.RecvChan(forkTo)).Block
	}()
	go func() {
		blockFrom = (<-mesh.RecvChan(forkFrom)).Block
	}()
	time.Sleep(time.Second)
	if !block.Equal(blockTo) {
		t.Fatalf("block was not sent to wanted fork")
	}
	if block.Equal(blockFrom) {
		t.Fatalf("block was sent to original fork")
	}
}

func TestForkMeshDropUnverified(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	mesh.DropUnverifiedBlocks(fork, blockgen.Block{})
	if !(<-mesh.RecvChan(fork)).Drop {
		t.Fatalf("mesh did not ask fork to drop unverified blocks")
	}
}

func TestForkMeshConnect_Concurrent(t *testing.T) {
	var mesh = NewForkMesh()
	var forks []Fork
	for i := 0; i < 100; i += 1 {
		var fork = &testFork{}
		forks = append(forks, fork)
		go mesh.Connect(fork)
	}
	for _, v := range forks {
		go mesh.Disconnect(v)
	}
}
