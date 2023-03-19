package blockgen

import (
	"fmt"
	"testing"
)

func TestGenerateNextFrom_Index(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{})
	var wantIndex = prev.Index + 1
	if next.Index != wantIndex {
		t.Errorf("index = %d; want %d", next.Index, wantIndex)
	}
}

func TestGenerateNextFrom_PrevHash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	prev.Hash.Write([]byte("test"))
	var next = GenerateNextFrom(prev, Data{})
	var prevSum = string(next.PrevHash.Sum(nil))
	var wantSum = string(prev.Hash.Sum(nil))
	if prevSum != wantSum {
		t.Errorf("previous hash is incorrect")
	}
}

func TestGenerateNextFrom_Data(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var data Data
	var text = []byte{11, 14, 14, 15}
	copy(data[:], text)
	var next = GenerateNextFrom(prev, data)
	if next.Data != data {
		t.Errorf("data = %x; want %x", next.Data, data)
	}
}

func TestGenerateNextFrom_Hash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var next = GenerateNextFrom(prev, Data{})
	var sum = fmt.Sprintf("%x", next.Hash.Sum(nil))
	var ending = sum[len(sum)-1:]
	const want = "0"
	if ending != want {
		t.Errorf("got hash ending with %s; want %s", ending, want)
	}
}

func TestGenerateNextFrom_HashUseIndex(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var nextOne = GenerateNextFrom(prev, Data{})
	var sumOne = string(nextOne.Hash.Sum(nil))
	prev.Index += 1
	var nextTwo = GenerateNextFrom(prev, Data{})
	var sumTwo = string(nextTwo.Hash.Sum(nil))
	if sumOne == sumTwo {
		t.Errorf("got same hashes for different index values")
	}
}

func TestGenerateNextFrom_HashUsePrevHash(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var nextOne = GenerateNextFrom(prev, Data{})
	var sumOne = string(nextOne.Hash.Sum(nil))
	prev.Hash.Write([]byte{42})
	var nextTwo = GenerateNextFrom(prev, Data{})
	var sumTwo = string(nextTwo.Hash.Sum(nil))
	if sumOne == sumTwo {
		t.Errorf("got same hashes for different prev hashes values")
	}
}

func TestGenerateNextFrom_HashUseData(t *testing.T) {
	var prev = GenerateGenesisBlock()
	var nextOne = GenerateNextFrom(prev, Data{1})
	var sumOne = string(nextOne.Hash.Sum(nil))
	var nextTwo = GenerateNextFrom(prev, Data{2})
	var sumTwo = string(nextTwo.Hash.Sum(nil))
	if sumOne == sumTwo {
		t.Errorf("got same hashes for different data values")
	}
}
