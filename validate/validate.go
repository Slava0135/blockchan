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

func AreSameChains(a, b []blockgen.Block) bool {
	for i := range a {
		if string(a[i].Hash.Sum(nil)) != string(b[i].Hash.Sum(nil)) {
			return false
		}
	}
	return true
}
