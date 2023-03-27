package main

import (
	"slava0135/blockchan/network"
	"flag"
)

func main() {
    name := flag.String("name", "NONAME", "node name for logging")
    genesis := flag.Bool("genesis", false, "generate genesis block on start?")
    addr := flag.String("addr", "localhost:8888", "address node will be running")
    remotes := flag.Args()
    flag.Parse()
    network.Launch(*name, *addr, remotes, *genesis)
}
