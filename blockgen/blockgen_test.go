package blockgen

import (
	"crypto/sha256"
	"testing"
	"fmt"
)

func TestGenerateNextFrom_Index(t *testing.T) {
	var prev = Block{}
	var next = GenerateNextFrom(prev)
	var wantIndex = prev.Index + 1
	if next.Index != wantIndex {
		t.Errorf("index = %d; want %d", next.Index, wantIndex)
	}
}

func TestGenerateNextFrom_PrevHash(t *testing.T) {
	var prev = Block{}
	prev.Hash = sha256.New()
	prev.Hash.Write([]byte("test"))
	var next = GenerateNextFrom(prev)
	var prevSum = string(next.PrevHash.Sum(nil))
	var wantSum = string(prev.Hash.Sum(nil))
	if prevSum != wantSum {
		t.Errorf("previous hash is incorrect")
	}
}

func TestGenerateNextFrom_Hash(t *testing.T) {
	var prev = Block{}
	var next = GenerateNextFrom(prev)
	var sum = fmt.Sprintf("%x", next.Hash.Sum(nil))
	var ending = sum[len(sum)-1:]
	const want = "0"
	if ending != want {
		t.Errorf("got hash ending with %s; want %s", ending, want)
	}
}
