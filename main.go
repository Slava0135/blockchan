package main

import (
	"slava0135/blockchan/network"
)

func main() {
    var remotes = []network.Remote{{Address: ":8888"}, {Address: ":8889"}, {Address: ":8890"}}
    go network.Launch(remotes[0].Address, []network.Remote{remotes[1], remotes[2]})
    go network.Launch(remotes[1].Address, []network.Remote{remotes[0], remotes[2]})
    network.Launch(remotes[2].Address, []network.Remote{remotes[0], remotes[1]})
}
