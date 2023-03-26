package tests

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/node"
	"slava0135/blockchan/validate"
	"testing"
	"time"
)

func runNode(node *node.Node, data blockgen.Data, stop *bool) {
	node.Enable()
	for !*stop {
		node.ProcessNextBlock(data)
	}
	node.Disable()
}

func nodeData(n byte) blockgen.Data {
	var d blockgen.Data
	for i := 0; i < len(d); i += 1 {
		d[i] = n
	}
	return d
}

func TestMeshAndTwoNodes(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var node1 = node.NewNode(mesh)
	var node2 = node.NewNode(mesh)
	var stop = false
	go runNode(node1, nodeData(0x11), &stop)
	go runNode(node2, nodeData(0x22), &stop)
	time.Sleep(time.Second)
	stop = true
	if !validate.AreEqualChains(node1.Blocks(0), node2.Blocks(0)) {
		t.Fatalf("chains diverged")
	}
	var chain = node1.Blocks(0)
	if !validate.IsValidChain(chain) {
		t.Fatalf("chain was not valid")
	}
}

func TestMeshAndThreeNodes(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var node1 = node.NewNode(mesh)
	var node2 = node.NewNode(mesh)
	var node3 = node.NewNode(mesh)
	var stop = false
	go runNode(node1, nodeData(0x11), &stop)
	go runNode(node2, nodeData(0x22), &stop)
	go runNode(node3, nodeData(0x33), &stop)
	time.Sleep(time.Second)
	stop = true
	if !validate.AreEqualChains(node1.Blocks(0), node2.Blocks(0)) || !validate.AreEqualChains(node2.Blocks(0), node3.Blocks(0)) {
		t.Fatalf("chains diverged")
	}
	var chain = node1.Blocks(0)
	if !validate.IsValidChain(chain) {
		t.Fatalf("chain was not valid")
	}
}
