package main

import (
	"slava0135/blockchan/network"
)

func main() {
    var self = network.Remote{Address: ":8888"}
    network.Launch(":8888", []network.Remote{self})
}
