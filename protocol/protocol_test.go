package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"slava0135/blockchan/validate"
	"testing"
	"time"
)

type testFork struct {
	blocks []blockgen.Block
}

func (f *testFork) Blocks(from blockgen.Index) []blockgen.Block {
	return f.blocks[from:]
}

func TestListen_SendBlock(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	var remote = NewRemoteFork(mesh, link, mentor, time.Second)
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	mentor.blocks = []blockgen.Block{block}
	go remote.Listen(nil)
	time.Sleep(time.Second)
	go mesh.SendBlockBroadcast(nil, block)
	var unpacked = messages.UnpackMessage(<-link.SendChan)
	var received, ok = unpacked.(messages.SendBlockMsg)
	if !ok {
		t.Fatalf("wrong message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("got corrupted block through link")
	}
}

func TestBlocks(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil, time.Second)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[len(chain)-1].Index
	go func() {
		<-link.SendChan
		for i := range chain {
			link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[i], LastBlockIndex: uint64(lastIndex)})
		}
	}()
	go remote.Listen(nil)
	var got = remote.Blocks(0)
	if !validate.AreEqualChains(chain, got) {
		t.Fatalf("failed to get blocks from remote")
	}
}

func TestBlocks_Unsorted(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil, time.Second)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[3].Index
	go func() {
		<-link.SendChan
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[3], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[2], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[0], LastBlockIndex: uint64(lastIndex)})
	}()
	go remote.Listen(nil)
	var got = remote.Blocks(0)
	if !validate.AreEqualChains(chain, got) {
		t.Fatalf("failed to get blocks from remote; want = %d; got = %d", len(chain), len(got))
	}
}

func TestBlocks_DoubleSend(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil, time.Second)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[3].Index
	go func() {
		<-link.SendChan
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[3], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[2], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[0], LastBlockIndex: uint64(lastIndex)})
	}()
	go remote.Listen(nil)
	var got = remote.Blocks(0)
	if !validate.AreEqualChains(chain, got) {
		t.Fatalf("failed to get blocks from remote; want = %d; got = %d", len(chain), len(got))
	}
}

func TestBlocks_OldBlock(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil, time.Second)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	var lastIndex = chain[3].Index
	go func() {
		<-link.SendChan
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[3], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[1], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[0], LastBlockIndex: uint64(lastIndex)})
		link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: chain[2], LastBlockIndex: uint64(lastIndex)})
	}()
	go remote.Listen(nil)
	var got = remote.Blocks(1)
	if !validate.AreEqualChains(chain[1:], got) {
		t.Fatalf("failed to get blocks from remote; want = %d; got = %d", len(chain), len(got))
	}
}

func TestListen_RequestedBlocks(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	mesh.Connect(mentor)
	var chain = []blockgen.Block{blockgen.GenerateGenesisBlock()}
	for i := byte(0); i < 3; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}, nil))
	}
	mentor.blocks = chain
	var linkSender = NewLink()
	var linkReceiver = Link{}
	linkReceiver.RecvChan = linkSender.SendChan
	linkReceiver.SendChan = linkSender.RecvChan
	var remoteSender = NewRemoteFork(mesh, linkSender, mentor, time.Second)
	var remoteReceiver = NewRemoteFork(mesh, linkReceiver, mentor, time.Second)
	go remoteSender.Listen(nil)
	go remoteReceiver.Listen(nil)
	var got = remoteReceiver.Blocks(0)
	if !validate.AreEqualChains(got, chain) {
		t.Fatalf("failed to send chain for request")
	}
}

func TestShutdown(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var shut = make(chan struct{})
	var remote = NewRemoteFork(mesh, link, nil, time.Second)
	go remote.Listen(shut)
	shut <- struct{}{}
}

func TestListen_SendBlockOnlyToMentor(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	mesh.Connect(mentor)
	var unwanted = &testFork{}
	mesh.Connect(unwanted)
	var link = NewLink()
	var remote = NewRemoteFork(mesh, link, mentor, time.Second)
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	go remote.Listen(nil)
	link.RecvChan <- messages.PackMessage(messages.SendBlockMsg{Block: block, LastBlockIndex: 0})
	var _ = <-mesh.RecvChan(mentor)
	var blockTo blockgen.Block
	go func() {
		blockTo = (<-mesh.RecvChan(unwanted)).Block
	}()
	time.Sleep(time.Second)
	if block.Equal(blockTo) {
		t.Fatalf("block was sent to unwanted fork")
	}
}

func TestListen_AskForBlocks_NoBlocks(t *testing.T) {
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	mesh.Connect(mentor)
	var link = NewLink()
	var remote = NewRemoteFork(mesh, link, mentor, time.Second)
	go remote.Listen(nil)
	link.RecvChan <- messages.PackMessage(messages.RequestBlocksMsg{Index: 0})
	time.Sleep(time.Second)
}

func TestBlocks_Timeout(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var remote = NewRemoteFork(mesh, link, nil, time.Second)
	go remote.Listen(nil)
	remote.Blocks(0)
}

func TestListen_AskDropBlock(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	var remote = NewRemoteFork(mesh, link, mentor, time.Second)
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	mentor.blocks = []blockgen.Block{block}
	go remote.Listen(nil)
	time.Sleep(time.Second)
	go mesh.DropUnverifiedBlocks(remote, block)
	var unpacked = messages.UnpackMessage(<-link.SendChan)
	var received, ok = unpacked.(messages.DropBlockMsg)
	if !ok {
		t.Fatalf("wrong message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("got corrupted block through link")
	}
}


func TestListen_DropBlock(t *testing.T) {
	var link = NewLink()
	var mesh = mesh.NewForkMesh()
	var mentor = &testFork{}
	mesh.Connect(mentor)
	var remote = NewRemoteFork(mesh, link, mentor, time.Second)
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	mentor.blocks = []blockgen.Block{block}
	go remote.Listen(nil)
	link.RecvChan <- messages.PackMessage(messages.DropBlockMsg{Block: block, LastBlockIndex: 0})
	var drop bool
	go func() {
		drop = (<-mesh.RecvChan(mentor)).Drop
	}()
	time.Sleep(time.Second)
	if !drop {
		t.Fatalf("mentor was not asked to drop blocks")
	}
}