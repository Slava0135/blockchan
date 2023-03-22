package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"testing"
	"time"
)

func TestNodeMesh_Interface(t *testing.T) {
	var _ node.Mesh = &NodeMesh{}
}

func TestNodeMesh_SendAndReceive(t *testing.T) {
	var mesh = NewNodeMesh()
	var nodeFrom = node.NewNode(mesh)
	var nodeTo = node.NewNode(mesh)
	var sent = blockgen.GenerateGenesisBlock()
	go mesh.SendBlock(nodeFrom, sent)
	var received = <-mesh.ReceiveChan(nodeTo)
	if received != sent {
		t.Fatalf("block was not sent")
	}
}

func TestNodeMesh_SendLoopback(t *testing.T) {
	var mesh = NewNodeMesh()
	var node = node.NewNode(mesh)
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
	var nodeTo1 = node.NewNode(mesh)
	var nodeTo2 = node.NewNode(mesh)
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

func TestNodeMesh_ConnectFirst(t *testing.T) {
	var mesh = NewNodeMesh()
	var node = &node.Node{}
	defer func() { _ = recover() }()
	mesh.ReceiveChan(node)
	t.Fatalf("node got receive channel without connecting to mesh")
}

func TestNodeMesh_AllExistingBlocks(t *testing.T) {
	var mesh = NewNodeMesh()
	var node = node.NewNode(mesh)
	node.Start()
	time.Sleep(time.Second)
	node.Shutdown()
	if len(mesh.AllExistingBlocks()) != len(node.Blocks) {
		t.Fatalf("mesh did not get existing blocks")
	}
}
