package protocol

import (
	"slava0135/blockchan/blockgen"
	"slava0135/blockchan/encode"
)

func PackMessage(b blockgen.Block) []byte {
	encoded, _ := encode.Encode(b)
	return encoded
}

func UnpackMessage(msg []byte) blockgen.Block {
	decoded, _ := encode.Decode(msg)
	return decoded
}
