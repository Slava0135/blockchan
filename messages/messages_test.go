package messages

import (
	"slava0135/blockchan/blockgen"
	"testing"
)

func TestPackMessage_SendBlock(t *testing.T) {
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	var msg = PackMessage(SendBlockMsg{block})
	var unpacked = UnpackMessage(msg)
	var received, ok = unpacked.(SendBlockMsg)
	if !ok {
		t.Fatalf("failed to determine message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("failed to send message with block")
	}
}

func TestPackMessage_AskForBlocks(t *testing.T) {
	var index uint64 = 3
	var msg = PackMessage(AskForBlocksMsg{index})
	var unpacked = UnpackMessage(msg)
	var received, ok = unpacked.(AskForBlocksMsg)
	if !ok {
		t.Fatalf("failed to determine message type")
	}
	if received.Index != index {
		t.Fatalf("failed to ask for message")
	}
}

func TestUnpackMessage_InvalidMsg(t *testing.T) {
	var unpacked = UnpackMessage([]byte(askForBlocks))
	if unpacked != nil {
		t.Fatalf("accepted invalid message")
	}
}

func TestPackMessage_InvalidInput(t *testing.T) {
	var packed = PackMessage("marko zajc")
	if packed != nil {
		t.Fatalf("accepted invalid input")
	}
}

func TestUnpackMessage_InvalidInput(t *testing.T) {
	var unpacked = UnpackMessage([]byte("marko\nzajc"))
	if unpacked != nil {
		t.Fatalf("accepted invalid input")
	}
}
