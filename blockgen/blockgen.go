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

func GenerateNextFrom(prev Block) Block {
	next := Block{}
	next.Index = prev.Index + 1
	next.PrevHash = prev.Hash
	hash := sha256.New()
	nonce := 0
	sum := hash.Sum(nil)
	for sum[len(sum)-1] % 16 != 0 {
		hash.Reset()
		hash.Write([]byte(strconv.Itoa(nonce)))
		sum = hash.Sum(nil)
		nonce += 1
	}
	next.Hash = hash
	next.Nonce = Nonce(nonce)
	return next;
}
