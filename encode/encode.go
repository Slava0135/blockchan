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

func Decode(text []byte) (b blockgen.Block, err error) {
	var e = encodingBlock{}
	err = json.Unmarshal(text, &e)
	if err != nil {
		return
	}
	b = blockgen.Block{}
	b.Index = blockgen.Index(e.Index)
	b.PrevHash = e.PrevHash
	b.Hash = e.Hash
	copy(b.Data[:], e.Data)
	b.Nonce = blockgen.Nonce(e.Nonce)
	return
}
