package mesh

import (
	"slava0135/blockchan/node"
	"testing"
)

func TestNodeMesh_Interface(t *testing.T) {
	var _ node.Mesh = &NodeMesh{}
}
