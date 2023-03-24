package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/encode"
)

type SendBlockMsg struct {
	Block blockgen.Block
}

func PackMessage(input any) []byte {
	switch v := input.(type) {
	case SendBlockMsg:
		var encoded, _ = encode.Encode(v.Block)
		return encoded
	}
	return nil
}

func UnpackMessage(text []byte) SendBlockMsg {
	var decoded, _ = encode.Decode(text)
	return SendBlockMsg{decoded}
}
