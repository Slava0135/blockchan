package blockgen

import (
	"hash"
	"crypto/sha256"
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
	var hash = sha256.New()
	var nonce = 0
	var sum = hash.Sum(nil)
	for sum[len(sum)-1] % 16 != 0 {
		hash.Reset()
		hash.Write([]byte(strconv.Itoa(nonce)))
		hash.Write([]byte(strconv.Itoa(next.Index)))
		hash.Write(next.PrevHash.Sum(nil))
		hash.Write(data[:])
		sum = hash.Sum(nil)
		nonce += 1
	}
	next.Hash = hash
	next.Nonce = Nonce(nonce)
	return next;
}
