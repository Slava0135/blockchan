package tests

import (
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/node"
	"slava0135/blockchan/validate"
	"testing"
	"time"
)

func runNode(node *node.Node, stop *bool) {
	node.Enable()
	for !*stop {
		node.ProcessNextBlock()
	}
	node.Disable()
}

func TestMeshAndTwoNodes(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var node1 = node.NewNode(mesh)
	var node2 = node.NewNode(mesh)
	var stop = false
	go runNode(node1, &stop)
	go runNode(node2, &stop)
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
	go runNode(node1, &stop)
	go runNode(node2, &stop)
	go runNode(node3, &stop)
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
