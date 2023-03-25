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
}

type Remote struct {
	Address string
}

func (l *NetworkLink) SendChannel() chan []byte {
	return l.sendChannel
}

func (l *NetworkLink) RecvChannel() chan []byte {
	return l.recvChannel
}

func newNetworkLink() *NetworkLink {
	var l = NetworkLink{}
	l.sendChannel = make(chan []byte)
	l.recvChannel = make(chan []byte)
	return &l
}

func Launch(address string, remotes []Remote) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	var mesh = mesh.NewForkMesh()
	var node = node.NewNode(mesh)
	mesh.Mentor = node
	var link = newNetworkLink()
	var fork = protocol.NewRemoteFork(mesh, link)
	go runRemoteListener(conn, fork)
	for _, v := range remotes {
		var link = newNetworkLink()
		var fork = protocol.NewRemoteFork(mesh, link)
		go runRemoteSender(conn, v, fork)
	}
	go runNode(node)
}

func runNode(node *node.Node) {
	node.Enable()
	for {
		node.ProcessNextBlock(blockgen.Data{})
	}
}

func runRemoteSender(conn *net.UDPConn, remote Remote, fork *protocol.RemoteFork) {
	addr, err := net.ResolveUDPAddr("udp", remote.Address)
	if err != nil {
		panic(err)
	}
	go func() {
		for msg := range fork.Link.SendChannel() {
			conn.WriteToUDP(msg, addr)
		}
	}()
	for {
		fork.Listen(nil)
	}
}

func runRemoteListener(conn *net.UDPConn, fork *protocol.RemoteFork) {
	go func() {
		for range fork.Link.SendChannel() {
		}
	}()
	go func() {
		var buf [1024]byte
		for {
			rlen, _, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				continue
			}
			fork.Link.RecvChannel() <- buf[:rlen]
		}
	}()
	for {
		fork.Listen(nil)
	}
}
