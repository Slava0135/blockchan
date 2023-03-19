package blockgen

import (
	"testing"
)

func TestGenerateNextFrom(t *testing.T) {
	var prev = Block{}
	var next = GenerateNextFrom(prev)
	var wantIndex = prev.Index + 1
	if next.Index != wantIndex {
		t.Errorf("Next block index = %d; want %d", next.Index, wantIndex)
	}
}
