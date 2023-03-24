package blockgen

import (
	"bytes"
	"encoding"
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

func TestHasValidHash_Nil(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{}, nil)
	next.Hash = nil
	if next.HasValidHash() {
		t.Fatalf("nil hash is valid")
	}
}

func TestGenerateNextFrom_Cancel(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var cancel = false
	var next Block
	go func() {
		next = GenerateNextFrom(prev, Data{}, &cancel)
	}()
	cancel = true
	if next.HasValidHash() {
		t.Fatalf("block generation was not cancelled")
	}
}

func TestBlockMarshal(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{1, 2, 3, 4, 5}, nil)
	var i any = &next
	marshaler, ok := i.(encoding.BinaryMarshaler)
	if !ok {
		t.Fatalf("block does not implement encoding.BinaryMarshaler")
	}
	data, err := marshaler.MarshalBinary()
	if err != nil {
		t.Fatalf("failed to marshal valid block")
	}
	var restored = Block{}
	i = &restored
	unmarshaler, ok := i.(encoding.BinaryUnmarshaler)
	if !ok {
		t.Fatalf("block does not implement encoding.BinaryUnmarshaler")
	}
	unmarshaler.UnmarshalBinary(data)
	if !AreEqualBlocks(next, restored) {
		t.Fatalf("original block and restored blocks are different")
	}
}

func TestAreEqualBlocks_Index(t *testing.T) {
	var a = GenerateGenesisBlock()
	var b = GenerateGenesisBlock()
	b.Index += 1
	if AreEqualBlocks(a, b) {
		t.Fatalf("blocks with different index are equal")
	}
}
