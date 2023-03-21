package validate

import (
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
			if chain[i].PrevHash != chain[i-1].Hash {
				return false
			}
		}
	}
	return true
}
