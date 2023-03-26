package network

import (
	log "github.com/sirupsen/logrus"
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

func Launch(name string, address string, remotes []Remote, genesis bool) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	log.Info("socket ", addr, " initialised")
	var mesh = mesh.NewForkMesh()
	var node = node.NewNode(mesh)
	node.Name = name
	var forks = make(map[string]*protocol.RemoteFork)
	for _, v := range remotes {
		addr, err := net.ResolveUDPAddr("udp", v.Address)
		if err != nil {
			log.Panic(err)
		}
		var link = newNetworkLink()
		var fork = protocol.NewRemoteFork(mesh, link, node)
		forks[addr.String()] = fork
		log.Info("starting sender to ", v.Address)
		go runRemoteSender(conn, addr, fork)
	}
	go runNode(node, name, genesis)
	log.Infof("starting node %s on %s", node.Name, addr)
	for {
		var buf [1024]byte
		for {
			rlen, rem, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				continue
			}
			if f, ok := forks[rem.String()]; ok {
				log.Infof("%s received message from %s of length %d bytes", conn.LocalAddr(), rem, rlen)
				var msg = make([]byte, rlen)
				copy(msg, buf[:rlen])
				f.Link.RecvChannel() <- msg
			}
		}
	}
}

func runNode(node *node.Node, data string, genesis bool) {
	node.Enable(genesis)
	var d blockgen.Data
	copy(d[:], []byte(data))
	for {
		node.ProcessNextBlock(d)
	}
}

func runRemoteSender(conn *net.UDPConn, addr *net.UDPAddr, fork *protocol.RemoteFork) {
	go func() {
		for msg := range fork.Link.SendChannel() {
			log.Infof("%s sending message to %s of length %d bytes", conn.LocalAddr(), addr, len(msg))
			log.Debug(string(msg))
			conn.WriteToUDP(msg, addr)
		}
	}()
	for {
		fork.Listen(nil)
	}
}
