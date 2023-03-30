package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/mesh"
	"slava0135/blockchan/node"
	"slava0135/blockchan/protocol"
	"time"

	log "github.com/sirupsen/logrus"
)

func Launch(name string, address string, remotes []string, genesis bool, timeout time.Duration) {
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
		addr, err := net.ResolveUDPAddr("udp", v)
		if err != nil {
			log.Panic(err)
		}
		var link = protocol.NewLink()
		var fork = protocol.NewRemoteFork(mesh, link, node, timeout)
		forks[addr.String()] = fork
		log.Info("starting sender to ", v)
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
				log.Debugf("%s received message from %s of length %d bytes", conn.LocalAddr(), rem, rlen)
				var msg = make([]byte, rlen)
				copy(msg, buf[:rlen])
				f.Link.RecvChan <- msg
			}
		}
	}
}

func runNode(node *node.Node, data string, genesis bool) {
	filename := fmt.Sprintf("/blockchan/%s.txt", node.Name)
	os.Remove(filename)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err == nil {
		defer f.Close()
	} else {
		log.Errorf("node %s failed to open file %s: %v", node.Name, filename, err)
	}

	var seed int64
	for _, v := range []byte(data) {
		seed += int64(v)
	}
	var gen = rand.New(rand.NewSource(seed))
	node.Enable(genesis)
	var nextVerified = 0
	for {
		var d blockgen.Data
		gen.Read(d[:])
		node.ProcessNextBlock(d)
		if err != nil {
			var blocks = node.Blocks(0)
			for ; nextVerified <= int(node.Verified); nextVerified += 1 {
				f.WriteString(blocks[nextVerified].String())
			}
		}
	}
}

func runRemoteSender(conn *net.UDPConn, addr *net.UDPAddr, fork *protocol.RemoteFork) {
	go func() {
		for msg := range fork.Link.SendChan {
			log.Debugf("%s sending message to %s of length %d bytes", conn.LocalAddr(), addr, len(msg))
			log.Debug(string(msg))
			conn.WriteToUDP(msg, addr)
		}
	}()
	for {
		fork.Listen(nil)
	}
}
