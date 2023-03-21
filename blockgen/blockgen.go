package blockgen

import (
	"crypto/sha256"
	"hash"
	"strconv"
)

type Block struct {
	Index    int
	PrevHash hash.Hash
	Hash     hash.Hash
	Data     Data
	Nonce    Nonce
}

type Data [256]byte
type Nonce int

func (b Block) HasValidHash() bool {
	var hash = CalculateHashFrom(b)
	return string(b.Hash.Sum(nil)) == string(hash.Sum(nil)) && hasValidEnding(hash)
}

func GenerateNextFrom(prev Block, data Data) Block {
	var next = Block{}
	next.Index = prev.Index + 1
	next.PrevHash = prev.Hash
	next.Data = data
	next.Nonce = Nonce(0)
	for {
		var hash = CalculateHashFrom(next)
		if hasValidEnding(hash) {
			next.Hash = hash
			return next
		}
		next.Nonce += 1
	}
}

func GenerateGenesisBlock() Block {
	var b = Block{}
	b.PrevHash = sha256.New()
	b.Hash = CalculateHashFrom(b)
	return b
}

func CalculateHashFrom(b Block) hash.Hash {
	var hash = sha256.New()
	hash.Write([]byte(strconv.Itoa(int(b.Nonce))))
	hash.Write([]byte(strconv.Itoa(b.Index)))
	hash.Write(b.PrevHash.Sum(nil))
	hash.Write(b.Data[:])
	return hash
}

func hasValidEnding(h hash.Hash) bool {
	var sum = h.Sum(nil)
	return sum[len(sum)-1] == 0 && sum[len(sum)-2] == 0
}
