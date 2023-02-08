package internal

import (
	"golang.org/x/exp/slices"
)

type Node struct {
	value    string
	children []Node
	leaves   int
}

func NewRoot() Node {
	return Node{children: []Node{}}
}

func NewNode(children []Node) *Node {
	return &Node{children: children, leaves: leaves(children)}
}

func NewLeaf(value string) *Node {
	return &Node{value: value, leaves: 1}
}

func (n *Node) Value() string {
	return n.value
}

func (n *Node) Children() []Node {
	return n.children
}

func (n *Node) Child(i int) *Node {
	return &n.children[i]
}

func (n *Node) Degree() int {
	return len(n.children)
}

func (n *Node) Leaves() int {
	return n.leaves
}

func (n *Node) First() *Node {
	if n.IsLeaf() {
		return n
	}
	return n.children[0].First()
}

func (n *Node) Last() *Node {
	if n.IsLeaf() {
		return n
	}
	return n.children[len(n.children)-1].Last()
}

func (n *Node) IsLeaf() bool {
	return n.children == nil || (len(n.children) == 0 && n.value != "")
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
		head := n.children[:i+1]
		tail := n.children[i+1:]

		if leaves(head) < leaves(tail) {
			break
		}
		if n.children[i].Last().value != value {
			continue
		}

		n.children = n.children[:i+1]
		if n.append(tail) {
			n.leaves = leaves(n.children)
			return
		} else {
			n.children = append(n.children, tail...)
		}
	}

	n.leaves += 1
}

func (n *Node) append(nodes []Node) bool {
	if n.Last().value != nodes[len(nodes)-1].Last().value {
		return false
	}

	if n.children[len(n.children)-1].leaves > leaves(nodes) {
		return n.children[len(n.children)-1].append(nodes)
	}

	for i := len(n.children) - 1; i >= 0; i-- {
		if leaves(n.children[i:]) == leaves(nodes) {
			node, ok := Merge(n.children[i:], nodes)
			if ok {
				n.leaves -= leaves(n.children[i:])
				n.leaves += node.leaves
				n.children = n.children[:i+1]
				n.children[i] = *node
			}
			return ok
		} else if leaves(n.children[i:]) > leaves(nodes) {
			return false
		}
	}
	return false
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
			if leaves(src[:i]) != leaves(dest[:j]) {
				continue
			}
			return append(merge(src[:i], dest[:j]), merge(src[i:], dest[j:])...)
		}
	}

	return nil
}

func leaves(nodes []Node) int {
	r := 0
	for i := range nodes {
		r += nodes[i].leaves
	}
	return r
}

func flatCompare(src, dest []Node) bool {
	if src[0].First().value != dest[0].First().value {
		return false
	}
	srcSize := leaves(src)
	destSize := leaves(dest)
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
			flatten(ns[j].children, result[i:i+ns[j].leaves])
			i += ns[j].leaves
		}
	}
	return result
}
