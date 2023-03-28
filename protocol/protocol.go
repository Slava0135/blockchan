package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/messages"
	"time"
)

type RemoteFork struct {
	Link      Link
	mesh      mesh.Mesh
	mentor    mesh.Fork
	blocksReq chan blockgen.Index
	blocksAns chan []blockgen.Block
}

type Link struct {
	SendChan chan []byte
	RecvChan chan []byte
}

func NewLink() Link {
	var link = Link{}
	link.SendChan = make(chan []byte)
	link.RecvChan = make(chan []byte)
	return link
}

func NewRemoteFork(mesh *mesh.ForkMesh, link Link, mentor mesh.Fork) *RemoteFork {
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
	chain := <-f.blocksAns
	return chain
}

func (f *RemoteFork) Listen(shutdown chan struct{}) {
	f.mesh.Connect(f)
	defer f.mesh.Disconnect(f)
	for {
		select {
		case <-shutdown:
			return
		case b := <-f.mesh.RecvChan(f):
			var chain = f.mentor.Blocks(0)
			var lastIndex = chain[len(chain)-1].Index
			f.Link.SendChan <- messages.PackMessage(messages.SendBlockMsg{Block: b.Block, LastBlockIndex: uint64(lastIndex)})
		case msg := <-f.Link.RecvChan:
			var i = messages.UnpackMessage(msg)
			switch v := i.(type) {
			case messages.SendBlockMsg:
				f.mesh.SendBlockTo(f.mentor, mesh.ForkBlock{Block: v.Block, From: f})
			case messages.RequestBlocksMsg:
				var chain = f.mentor.Blocks(blockgen.Index(v.Index))
				if len(chain) == 0 {
					continue
				}
				var lastIndex = chain[len(chain)-1].Index
				for _, b := range chain {
					var b = b
					go func() {
						f.Link.SendChan <- messages.PackMessage(messages.SendBlockMsg{Block: b, LastBlockIndex: uint64(lastIndex)})
					}()
				}
			}
		case index := <-f.blocksReq:
			go func() {
				f.Link.SendChan <- messages.PackMessage(messages.RequestBlocksMsg{Index: uint64(index)})
			}()
			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(10 * time.Millisecond)
				timeout <- true
			}()
			var chain = make(map[blockgen.Index]blockgen.Block)
			var expectedLen uint64 = 0
			for {
				select {
				case msg := <-f.Link.RecvChan:
					var got = messages.UnpackMessage(msg)
					var b, ok = got.(messages.SendBlockMsg)
					if ok && b.Block.Index >= index {
						chain[b.Block.Index] = b.Block
						var newLen = b.LastBlockIndex - uint64(index) + 1
						if newLen > expectedLen {
							expectedLen = newLen
						}
						if len(chain) >= int(expectedLen) {
							goto ret
						}
					}
				case <-timeout:
					goto ret
				}
			}
		ret:
			var sortedChain = make([]blockgen.Block, len(chain))
			for _, b := range chain {
				var i = b.Index - index
				if int(i) < len(chain) {
					sortedChain[b.Index-index] = b
				}
			}
			f.blocksAns <- sortedChain
		}
	}
}
