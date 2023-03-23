package blockgen

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"strconv"
)

type Block struct {
	Index    Index
	PrevHash hash.Hash
	Hash     hash.Hash
	Data     Data
	Nonce    Nonce
}

type Index uint64
type Data [256]byte
type Nonce uint64

func (b Block) HasValidHash() bool {
	if b.Hash == nil {
		return false
	}
	var hash = calculateHashFrom(b)
	return string(b.Hash.Sum(nil)) == string(hash.Sum(nil)) && hasValidEnding(hash)
}

func GenerateNextFrom(prev Block, data Data, cancel *bool) Block {
	var next = Block{}
	next.Index = prev.Index + 1
	next.PrevHash = prev.Hash
	next.Data = data
	next.Nonce = Nonce(0)
	next.GenerateValidHash(cancel)
	return next
}

func GenerateGenesisBlock() Block {
	var b = Block{}
	b.PrevHash = sha256.New()
	b.GenerateValidHash(nil)
	return b
}

func (b *Block) GenerateValidHash(cancel *bool) {
	for cancel == nil || !*cancel {
		var hash = calculateHashFrom(*b)
		if hasValidEnding(hash) {
			b.Hash = hash
			return
		}
		b.Nonce += 1
	}
}

func calculateHashFrom(b Block) hash.Hash {
	var hash = sha256.New()
	hash.Write([]byte(strconv.FormatUint(uint64(b.Nonce), 10)))
	hash.Write([]byte(strconv.FormatUint(uint64(b.Index), 10)))
	hash.Write(b.PrevHash.Sum(nil))
	hash.Write(b.Data[:])
	return hash
}

func hasValidEnding(h hash.Hash) bool {
	var sum = h.Sum(nil)
	return sum[len(sum)-1] == 0 && sum[len(sum)-2] == 0
}

func (b *Block) MarshalBinary() (data []byte, err error) {
	data = binary.BigEndian.AppendUint64(data, uint64(b.Index))
	return
}

func (b *Block) UnmarshalBinary(data []byte) error {
	var _, val = consumeUint64(data)
	b.Index = Index(val)
	return nil
}

func consumeUint64(b []byte) ([]byte, uint64) {
	_ = b[7]
	x := uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	return b[8:], x
}


func AreSameBlocks(a, b Block) bool {
	return a.Index == b.Index
}
