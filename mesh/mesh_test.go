package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"testing"
)

type testFork struct {
	blocks []blockgen.Block
}

func (f *testFork) Blocks() []blockgen.Block {
	return f.blocks
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
	go mesh.SendBlock(forkFrom, sent)
	var received = <-mesh.ReceiveChan(forkTo)
	if received != sent {
		t.Fatalf("block was not sent")
	}
}

func TestForkMeshSendBlock_Loopback(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlock(fork, block)
	select {
	case <-mesh.ReceiveChan(fork):
		t.Fatalf("mesh tried to send block back to sender")
	default:
	}
}

func TestForkMeshSendBlock_ThreeForks(t *testing.T) {
	var mesh = NewForkMesh()
	var forkFrom = newTestFork(mesh)
	var forkTo1 = newTestFork(mesh)
	var forkTo2 = newTestFork(mesh)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlock(forkFrom, block)
	var received = <-mesh.ReceiveChan(forkTo1)
	if received != block {
		t.Fatalf("block was not sent to first fork")
	}
	select {
	case <-mesh.ReceiveChan(forkTo2):
	default:
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
	var want = len(fork1.Blocks())
	if got != want {
		t.Fatalf("got %d blocks; want %d blocks", got, want)
	}
	fork2.blocks = chain[0:4]
	got = len(mesh.AllExistingBlocks(0))
	want = len(fork2.Blocks())
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
	if mesh.SendBlock(fork, block) {
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
