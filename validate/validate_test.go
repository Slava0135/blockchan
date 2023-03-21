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
