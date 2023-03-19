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

func GenerateNextFrom(prev Block, data Data) Block {
	var next = Block{}
	next.Index = prev.Index + 1
	next.PrevHash = prev.Hash
	next.Data = data
	next.Nonce = Nonce(0)
	for {
		var hash = CalculateHashFrom(next)
		var sum = hash.Sum(nil)
		if sum[len(sum)-1]%16 == 0 {
			next.Hash = hash
			return next
		}
		next.Nonce += 1
	}
}

func GenerateGenesisBlock() Block {
	var b = Block{}
	b.Hash = sha256.New()
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
