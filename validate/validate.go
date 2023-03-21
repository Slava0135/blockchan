package validate

import (
	"slava0135/blockchan/blockgen"
)

func IsChainValid(chain []blockgen.Block) bool {
	for _, b := range chain {
		if string(blockgen.CalculateHashFrom(b).Sum(nil)) != string(b.Hash.Sum(nil)) {
			return false
		} 
	}
	return true
}
