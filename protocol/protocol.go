package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"slava0135/blockchan/node"
)

type RemoteFork struct {
	Link   Link
	mesh   node.Mesh
	mentor node.Fork
}

type Link interface {
	SendChannel() chan []byte
	RecvChannel() chan []byte
}

func NewRemoteFork(mesh *mesh.ForkMesh, link Link, mentor node.Fork) *RemoteFork {
	var f = &RemoteFork{}
	f.Link = link
	f.mesh = mesh
	f.mentor = mentor
	return f
}

func (f *RemoteFork) Blocks(index blockgen.Index) []blockgen.Block {
	go func() {
		f.Link.SendChannel() <- messages.PackMessage(messages.AskForBlocksMsg{Index: uint64(index)})
	}()
	var chain = make(map[blockgen.Index]blockgen.Block)
	var expectedLen uint64 = 0
	for msg := range f.Link.RecvChannel() {
		var got = messages.UnpackMessage(msg)
		var b, ok = got.(messages.SendBlockMsg)
		if ok && b.Block.Index >= index {
			chain[b.Block.Index] = b.Block
			var newLen = b.LastBlockIndex - uint64(index) + 1
			if newLen > expectedLen {
				expectedLen = newLen
			}
			if len(chain) >= int(expectedLen) {
				break
			}
		}
	}
	var sortedChain = make([]blockgen.Block, len(chain))
	for _, b := range chain {
		sortedChain[b.Index-index] = b
	}
	return sortedChain
}

func (f *RemoteFork) sendBlock(b blockgen.Block, lastBlockIndex blockgen.Index) {
	f.Link.SendChannel() <- messages.PackMessage(messages.SendBlockMsg{Block: b, LastBlockIndex: uint64(lastBlockIndex)})
}

func (f *RemoteFork) Listen(shutdown chan struct{}) {
	f.mesh.Connect(f)
	defer f.mesh.Disconnect(f)
	for {
		select {
		case <-shutdown:
			return
		case b := <-f.mesh.ReceiveChan(f):
			go f.sendBlock(b, b.Index)
		case msg := <-f.Link.RecvChannel():
			var i = messages.UnpackMessage(msg)
			switch v := i.(type) {
			case messages.SendBlockMsg:
				f.mesh.SendBlockBroadcast(f, v.Block)
			case messages.AskForBlocksMsg:
				var blocks = f.mentor.Blocks(blockgen.Index(v.Index))
				var lastIndex = blocks[len(blocks)-1].Index
				for _, b := range blocks {
					f.sendBlock(b, lastIndex)
				}
			}
		}
	}
}
