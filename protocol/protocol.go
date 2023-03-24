package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"slava0135/blockchan/node"
)

type RemoteFork struct {
	link Link
	mesh node.Mesh
}

type Link interface {
	SendChannel() chan []byte
	RecvChannel() chan []byte
}

func NewRemoteFork(mesh *mesh.ForkMesh, link Link) *RemoteFork {
	var f = &RemoteFork{}
	f.link = link
	f.mesh = mesh
	return f
}

func (f *RemoteFork) Blocks(index blockgen.Index) []blockgen.Block {
	go func() {
		f.link.SendChannel() <- messages.PackMessage(messages.AskForBlocksMsg{Index: uint64(index)})
	}()
	var chain []blockgen.Block
	var expectedLen uint64 = 0
	for msg := range f.link.RecvChannel() {
		var got = messages.UnpackMessage(msg)
		var b, ok = got.(messages.SendBlockMsg)
		if ok {
			chain = append(chain, b.Block)
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

func (f *RemoteFork) sendBlock(b blockgen.Block) {
	f.link.SendChannel() <- messages.PackMessage(messages.SendBlockMsg{Block: b})
}

func (f *RemoteFork) Listen(shutdown chan struct{}) {
	f.mesh.Connect(f)
	defer f.mesh.Disconnect(f)
	for {
		select {
		case <-shutdown:
			return
		case b := <-f.mesh.ReceiveChan(f):
			go f.sendBlock(b)
		}
	}
}
