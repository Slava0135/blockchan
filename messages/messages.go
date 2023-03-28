package messages

import (
	"bytes"
	"fmt"
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/encode"
	"strconv"
)

const (
	sendBlock     = "SENDING BLOCK"
	requestBlocks = "REQUESTING BLOCKS"
	dropBlock     = "DROP BLOCK"
)

type SendBlockMsg struct {
	Block          blockgen.Block
	LastBlockIndex uint64
}

type RequestBlocksMsg struct {
	Index uint64
}

type DropBlockMsg struct {
	Block          blockgen.Block
	LastBlockIndex uint64
}

func PackMessage(input any) []byte {
	switch v := input.(type) {
	case SendBlockMsg:
		var encoded, _ = encode.Encode(v.Block)
		return []byte(fmt.Sprintf("%s\n%d\n%s", sendBlock, v.LastBlockIndex, encoded))
	case RequestBlocksMsg:
		return []byte(fmt.Sprintf("%s\n%d", requestBlocks, v.Index))
	case DropBlockMsg:
		var encoded, _ = encode.Encode(v.Block)
		return []byte(fmt.Sprintf("%s\n%d\n%s", dropBlock, v.LastBlockIndex, encoded))
	}
	return nil
}

func UnpackMessage(text []byte) any {
	var slices = bytes.SplitN(text, []byte{'\n'}, 3)
	switch string(slices[0]) {
	case sendBlock:
		var lastIndex, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		var decoded, _ = encode.Decode(slices[2])
		return SendBlockMsg{decoded, lastIndex}
	case requestBlocks:
		if len(slices) != 2 {
			return nil
		}
		var index, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		return RequestBlocksMsg{index}
	case dropBlock:
		var lastIndex, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		var decoded, _ = encode.Decode(slices[2])
		return DropBlockMsg{decoded, lastIndex}
	}
	return nil
}
