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

func (p *node) first() *node {
	if p.IsLeaf() {
		return p
	}
	return p.children[0].first()
}

func (p *node) last() *node {
	if p.IsLeaf() {
		return p
	}
	return p.children[len(p.children)-1].last()
}

func (p *node) IsLeaf() bool {
	return p.children == nil || len(p.children) == 0
}

func (p *node) String(root bool) string {
	if p.IsLeaf() {
		return p.value
	}

	childrenStr := ""
	for i, child := range p.children {
		if i == 0 {
			childrenStr += child.String(false)
		} else {
			childrenStr += " -> " + child.String(false)
		}
	}

	if root {
		return childrenStr
	}
	return "(" + childrenStr + ")*"
}

func (p *node) Append(value string) {
	l := len(p.children)
	for i := l - 1; i >= 0; i-- {
		if p.children[i].last().value != value {
			continue
		}

		s := append(p.children[i+1:l], *newLeaf(value))

		for j := i; j >= 0; j-- {
			// Check if p.children[j:i+1] equals s
			if mergedNode, ok := merge(p.children[j:i+1], s); ok {
				p.elems -= elems(p.children[j:])
				p.elems += mergedNode.elems
				p.children = p.children[:j+1]
				p.children[j] = *mergedNode
				return
			}

			// Check if p.children[i]'s tail matches s
			if i == j && p.children[i].elems > elems(s) {
				ll := len(p.children[i].children)
				for k := ll - 1; k >= 0; k-- {
					if mergedNode, ok := merge(p.children[i].children[k:], s); ok {
						p.children[i].elems -= elems(p.children[i].children[k:])
						p.children[i].elems += mergedNode.elems
						p.children[i].children = p.children[i].children[:k+1]
						p.children[i].children[k] = *mergedNode
						p.elems -= elems(p.children[i+1:])
						p.children = p.children[:i+1]
						return
					}
				}
			}

		}
	}

	if p.children == nil {
		p.children = []node{*newLeaf(value)}
		p.elems = 1
	} else {
		p.children = append(p.children, *newLeaf(value))
		p.elems += 1
	}
}

func merge(src []node, dest []node) (*node, bool) {
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

func _merge(src []node, dest []node) []node {
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

func flatCompare(src []node, dest []node) bool {
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

func flatten(ps []node, result []string) []string {
	i := 0
	for _, p := range ps {
		if p.IsLeaf() {
			result[i] = p.value
			i++
		} else {
			for _, v := range flatten(p.children, result[i:i+p.elems]) {
				result[i] = v
				i++
			}
		}
	}
	return result
}
