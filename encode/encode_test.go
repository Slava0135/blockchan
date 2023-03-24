package encode

import (
	"fmt"
	"slava0135/blockchan/blockgen"
	"testing"
)

func TestEncodeBlock(t *testing.T) {
	var g = blockgen.GenerateGenesisBlock()
	var block = blockgen.GenerateNextFrom(g, blockgen.Data{1, 2, 3, 4, 5}, nil)
	var text, err = Encode(block)
	if err != nil {
		t.Fatalf("error when encoding block: %v", err)
	}
	restored, err := Decode(text)
	if err != nil {
		t.Fatalf("error when decoding block: %v", err)
	}
	fmt.Print(string(text))
	if !block.Equal(restored) {
		t.Fatalf("restored block not equals original block")
	}
}
