package validate

import (
	"slava0135/blockchan/blockgen"
	"testing"
)

func TestIsChainValid_Empty(t *testing.T) {
	if !IsChainValid(nil) {
		t.Errorf("empty chain is not valid")
	}
}

func TestIsChainValid_InvalidHash(t *testing.T) {
	var one = blockgen.GenerateGenesisBlock()
	one.Nonce += 1
	if IsChainValid([]blockgen.Block{one}) {
		t.Errorf("block with different hash is valid")
	}
}

func TestIsChainValid_ValidChain(t *testing.T) {
	var gen = blockgen.GenerateGenesisBlock()
	var chain = []blockgen.Block{gen}
	for i := byte(0); i < 10; i += 1 {
		chain = append(chain, blockgen.GenerateNextFrom(chain[i], blockgen.Data{}))
	}
	if !IsChainValid(chain) {
		t.Errorf("valid chain is invalid")
	}
}
