package stool

import (
	"golang.org/x/exp/slices"
)

type node struct {
	value    string
	children []node
	elems    int
}

func newNode(children []node) *node {
	return &node{children: children, elems: elems(children)}
}

func newLeaf(value string) *node {
	return &node{value: value, elems: 1}
}

func (n *node) first() *node {
	if n.IsLeaf() {
		return n
	}
	return n.children[0].first()
}

func (n *node) last() *node {
	if n.IsLeaf() {
		return n
	}
	return n.children[len(n.children)-1].last()
}

func (n *node) IsLeaf() bool {
	return n.children == nil || len(n.children) == 0
}

func (n *node) String(root bool) string {
	if n.IsLeaf() {
		return n.value
	}

	str := ""
	for i, child := range n.children {
		if i == 0 {
			str += child.String(false)
		} else {
			str += " -> " + child.String(false)
		}
	}

	if root {
		return str
	}
	return "(" + str + ")*"
}

func (n *node) Append(value string) {
	n.children = append(n.children, *newLeaf(value))
	for i := len(n.children) - 2; i >= 0; i-- {
		if n.children[i].last().value != value {
			continue
		}

		s := n.children[i+1:]
		for j := i; j >= 0; j-- {
			// Check if n.children[j:i+1] equals s
			if mergedNode, ok := merge(n.children[j:i+1], s); ok {
				n.elems += -elems(n.children[j:]) + mergedNode.elems
				n.children = n.children[:j+1]
				n.children[j] = *mergedNode
				return
			}

			// Check if n.children[i]'s tail matches s
			if i == j && n.children[i].elems > elems(s) {
				for k := len(n.children[i].children) - 1; k >= 0; k-- {
					if mergedNode, ok := merge(n.children[i].children[k:], s); ok {
						n.children[i].elems += -elems(n.children[i].children[k:]) + mergedNode.elems
						n.children[i].children = n.children[i].children[:k+1]
						n.children[i].children[k] = *mergedNode
						n.elems -= elems(n.children[i+1:])
						n.children = n.children[:i+1]
						return
					}
				}
			}

		}
	}

	n.elems += 1
}

func merge(src, dest []node) (*node, bool) {
	if !flatCompare(src, dest) {
		return nil, false
	}

	newChildren := _merge(src, dest)
	if newChildren == nil {
		return nil, false
	}

	if len(newChildren) == 1 && !newChildren[0].IsLeaf() {
		return &newChildren[0], true
	}
	return newNode(newChildren), true
}

func _merge(src, dest []node) []node {
	if len(src) == 0 || len(dest) == 0 {
		return nil
	}

	if len(src) == 1 && src[0].IsLeaf() && len(dest) == 1 && dest[0].IsLeaf() {
		if src[0].value != dest[0].value {
			return nil
		}
		return []node{*newLeaf(src[0].value)}
	}

	s := src
	d := dest
	if len(src) == 1 && !src[0].IsLeaf() {
		s = src[0].children
	}
	if len(dest) == 1 && !dest[0].IsLeaf() {
		d = dest[0].children
	}

	if len(src) == 1 || len(dest) == 1 {
		return []node{*newNode(_merge(s, d))}
	}

	for i := 1; i < len(src); i++ {
		for j := 1; j < len(dest); j++ {
			if elems(src[:i]) != elems(dest[:j]) {
				continue
			}
			return append(_merge(src[:i], dest[:j]), _merge(src[i:], dest[j:])...)
		}
	}

	return nil
}

func elems(nodes []node) int {
	s := 0
	for _, n := range nodes {
		s += n.elems
	}
	return s
}

func flatCompare(src, dest []node) bool {
	if src[0].first().value != dest[0].first().value {
		return false
	}
	srcSize := elems(src)
	destSize := elems(dest)
	if srcSize != destSize {
		return false
	}

	return slices.Equal(
		flatten(src, make([]string, srcSize)),
		flatten(dest, make([]string, destSize)))
}

func flatten(ns []node, result []string) []string {
	i := 0
	for _, n := range ns {
		if n.IsLeaf() {
			result[i] = n.value
			i++
		} else {
			flatten(n.children, result[i:i+n.elems])
			i += n.elems
		}
	}
	return result
}
