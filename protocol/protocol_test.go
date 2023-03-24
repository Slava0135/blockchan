package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"testing"
)

type testLink struct {
	sendChan chan []byte
	recvChan chan []byte
}

func newTestLink() *testLink {
	var link = &testLink{}
	link.sendChan = make(chan []byte)
	link.recvChan = make(chan []byte)
	return link
}

func (l *testLink) SendChannel() chan []byte {
	return l.sendChan
}

func (l *testLink) RecvChannel() chan []byte {
	return l.recvChan
}

func TestSendBlock(t *testing.T) {
	var link = newTestLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link)
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	remote.SendBlock(block)
	var unpacked = messages.UnpackMessage(<-link.sendChan)
	var received, ok = unpacked.(messages.SendBlockMsg)
	if !ok {
		t.Fatalf("wrong message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("got corrupted block through link")
	}
}
