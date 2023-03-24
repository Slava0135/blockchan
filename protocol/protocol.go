package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
)

type RemoteFork struct {
	link Link
}

type Link interface {
	SendChannel() chan []byte
	RecvChannel() chan []byte
}

func NewRemoteFork(mesh *mesh.ForkMesh, link Link) *RemoteFork {
	var f = &RemoteFork{}
	f.link = link
	return f
}

func (f *RemoteFork) SendBlock(b blockgen.Block) {
	go func() {
		f.link.SendChannel() <- messages.PackMessage(messages.SendBlockMsg{Block: b})
	}()
}

func (f *RemoteFork) Blocks(index blockgen.Index) []blockgen.Block {
	go func() {
		f.link.SendChannel() <- messages.PackMessage(messages.AskForBlocksMsg{Index: uint64(index)})
	}()
	var chain []blockgen.Block
	for msg := range f.link.RecvChannel() {
		var got = messages.UnpackMessage(msg)
		var b, ok = got.(messages.SendBlockMsg)
		if ok {
			chain = append(chain, b.Block)
			if b.LastBlockIndex == uint64(b.Block.Index) {
				break
			}
		}
	}
	return chain
}
