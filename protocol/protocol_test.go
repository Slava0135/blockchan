package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"slava0135/blockchan/validate"
	"testing"
	"time"
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

type testFork struct {
	blocks []blockgen.Block
}

func (f *testFork) Blocks(from blockgen.Index) []blockgen.Block {
	return f.blocks[from:]
}

func (l *testLink) SendChannel() chan []byte {
	return l.sendChan
}

func (l *testLink) RecvChannel() chan []byte {
	return l.recvChan
}

func TestListen_SendBlock(t *testing.T) {
	var link = newTestLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil)
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	go remote.Listen(nil)
	time.Sleep(time.Second)
	go mesh.SendBlockBroadcast(nil, block)
	var unpacked = messages.UnpackMessage(<-link.sendChan)
	var received, ok = unpacked.(messages.SendBlockMsg)
	if !ok {
		t.Fatalf("wrong message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("got corrupted block through link")
	}
}

func TestBlocks(t *testing.T) {
	var link = newTestLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[len(chain)-1].Index
	go func() {
		for i := range chain {
			link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[i], LastBlockIndex: uint64(lastIndex)})
		}
	}()
	var got = remote.Blocks(0)
	if !validate.AreEqualChains(chain, got) {
		t.Fatalf("failed to get blocks from remote")
	}
}

func TestBlocks_Unsorted(t *testing.T) {
	var link = newTestLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[3].Index
	go func() {
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[3], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[2], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[0], LastBlockIndex: uint64(lastIndex)})
	}()
	var got = remote.Blocks(0)
	if !validate.AreEqualChains(chain, got) {
		t.Fatalf("failed to get blocks from remote; want = %d; got = %d", len(chain), len(got))
	}
}

func TestBlocks_DoubleSend(t *testing.T) {
	var link = newTestLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[3].Index
	go func() {
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[3], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[2], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[0], LastBlockIndex: uint64(lastIndex)})
	}()
	var got = remote.Blocks(0)
	if !validate.AreEqualChains(chain, got) {
		t.Fatalf("failed to get blocks from remote; want = %d; got = %d", len(chain), len(got))
	}
}

func TestBlocks_OldBlock(t *testing.T) {
	var link = newTestLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[3].Index
	go func() {
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[3], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[0], LastBlockIndex: uint64(lastIndex)})
		link.recvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[2], LastBlockIndex: uint64(lastIndex)})
	}()
	var got = remote.Blocks(1)
	if !validate.AreEqualChains(chain[1:], got) {
		t.Fatalf("failed to get blocks from remote; want = %d; got = %d", len(chain), len(got))
	}
}

func TestListen_AskedForBlocks(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	mentor.blocks = chain
	var linkSender = newTestLink()
	var linkReceiver = &testLink{}
	linkReceiver.recvChan = linkSender.sendChan
	linkReceiver.sendChan = linkSender.recvChan
	var remoteSender = NewRemoteFork(mesh, linkSender, mentor)
	var remoteReceiver = NewRemoteFork(mesh, linkReceiver, mentor)
	go remoteSender.Listen(nil)
	go remoteReceiver.Listen(nil)
	var got = remoteReceiver.Blocks(0)
	if !validate.AreEqualChains(got, chain) {
		t.Fatalf("failed to send chain for request")
	}
}
