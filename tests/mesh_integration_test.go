package tests

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/node"
	"slava0135/blockchan/validate"
	"sort"
	"testing"
	"time"
)

func runNode(node *node.Node, data blockgen.Data, stop *bool, genesis bool) {
	node.Enable(genesis)
	for !*stop {
		node.ProcessNextBlock(data)
	}
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
	node1.Name = "FIRST"
	var node2 = node.NewNode(mesh)
	node2.Name = "SECOND"
	var stop = false
	go runNode(node1, nodeData(0x11), &stop, true)
	go runNode(node2, nodeData(0x22), &stop, false)
	time.Sleep(time.Second)
	stop = true
	var verified = []int{int(node1.Verified), int(node2.Verified)}
	sort.Ints(verified)
	var last = verified[0]
	if !validate.AreEqualChains(node1.Blocks(0)[:last], node2.Blocks(0)[:last]) {
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
	node1.Name = "FIRST"
	var node2 = node.NewNode(mesh)
	node2.Name = "SECOND"
	var node3 = node.NewNode(mesh)
	node3.Name = "THIRD"
	var stop = false
	go runNode(node1, nodeData(0x11), &stop, true)
	go runNode(node2, nodeData(0x22), &stop, false)
	time.Sleep(time.Second)
	go runNode(node3, nodeData(0x33), &stop, false)
	time.Sleep(3 * time.Second)
	stop = true
	var verified = []int{int(node1.Verified), int(node2.Verified), int(node3.Verified)}
	sort.Ints(verified)
	var last = verified[0]
	if !validate.AreEqualChains(node1.Blocks(0)[:last], node2.Blocks(0)[:last]) || !validate.AreEqualChains(node2.Blocks(0)[:last], node3.Blocks(0)[:last]) {
		t.Fatalf("chains diverged")
	}
	var chain = node1.Blocks(0)
	if !validate.IsValidChain(chain) {
		t.Fatalf("chain was not valid")
	}
}
