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
		t.Errorf("node %s did not verify any blocks", node1.Name)
	}
	if node2.Verified == 0 {
		t.Errorf("node %s did not verify any blocks", node2.Name)
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

func TestThreeMeshes(t *testing.T) {
	var mesh1 = mesh.NewForkMesh()
	var node1 = node.NewNode(mesh1)
	node1.Name = "FIRST"
	var link12 = protocol.NewLink()
	var remote12 = protocol.NewRemoteFork(mesh1, link12, node1, 100 * time.Millisecond)
	var link13 = protocol.NewLink()
	var remote13 = protocol.NewRemoteFork(mesh1, link13, node1, 100 * time.Millisecond)
	go remote12.Listen(nil)
	go remote13.Listen(nil)

	var mesh2 = mesh.NewForkMesh()
	var node2 = node.NewNode(mesh2)
	node2.Name = "SECOND"
	var link21 = protocol.Link{}
	link21.RecvChan = link12.SendChan
	link21.SendChan = link12.RecvChan
	var remote21 = protocol.NewRemoteFork(mesh2, link21, node2, 100 * time.Millisecond)
	var link23 = protocol.NewLink()
	var remote23 = protocol.NewRemoteFork(mesh2, link23, node2, 100 * time.Millisecond)
	go remote21.Listen(nil)
	go remote23.Listen(nil)

	var mesh3 = mesh.NewForkMesh()
	var node3 = node.NewNode(mesh3)
	node3.Name = "THIRD"
	var link31 = protocol.Link{}
	link31.RecvChan = link13.SendChan
	link31.SendChan = link13.RecvChan
	var remote31 = protocol.NewRemoteFork(mesh3, link31, node3, 100 * time.Millisecond)
	var link32 = protocol.Link{}
	link32.RecvChan = link23.SendChan
	link32.SendChan = link23.RecvChan
	var remote32 = protocol.NewRemoteFork(mesh3, link32, node3, 100 * time.Millisecond)
	go remote31.Listen(nil)
	go remote32.Listen(nil)

	var stop = false
	go runNode(node1, nodeData(0x11), &stop, true)
	go runNode(node2, nodeData(0x22), &stop, false)
	time.Sleep(time.Second)
	go runNode(node3, nodeData(0x33), &stop, false)
	time.Sleep(3 * time.Second)
	stop = true

	if node1.Verified == 0 {
		t.Errorf("node %s did not verify any blocks", node1.Name)
	}
	if node2.Verified == 0 {
		t.Errorf("node %s did not verify any blocks", node2.Name)
	}
	if node3.Verified == 0 {
		t.Errorf("node %s did not verify any blocks", node3.Name)
	}
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
