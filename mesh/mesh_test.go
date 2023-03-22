package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"testing"
	"time"
)

func TestNodeMesh_Interface(t *testing.T) {
	var _ node.Mesh = &ForkMesh{}
}

func TestNodeMesh_SendAndReceive(t *testing.T) {
	var mesh = NewNodeMesh()
	var nodeFrom = node.NewNode(mesh)
	mesh.Connect(nodeFrom)
	var nodeTo = node.NewNode(mesh)
	mesh.Connect(nodeTo)
	var sent = blockgen.GenerateGenesisBlock()
	go mesh.SendBlock(nodeFrom, sent)
	var received = <-mesh.ReceiveChan(nodeTo)
	if received != sent {
		t.Fatalf("block was not sent")
	}
}

func TestNodeMeshSendBlock_Loopback(t *testing.T) {
	var mesh = NewNodeMesh()
	var node = node.NewNode(mesh)
	mesh.Connect(node)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlock(node, block)
	select {
	case <-mesh.ReceiveChan(node):
		t.Fatalf("mesh tried to send block back to sender")
	default:
	}
}

func TestNodeMesh_ThreeNodes(t *testing.T) {
	var mesh = NewNodeMesh()
	var nodeFrom = node.NewNode(mesh)
	mesh.Connect(nodeFrom)
	var nodeTo1 = node.NewNode(mesh)
	mesh.Connect(nodeTo1)
	var nodeTo2 = node.NewNode(mesh)
	mesh.Connect(nodeTo2)
	var block = blockgen.GenerateGenesisBlock()
	go mesh.SendBlock(nodeFrom, block)
	var received = <-mesh.ReceiveChan(nodeTo1)
	if received != block {
		t.Fatalf("block was not sent to first node")
	}
	select {
	case <-mesh.ReceiveChan(nodeTo2):
	default:
		t.Fatalf("block was not sent to second node")
	}
}

func TestNodeMeshConnection_EarlyReceive(t *testing.T) {
	var mesh = NewNodeMesh()
	var node = &node.Node{}
	defer func() { _ = recover() }()
	mesh.ReceiveChan(node)
	t.Fatalf("node got receive channel without connecting to mesh")
}

func TestNodeMeshAllExistingBlocks(t *testing.T) {
	var mesh = NewNodeMesh()
	var node = node.NewNode(mesh)
	node.Start()
	time.Sleep(time.Second)
	node.Shutdown()
	if len(mesh.AllExistingBlocks()) != len(node.Blocks()) {
		t.Fatalf("mesh did not get existing blocks")
	}
}

func TestNodeMeshAllExistingBlocks_NoNodes(t *testing.T) {
	var mesh = NewNodeMesh()
	if len(mesh.AllExistingBlocks()) != 0 {
		t.Fatalf("mesh found non existant blocks")
	}
}
