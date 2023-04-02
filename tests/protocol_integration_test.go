package tests

import (
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/node"
	"slava0135/blockchan/protocol"
	"slava0135/blockchan/validate"
	"sort"
	"testing"
	"time"
)

func TestTwoMeshes(t *testing.T) {
	var mesh1 = mesh.NewForkMesh()
	var node1 = node.NewNode(mesh1)
	node1.Name = "FIRST"
	var link1 = protocol.NewLink()
	var remote1 = protocol.NewRemoteFork(mesh1, link1, node1, 100 * time.Millisecond)
	go remote1.Listen(nil)

	var mesh2 = mesh.NewForkMesh()
	var node2 = node.NewNode(mesh2)
	node2.Name = "SECOND"
	var link2 = protocol.Link{}
	link2.RecvChan = link1.SendChan
	link2.SendChan = link1.RecvChan
	var remote2 = protocol.NewRemoteFork(mesh2, link2, node2, 100 * time.Millisecond)
	go remote2.Listen(nil)

	var stop = false
	go runNode(node1, nodeData(0x11), &stop, true)
	go runNode(node2, nodeData(0x22), &stop, false)
	time.Sleep(time.Second)
	stop = true

	if node1.Verified == 0 {
		t.Errorf("node %s did not verified any blocks", node1.Name)
	}
	if node2.Verified == 0 {
		t.Errorf("node %s did not verified any blocks", node2.Name)
	}
	var verified = []int{int(node1.Verified), int(node2.Verified)}
	sort.Ints(verified)
	var last = verified[0]
	if !validate.AreEqualChains(node1.Blocks(0)[:last], node2.Blocks(0)[:last]) {
		t.Fatalf("chains diverged")
	}
	var chain = node1.Blocks(0)
	if !validate.IsValidChain(chain) {
		t.Errorf("chain was not valid")
	}
}
