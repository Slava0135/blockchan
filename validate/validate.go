package validate

import (
	"slava0135/blockchan/blockgen"
)

func IsChainValid(chain []blockgen.Block) bool {
	for i, b := range chain {
		if string(blockgen.CalculateHashFrom(b).Sum(nil)) != string(b.Hash.Sum(nil)) {
			return false
		}
		if i > 0 {
			if chain[i].PrevHash != chain[i-1].Hash {
				return false
			}
		}
	}
	return true
}
