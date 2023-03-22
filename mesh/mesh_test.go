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

func TestForkMesh_ThreeForks(t *testing.T) {
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

func TestForkMeshAllExistingBlocks(t *testing.T) {
	var mesh = NewForkMesh()
	var fork = newTestFork(mesh)
	if len(mesh.AllExistingBlocks()) != len(fork.Blocks()) {
		t.Fatalf("mesh did not get existing blocks")
	}
}

func TestForkMeshAllExistingBlocks_NoForks(t *testing.T) {
	var mesh = NewForkMesh()
	if len(mesh.AllExistingBlocks()) != 0 {
		t.Fatalf("mesh found non existant blocks")
	}
}
