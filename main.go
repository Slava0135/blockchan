package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	name := flag.String("name", "NONAME", "node name for logging")
	genesis := flag.Bool("genesis", false, "generate genesis block on start?")
	addr := flag.String("addr", "localhost:8888", "address node will be running")
	debug := flag.Bool("debug", false, "enable debug level logging")
	timeout := flag.Duration("timeout", 100 * time.Millisecond, "timeout for missing blocks request")
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	remotes := flag.Args()
	Launch(*name, *addr, remotes, *genesis, *timeout)
}
