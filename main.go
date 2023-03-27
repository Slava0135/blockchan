package main

import (
	"flag"
	"slava0135/blockchan/network"

	log "github.com/sirupsen/logrus"
)

func main() {
    name := flag.String("name", "NONAME", "node name for logging")
    genesis := flag.Bool("genesis", false, "generate genesis block on start?")
    addr := flag.String("addr", "localhost:8888", "address node will be running")
    debug := flag.Bool("debug", false, "enable debug level logging")
    flag.Parse()
    if *debug {
        log.SetLevel(log.DebugLevel)
    }
    remotes := flag.Args()
    network.Launch(*name, *addr, remotes, *genesis)
}
