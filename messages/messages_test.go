package messages

import (
	"slava0135/blockchan/blockgen"
	"testing"
)

func TestPackMessage_SendBlock(t *testing.T) {
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	var msg = PackMessage(SendBlockMsg{block, uint64(block.Index)})
	var unpacked = UnpackMessage(msg)
	var received, ok = unpacked.(SendBlockMsg)
	if !ok {
		t.Fatalf("failed to determine message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("failed to send message with block")
	}
	if received.LastBlockIndex != uint64(block.Index) {
		t.Fatalf("last block index is incorrect")
	}
}

func TestPackMessage_AskForBlocks(t *testing.T) {
	var index uint64 = 3
	var msg = PackMessage(RequestBlocksMsg{index})
	var unpacked = UnpackMessage(msg)
	var received, ok = unpacked.(RequestBlocksMsg)
	if !ok {
		t.Fatalf("failed to determine message type")
	}
	if received.Index != index {
		t.Fatalf("failed to ask for message")
	}
}

func TestPackMessage_DropBlock(t *testing.T) {
	var block = blockgen.GenerateNextFrom(blockgen.GenerateGenesisBlock(), blockgen.Data{1, 2, 3}, nil)
	var msg = PackMessage(DropBlockMsg{block, uint64(block.Index)})
	var unpacked = UnpackMessage(msg)
	var received, ok = unpacked.(DropBlockMsg)
	if !ok {
		t.Fatalf("failed to determine message type")
	}
	if !block.Equal(received.Block) {
		t.Fatalf("failed to send message with block")
	}
	if received.LastBlockIndex != uint64(block.Index) {
		t.Fatalf("last block index is incorrect")
	}
}

func TestUnpackMessage_InvalidMsg(t *testing.T) {
	var unpacked = UnpackMessage([]byte(requestBlocks))
	if unpacked != nil {
		t.Fatalf("accepted invalid message")
	}
	unpacked = UnpackMessage([]byte(requestBlocks))
	if unpacked != nil {
		t.Fatalf("accepted invalid message")
	}
	unpacked = UnpackMessage([]byte(sendBlock))
	if unpacked != nil {
		t.Fatalf("accepted invalid message")
	}
	unpacked = UnpackMessage([]byte(dropBlock))
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

func FuzzUnpackMessage(f *testing.F) {
	f.Fuzz(func(t *testing.T, text []byte) {
		UnpackMessage(text)
	})
}
