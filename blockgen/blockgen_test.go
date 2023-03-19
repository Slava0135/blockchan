package blockgen

import (
	"crypto/sha256"
	"testing"
)

func TestGenerateNextFrom_Index(t *testing.T) {
	var prev = Block{}
	var next = GenerateNextFrom(prev)
	var wantIndex = prev.Index + 1
	if next.Index != wantIndex {
		t.Errorf("Next block index = %d; want %d", next.Index, wantIndex)
	}
}

func TestGenerateNextFrom_PrevHash(t *testing.T) {
	var prev = Block{}
	prev.Hash = sha256.New()
	prev.Hash.Write([]byte("test"))
	var next = GenerateNextFrom(prev)
	var prevSum = string(next.PrevHash.Sum([]byte{}))
	var wantSum = string(prev.Hash.Sum([]byte{}))
	if prevSum != wantSum {
		t.Errorf("Next block: previous hash is incorrect")
	}
}
