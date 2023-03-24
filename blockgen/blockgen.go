package blockgen

import (
	"bytes"
	"crypto/sha256"
	"strconv"
)

type Block struct {
	Index    Index
	PrevHash HashSum
	Hash     HashSum
	Data     Data
	Nonce    Nonce
}

type Index uint64
type HashSum []byte
type Data [256]byte
type Nonce uint64

func (b Block) HasValidHash() bool {
	if b.Hash == nil {
		return false
	}
	var hash = calculateHashFrom(b)
	return bytes.Equal(b.Hash, hash) && hasValidEnding(hash)
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
	b.PrevHash = []byte{}
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

func calculateHashFrom(b Block) HashSum {
	var hash = sha256.New()
	hash.Write([]byte(strconv.FormatUint(uint64(b.Nonce), 10)))
	hash.Write([]byte(strconv.FormatUint(uint64(b.Index), 10)))
	hash.Write(b.PrevHash)
	hash.Write(b.Data[:])
	return hash.Sum(nil)
}

func hasValidEnding(h HashSum) bool {
	return bytes.HasSuffix(h, []byte{0, 0})
}

func (a Block) Equal(b Block) bool {
	return a.Index == b.Index &&
		bytes.Equal(a.PrevHash, b.PrevHash) &&
		bytes.Equal(a.Hash, b.Hash) &&
		bytes.Equal(a.Hash, b.Hash) &&
		bytes.Equal(a.Data[:], b.Data[:]) &&
		a.Nonce == b.Nonce
}
