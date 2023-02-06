package internal

import (
	"golang.org/x/exp/slices"
)

type Node struct {
	value    string
	children []Node
	elems    int
}

func NewNode(children []Node) *Node {
	return &Node{children: children, elems: elems(children)}
}

func NewLeaf(value string) *Node {
	return &Node{value: value, elems: 1}
}

func (n *Node) first() *Node {
	if n.IsLeaf() {
		return n
	}
	return n.children[0].first()
}

func (n *Node) last() *Node {
	if n.IsLeaf() {
		return n
	}
	return n.children[len(n.children)-1].last()
}

func (n *Node) IsLeaf() bool {
	return n.children == nil || len(n.children) == 0
}

func (n *Node) String(root bool) string {
	if n.IsLeaf() {
		return n.value
	}

	str := ""
	for i := range n.children {
		if i == 0 {
			str += n.children[i].String(false)
		} else {
			str += " -> " + n.children[i].String(false)
		}
	}

	if root {
		return str
	}
	return "(" + str + ")*"
}

func (n *Node) Append(value string) {
	n.children = append(n.children, *NewLeaf(value))
	for i := len(n.children) - 2; i >= 0; i-- {
		if n.children[i].last().value != value {
			continue
		}

		s := n.children[i+1:]
		for j := i; j >= 0; j-- {
			// Check if n.children[j:i+1] equals s
			if mergedNode, ok := Merge(n.children[j:i+1], s); ok {
				n.elems += -elems(n.children[j:]) + mergedNode.elems
				n.children = n.children[:j+1]
				n.children[j] = *mergedNode
				return
			}

			// Check if n.children[i]'s tail matches s
			if i == j && n.children[i].elems > elems(s) {
				for k := len(n.children[i].children) - 1; k >= 0; k-- {
					if mergedNode, ok := Merge(n.children[i].children[k:], s); ok {
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

func Merge(src, dest []Node) (*Node, bool) {
	if !flatCompare(src, dest) {
		return nil, false
	}

	newChildren := merge(src, dest)
	if newChildren == nil {
		return nil, false
	}

	if len(newChildren) == 1 && !newChildren[0].IsLeaf() {
		return &newChildren[0], true
	}
	return NewNode(newChildren), true
}

func merge(src, dest []Node) []Node {
	if len(src) == 0 || len(dest) == 0 {
		return nil
	}

	if len(src) == 1 && src[0].IsLeaf() && len(dest) == 1 && dest[0].IsLeaf() {
		if src[0].value != dest[0].value {
			return nil
		}
		return []Node{*NewLeaf(src[0].value)}
	}

	if len(src) == 1 && !src[0].IsLeaf() && len(dest) == 1 && !dest[0].IsLeaf() {
		return []Node{*NewNode(merge(src[0].children, dest[0].children))}
	}
	if len(src) == 1 && !src[0].IsLeaf() {
		return []Node{*NewNode(merge(src[0].children, dest))}
	}
	if len(dest) == 1 && !dest[0].IsLeaf() {
		return []Node{*NewNode(merge(src, dest[0].children))}
	}

	for i := 1; i < len(src); i++ {
		for j := 1; j < len(dest); j++ {
			if elems(src[:i]) != elems(dest[:j]) {
				continue
			}
			return append(merge(src[:i], dest[:j]), merge(src[i:], dest[j:])...)
		}
	}

	return nil
}

func elems(nodes []Node) int {
	r := 0
	for i := range nodes {
		r += nodes[i].elems
	}
	return r
}

func flatCompare(src, dest []Node) bool {
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

func flatten(ns []Node, result []string) []string {
	i := 0
	for j := range ns {
		if ns[j].IsLeaf() {
			result[i] = ns[j].value
			i++
		} else {
			flatten(ns[j].children, result[i:i+ns[j].elems])
			i += ns[j].elems
		}
	}
	return result
}
