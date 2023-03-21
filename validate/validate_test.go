package validate

import (
	"slava0135/blockchan/blockgen"
	"testing"
)

func TestIsValidChain_Empty(t *testing.T) {
	if !IsValidChain(nil) {
		t.Fatalf("empty chain is not valid")
	}
}

func TestIsValidChain_InvalidHash(t *testing.T) {
	var one = blockgen.GenerateGenesisBlock()
	one.Nonce += 1
	if IsValidChain([]blockgen.Block{one}) {
		t.Fatalf("block with different hash is valid")
	}
}

func TestIsValidChain_ValidChain(t *testing.T) {
	var gen = blockgen.GenerateGenesisBlock()
	var chain = []blockgen.Block{gen}
	for i := byte(0); i < 10; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}))
	}
	if !IsValidChain(chain) {
		t.Fatalf("valid chain is invalid")
	}
}

func TestIsValidChain_PrevHashNotMatching(t *testing.T) {
	var one = blockgen.GenerateGenesisBlock()
	var two = blockgen.GenerateNextFrom(one, blockgen.Data{})
	one.Data[0] += 1
	one.GenerateValidHash()
	if IsValidChain([]blockgen.Block{one, two}) {
		t.Fatalf("block with invalid prev hash is valid")
	}
}

func TestIsValidChain_WrongIndex(t *testing.T) {
	var gen = blockgen.GenerateGenesisBlock()
	var chain = []blockgen.Block{gen}
	for i := byte(0); i < 10; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}))
		const mut = 3
		if i == mut-1 {
			chain[mut].Index += 42
			chain[mut].GenerateValidHash()
		}
	}
	if IsValidChain(chain) {
		t.Fatalf("block with invalid index is valid")
	}
}
