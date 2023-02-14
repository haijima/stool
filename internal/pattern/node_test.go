package pattern

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pattern(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []string
		want      string
	}{{
		name:      "empty",
		endpoints: []string{},
		want:      "",
	}, {
		name:      "single",
		endpoints: []string{"A"},
		want:      "A",
	}, {
		name:      "normal",
		endpoints: []string{"A", "B", "C"},
		want:      "A -> B -> C",
	}, {
		name:      "repetition",
		endpoints: []string{"A", "B", "B", "B"},
		want:      "A -> (B)*",
	}, {
		name:      "multiple repetition",
		endpoints: []string{"A", "A", "A", "B", "B", "C", "A", "A"},
		want:      "(A)* -> (B)* -> C -> (A)*",
	}, {
		name:      "Node",
		endpoints: []string{"A", "B", "C", "A", "B", "C", "D"},
		want:      "(A -> B -> C)* -> D",
	}, {
		name:      "repetition in the Node",
		endpoints: []string{"A", "B", "B", "C", "A", "B", "C", "D"},
		want:      "(A -> (B)* -> C)* -> D",
	}, {
		name:      "nest Node",
		endpoints: []string{"A", "B", "C", "B", "C", "D", "B", "C", "D", "E"},
		want:      "A -> ((B -> C)* -> D)* -> E",
	}, {
		name:      "repeat the tail of Node",
		endpoints: []string{"A", "B", "A", "B", "B", "C"},
		want:      "(A -> (B)*)* -> C",
	}, {
		name:      "repeat the tails of pattern",
		endpoints: []string{"A", "B", "C", "A", "B", "C", "B", "C", "D"},
		want:      "(A -> (B -> C)*)* -> D",
	}, {
		name:      "repeat the tails of pattern2",
		endpoints: []string{"A", "B", "C", "B", "C", "A", "B", "C", "B", "C", "D"},
		want:      "(A -> (B -> C)*)* -> D",
	}, {
		name:      "repeat the tails of pattern3",
		endpoints: []string{"A", "B", "C", "D", "E", "B", "C", "D", "E", "C", "D", "E", "D", "E"},
		want:      "A -> (B -> (C -> (D -> E)*)*)*",
	}, {
		name:      "count Merge",
		endpoints: []string{"A", "B", "B", "C", "A", "A", "B", "C", "C", "D"},
		want:      "((A)* -> (B)* -> (C)*)* -> D",
	}, {
		name:      "nest count Merge",
		endpoints: []string{"A", "B", "C", "B", "C", "D", "B", "C", "C", "D", "E"},
		want:      "A -> ((B -> (C)*)* -> D)* -> E",
	}, {
		name:      "structure merge1",
		endpoints: []string{"A", "B", "C", "C", "A", "B", "B", "C", "B", "C", "D"},
		want:      "(A -> ((B)* -> (C)*)*)* -> D",
	}, {
		name:      "structure merge2",
		endpoints: []string{"A", "B", "C", "C", "A", "B", "B", "C", "B", "C", "D"},
		want:      "(A -> ((B)* -> (C)*)*)* -> D",
	}, {
		name:      "structure merge3",
		endpoints: []string{"A", "B", "C", "C", "D", "A", "B", "B", "C", "B", "C", "D", "E"},
		want:      "(A -> ((B)* -> (C)*)* -> D)* -> E",
	}, {
		name:      "unmergeable  structure",
		endpoints: []string{"A", "B", "C", "B", "C", "A", "B", "A", "B", "C", "D"},
		want:      "A -> (B -> C)* -> (A -> B)* -> C -> D",
	}, {
		name:      "reduce depth",
		endpoints: []string{"A", "B", "A", "B", "C", "A", "B", "C"},
		want:      "((A -> B)* -> C)*",
	}, {
		name:      "shortest match",
		endpoints: []string{"A", "B", "C", "A", "A", "B", "C", "A"},
		// not "(A -> B -> C -> A)*",
		want: "((A)* -> B -> C)* -> A",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRoot()
			for _, endpoint := range tt.endpoints {
				root.Append(endpoint)
			}

			assert.Equal(t, tt.want, root.String(true))
			assert.NoError(t, validateElem(root))
		})
	}
}

func Test(t *testing.T) {
	root := NewRoot()
	for i, p := range []string{"A", "B", "C", "B", "C", "A", "B", "A", "B", "C", "D"} {
		root.Append(p)
		fmt.Printf("%2d: %s\n", i, root.String(true))
	}
}

func BenchmarkNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root := NewRoot()
		for _, p := range []string{"A", "B", "C", "B", "C", "A", "B", "A", "B", "C", "D"} {
			root.Append(p)
		}
		root.String(true)
	}
}

func BenchmarkDeepNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root := NewRoot()
		for _, p := range []string{"A", "A", "B", "A", "B", "C", "A", "B", "C", "D", "A", "B", "C", "D", "E", "A", "B", "C", "D", "E", "F", "A", "B", "C", "D", "E", "F", "G", "A", "B", "C", "D", "E", "F", "G", "H", "A", "B", "C", "D", "E", "F", "G", "H", "I", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"} {
			root.Append(p)
		}
		root.String(true)
	}
}

func validateElem(node Node) error {
	if node.IsLeaf() {
		if node.leaves != 1 {
			return fmt.Errorf("node elem should be 1 but: %d, node.String() = %s", node.leaves, node.String(false))
		}
		return nil
	}

	if node.leaves != leaves(node.children) {
		return fmt.Errorf("node elem should be %d but: %d, node.String() = %s", leaves(node.children), node.leaves, node.String(false))
	}

	for _, child := range node.children {
		err := validateElem(child)
		if err != nil {
			return err
		}
	}
	return nil
}
