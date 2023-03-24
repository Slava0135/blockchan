package protocol

import (
	"bytes"
	"fmt"
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/encode"
	"strconv"
)

const (
	SendBlock = "SENDING BLOCK"
	AskForBlocks = "ASKING FOR BLOCKS"
)

type SendBlockMsg struct {
	Block blockgen.Block
}

type AskForBlocksMsg struct {
	Index uint64
}

func PackMessage(input any) []byte {
	switch v := input.(type) {
	case SendBlockMsg:
		var encoded, _ = encode.Encode(v.Block)
		return []byte(fmt.Sprintf("%s\n%s", SendBlock, encoded))
	case AskForBlocksMsg:
		return []byte(fmt.Sprintf("%s\n%d", AskForBlocks, v.Index))
	}
	return nil
}

func UnpackMessage(text []byte) any {
	var slices = bytes.SplitN(text, []byte{'\n'}, 2)
	switch string(slices[0]) {
	case SendBlock:
		var decoded, _ = encode.Decode(slices[1])
		return SendBlockMsg{decoded}
	case AskForBlocks:
		var index, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		return AskForBlocksMsg{index}
	}
	return nil
}
