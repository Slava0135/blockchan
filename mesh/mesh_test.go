package mesh

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/node"
	"testing"
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
