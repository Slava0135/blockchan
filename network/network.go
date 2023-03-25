package network

import (
	"net"
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/node"
	"slava0135/blockchan/protocol"
)

type NetworkLink struct {
	sendChannel chan []byte
	recvChannel chan []byte
	remote Remote
}

type Remote net.TCPAddr

func (l *NetworkLink) SendChannel() chan []byte {
	return l.sendChannel
}

func (l *NetworkLink) RecvChannel() chan []byte {
	return l.recvChannel
}

func NewNetworkLink(remote Remote) *NetworkLink {
	var l = NetworkLink{}
	l.sendChannel = make(chan []byte)
	l.recvChannel = make(chan []byte)
	l.remote = remote
	return &l
}

func Launch(remotes []Remote) {
	var mesh = mesh.NewForkMesh()
	var node = node.NewNode(mesh)
	mesh.Mentor = node
	for _, v := range remotes {
		var link = NewNetworkLink(v)
		var fork = protocol.NewRemoteFork(mesh, link)
		go runRemote(fork)
	}
	go runNode(node)
}

func runNode(node *node.Node) {
	node.Enable()
	for {
		node.ProcessNextBlock(blockgen.Data{})
	}
}

func runRemote(remote *protocol.RemoteFork) {
	for {
		remote.Listen(nil)
	}
}
