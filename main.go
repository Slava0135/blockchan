package main

import (
	"slava0135/blockchan/network"
	"time"
)

func main() {
    var remotes = []network.Remote{{Address: "localhost:8888"}, {Address: "localhost:8889"}, {Address: "localhost:8890"}}
    go network.Launch("ONE", remotes[0].Address, []network.Remote{remotes[1], remotes[2]}, true)
    time.Sleep(50 * time.Millisecond)
    go network.Launch("TWO", remotes[1].Address, []network.Remote{remotes[0], remotes[2]}, false)
    time.Sleep(50 * time.Millisecond)
    go network.Launch("THREE", remotes[2].Address, []network.Remote{remotes[0], remotes[1]}, false)
    time.Sleep(5000 * time.Millisecond)
}
