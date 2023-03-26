package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"slava0135/blockchan/node"
)

type RemoteFork struct {
	Link      Link
	mesh      node.Mesh
	mentor    node.Fork
	blocksReq chan blockgen.Index
	blocksAns chan []blockgen.Block
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
	f.blocksReq = make(chan blockgen.Index)
	f.blocksAns = make(chan []blockgen.Block)
	return f
}

func (f *RemoteFork) Blocks(index blockgen.Index) []blockgen.Block {
	f.blocksReq <- index
	chain := <- f.blocksAns
	return chain
}

func (f *RemoteFork) Listen(shutdown chan struct{}) {
	f.mesh.Connect(f)
	defer f.mesh.Disconnect(f)
	for {
		select {
		case <-shutdown:
			return
		case b := <-f.mesh.ReceiveChan(f):
			var chain = f.mentor.Blocks(0)
			var lastIndex = chain[len(chain)-1].Index
			f.Link.SendChannel() <- messages.PackMessage(messages.SendBlockMsg{Block: b, LastBlockIndex: uint64(lastIndex)})
		case msg := <-f.Link.RecvChannel():
			var i = messages.UnpackMessage(msg)
			switch v := i.(type) {
			case messages.SendBlockMsg:
				f.mesh.SendBlockTo(f.mentor, v.Block)
			case messages.AskForBlocksMsg:
				var chain = f.mentor.Blocks(blockgen.Index(v.Index))
				var lastIndex = chain[len(chain)-1].Index
				for _, b := range chain {
					f.Link.SendChannel() <- messages.PackMessage(messages.SendBlockMsg{Block: b, LastBlockIndex: uint64(lastIndex)})
				}
			}
		case index := <-f.blocksReq:
			f.Link.SendChannel() <- messages.PackMessage(messages.AskForBlocksMsg{Index: uint64(index)})
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
			f.blocksAns <- sortedChain
		}
	}
}
