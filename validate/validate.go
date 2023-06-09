package validate

import (
	"bytes"
	"slava0135/blockchan/blockgen"
)

func IsValidChain(chain []blockgen.Block) bool {
	for i, b := range chain {
		if !b.HasValidHash() {
			return false
		}
		if i > 0 {
			if chain[i].Index != chain[i-1].Index+1 {
				return false
			}
			if !bytes.Equal(chain[i].PrevHash, chain[i-1].Hash) {
				return false
			}
		}
	}
	return true
}

func AreEqualChains(a, b []blockgen.Block) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}
