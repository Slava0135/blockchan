package main

import (
	"slava0135/blockchan/network"
	"time"
)

func main() {
    var remotes = []network.Remote{{Address: "localhost:8888"}, {Address: "localhost:8889"}, {Address: "localhost:8890"}}
    go network.Launch(0x11, remotes[0].Address, []network.Remote{remotes[1], remotes[2]})
    time.Sleep(10 * time.Millisecond)
    go network.Launch(0x22, remotes[1].Address, []network.Remote{remotes[0], remotes[2]})
    time.Sleep(10 * time.Millisecond)
    go network.Launch(0x33, remotes[2].Address, []network.Remote{remotes[0], remotes[1]})
    time.Sleep(time.Second)
}
