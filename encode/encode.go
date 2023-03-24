package encode

import (
	"encoding/json"
	"slava0135/blockchan/blockgen"
)

type encodingBlock struct {
	Index    uint64
	PrevHash []byte
	Hash     []byte
	Data     []byte
	Nonce    uint64
}

func Encode(b blockgen.Block) ([]byte, error) {
	var e = encodingBlock{}
	e.Index = uint64(b.Index)
	e.PrevHash = b.PrevHash
	e.Hash = b.Hash
	e.Data = b.Data[:]
	e.Nonce = uint64(b.Nonce)
	return json.MarshalIndent(e, "", "\t")
}

func Decode(text []byte) (blockgen.Block, error) {
	var e = encodingBlock{}
	var err = json.Unmarshal(text, &e)
	var b = blockgen.Block{}
	b.Index = blockgen.Index(e.Index)
	b.PrevHash = e.PrevHash
	b.Hash = e.Hash
	b.Data = blockgen.Data(e.Data)
	b.Nonce = blockgen.Nonce(e.Nonce)
	return b, err
}
