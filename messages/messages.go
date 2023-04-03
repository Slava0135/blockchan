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

func UnpackMessage(text []byte) (msg any, err error) {
	var slices = bytes.SplitN(text, []byte{'\n'}, 3)
	var gotLen = len(slices)
	var msgType = string(slices[0])
	switch string(slices[0]) {
	case sendBlock:
		const wantLen = 3
		if gotLen != wantLen {
			return nil, fmt.Errorf("<%s> message got %d args instead of %d", msgType, gotLen, wantLen)
		}
		var lastIndex, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		var decoded, _ = encode.Decode(slices[2])
		return SendBlockMsg{decoded, lastIndex}, nil
	case requestBlocks:
		const wantLen = 2
		if gotLen != wantLen {
			return nil, fmt.Errorf("<%s> message got %d args instead of %d", msgType, gotLen, wantLen)
		}
		var index, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		return RequestBlocksMsg{index}, nil
	case dropBlock:
		const wantLen = 3
		if gotLen != wantLen {
			return nil, fmt.Errorf("<%s> message got %d args instead of %d", msgType, gotLen, wantLen)
		}
		var lastIndex, _ = strconv.ParseUint(string(slices[1]), 10, 64)
		var decoded, _ = encode.Decode(slices[2])
		return DropBlockMsg{decoded, lastIndex}, nil
	}
	return nil, fmt.Errorf("input did not match any message pattern")
}
