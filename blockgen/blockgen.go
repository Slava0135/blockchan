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

func (b Block) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (b Block) UnmarshalBinary(data []byte) error {
	return nil
}
