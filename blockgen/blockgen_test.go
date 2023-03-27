package blockgen

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenerateNextFrom_Index(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{}, nil)
	var wantIndex = prev.Index + 1
	if next.Index != wantIndex {
		t.Fatalf("index = %d; want %d", next.Index, wantIndex)
	}
}

func TestGenerateNextFrom_PrevHash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	prev.Hash = []byte{}
	var next = GenerateNextFrom(prev, Data{}, nil)
	if !bytes.Equal(next.PrevHash, prev.Hash) {
		t.Fatalf("previous hash is incorrect")
	}
}

func TestGenerateNextFrom_Data(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var data Data
	var text = []byte{11, 14, 14, 15}
	copy(data[:], text)
	var next = GenerateNextFrom(prev, data, nil)
	if !bytes.Equal(data[:], next.Data[:]) {
		t.Fatalf("data = %x; want %x", next.Data, data)
	}
}

func TestGenerateNextFrom_Hash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{}, nil)
	var sum = fmt.Sprintf("%x", next.Hash)
	var ending = sum[len(sum)-4:]
	const want = "0000"
	if ending != want {
		t.Fatalf("got hash ending with %s; want %s", ending, want)
	}
}

func TestGenerateNextFrom_HashUseIndex(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var nextOne = GenerateNextFrom(prev, Data{}, nil)
	var sumOne = nextOne.Hash
	prev.Index += 1
	var nextTwo = GenerateNextFrom(prev, Data{}, nil)
	var sumTwo = nextTwo.Hash
	if bytes.Equal(sumOne, sumTwo) {
		t.Fatalf("got same hashes for different index values")
	}
}

func TestGenerateNextFrom_HashUsePrevHash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var nextOne = GenerateNextFrom(prev, Data{}, nil)
	prev.Hash = []byte{}
	var nextTwo = GenerateNextFrom(prev, Data{}, nil)
	if bytes.Equal(nextOne.Hash, nextTwo.Hash) {
		t.Fatalf("got same hashes for different prev hashes values")
	}
}

func TestGenerateNextFrom_HashUseData(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var nextOne = GenerateNextFrom(prev, Data{1}, nil)
	var nextTwo = GenerateNextFrom(prev, Data{2}, nil)
	if bytes.Equal(nextOne.Hash, nextTwo.Hash) {
		t.Fatalf("got same hashes for different data values")
	}
}

func TestGenerateGenesisBlock(t *testing.T) {
	var b = GenerateGenesisBlock()
	var want = Index(0)
	if b.Index != want {
		t.Fatalf("index = %d; want %d", b.Index, want)
	}
}

func TestHasValidHash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{42}, nil)
	if !next.HasValidHash() {
		t.Fatalf("generated block hash is not valid")
	}
}

func TestGenerateNextFrom_Cancel(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var cancel = make(chan struct{}, 1)
	var next Block
	go func() {
		next = GenerateNextFrom(prev, Data{}, cancel)
	}()
	cancel <- struct{}{}
	if next.HasValidHash() {
		t.Fatalf("block generation was not cancelled")
	}
}

func TestAreEqualBlocks_Index(t *testing.T) {
	var a = GenerateGenesisBlock()
	var b = GenerateGenesisBlock()
	b.Index += 1
	if a.Equal(b) {
		t.Fatalf("blocks with different index are equal")
	}
}

func TestAreEqualBlocks_PrevHash(t *testing.T) {
	var a = GenerateGenesisBlock()
	var b = GenerateGenesisBlock()
	b.PrevHash = append(b.PrevHash, 42)
	if a.Equal(b) {
		t.Fatalf("blocks with different prev hashes are equal")
	}
}

func TestAreEqualBlocks_Hash(t *testing.T) {
	var a = GenerateGenesisBlock()
	var b = GenerateGenesisBlock()
	b.Hash = append(b.Hash, 42)
	if a.Equal(b) {
		t.Fatalf("blocks with different hashes are equal")
	}
}

func TestAreEqualBlocks_Data(t *testing.T) {
	var a = GenerateGenesisBlock()
	var b = GenerateGenesisBlock()
	b.Data[0] += 1
	if a.Equal(b) {
		t.Fatalf("blocks with different data are equal")
	}
}

func TestAreEqualBlocks_Nonce(t *testing.T) {
	var a = GenerateGenesisBlock()
	var b = GenerateGenesisBlock()
	b.Nonce += 1
	if a.Equal(b) {
		t.Fatalf("blocks with different nonce are equal")
	}
}
