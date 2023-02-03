package stool

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
		want:      "A -> B *",
	}, {
		name:      "multiple repetition",
		endpoints: []string{"A", "A", "A", "B", "B", "C", "A", "A"},
		want:      "A * -> B * -> C -> A *",
	}, {
		name:      "node",
		endpoints: []string{"A", "B", "C", "A", "B", "C", "D"},
		want:      "(A -> B -> C)* -> D",
	}, {
		name:      "repetition in the node",
		endpoints: []string{"A", "B", "B", "C", "A", "B", "C", "D"},
		want:      "(A -> B * -> C)* -> D",
	}, {
		name:      "nest node",
		endpoints: []string{"A", "B", "C", "B", "C", "D", "B", "C", "D", "E"},
		want:      "A -> ((B -> C)* -> D)* -> E",
	}, {
		name:      "repeat the tail of node",
		endpoints: []string{"A", "B", "A", "B", "B", "C"},
		want:      "(A -> B *)* -> C",
	}, {
		name:      "repeat the tails of node",
		endpoints: []string{"A", "B", "C", "A", "B", "C", "B", "C", "D"},
		want:      "(A -> (B -> C)*)* -> D",
	}, {
		name:      "repeat the tails of pattern2",
		endpoints: []string{"A", "B", "C", "B", "C", "A", "B", "C", "B", "C", "D"},
		want:      "(A -> (B -> C)*)* -> D",
	}, {
		name:      "count merge",
		endpoints: []string{"A", "B", "B", "C", "A", "A", "B", "C", "C", "D"},
		want:      "(A * -> B * -> C *)* -> D",
	}, {
		name:      "nest count merge",
		endpoints: []string{"A", "B", "C", "B", "C", "D", "B", "C", "C", "D", "E"},
		want:      "A -> ((B -> C *)* -> D)* -> E",
	}, {
		name:      "structure merge1",
		endpoints: []string{"A", "B", "C", "C", "A", "B", "B", "C", "B", "C", "D"},
		want:      "(A -> (B * -> C *)*)* -> D",
	}, {
		name:      "structure merge2",
		endpoints: []string{"A", "B", "C", "C", "A", "B", "B", "C", "B", "C", "D"},
		want:      "(A -> (B * -> C *)*)* -> D",
	}, {
		name:      "structure merge3",
		endpoints: []string{"A", "B", "C", "C", "D", "A", "B", "B", "C", "B", "C", "D", "E"},
		want:      "(A -> (B * -> C *)* -> D)* -> E",
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
		want: "(A * -> B -> C)* -> A",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &node{}
			for _, endpoint := range tt.endpoints {
				root.Append(endpoint)
			}

			assert.Equal(t, tt.want, root.String())
		})
	}
}

func Test(t *testing.T) {
	root := &node{}
	for i, p := range []string{"A", "B", "C", "B", "C", "A", "B", "A", "B", "C", "D"} {
		root.Append(p)
		fmt.Printf("%2d: %s\n", i, root.String())
	}
}
