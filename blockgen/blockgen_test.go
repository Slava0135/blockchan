package blockgen

import (
	"testing"
	"crypto/sha256"
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
	if next.PrevHash != prev.Hash {
		t.Errorf("Next block: previous hash is incorrect")
	}
}
