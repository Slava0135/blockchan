package protocol

import (
	"slava0135/blockchan/blockgen"
	"testing"
)

func TestPackMessage_SendBlock(t *testing.T) {
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	var msg = PackMessage(block)
	var received = UnpackMessage(msg)
	if !block.Equal(received) {
		t.Fatalf("failed to send message with block")
	}
}
