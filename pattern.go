package stool

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type node struct {
	children []*node
	value    string
	root     bool
	elems    int // stores only after elements() is called
}

func NewRoot() *node {
	return &node{root: true}
}

func NewNode(children []*node) *node {
	return &node{children: children}
}

func NewLeaf(value string) *node {
	return &node{value: value, elems: 1}
}

func (p *node) First() *node {
	if p.IsLeaf() {
		return p
	}
	return p.children[0].First()
}

func (p *node) Last() *node {
	if p.IsLeaf() {
		return p
	}
	return p.children[len(p.children)-1].Last()
}

func (p *node) IsLeaf() bool {
	return p.children == nil || len(p.children) == 0
}

func (p *node) String() string {
	if p.IsLeaf() {
		return p.value
	}

	childrenStrs := make([]string, len(p.children))
	for i, child := range p.children {
		childrenStrs[i] = child.String()
	}
	childrenStr := strings.Join(childrenStrs, " -> ")

	if p.root {
		return childrenStr
	}
	return "(" + childrenStr + ")*"
}

func (p *node) Append(value string) {
	elements([]*node{p}) // refresh

	l := len(p.children)
	for i := l - 1; i >= 0; i-- {
		if p.children[i].Last().value != value {
			continue
		}

		s := make([]*node, l-i)
		copy(s, p.children[i+1:l])
		s[l-i-1] = NewLeaf(value)

		for j := i; j >= 0; j-- {
			// Check if p.children[j:i+1] equals s
			if newPattern, ok := merge(p.children[j:i+1], s); ok {
				p.children = p.children[:j+1]
				p.children[j] = newPattern
				return
			}
			// Check if p.children[i]'s tail matches s
			sSize := 0
			for _, n := range s {
				sSize += n.elems
			}
			if i == j && !p.children[i].IsLeaf() && p.children[i].elems > sSize {
				ll := len(p.children[i].children)
				for k := ll - 1; k >= 0; k-- {
					if newPattern, ok := merge(p.children[i].children[k:], s); ok {
						p.children[i].children = p.children[i].children[:k+1]
						p.children[i].children[k] = newPattern
						p.children = p.children[:i+1]
						return
					}
				}
			}
		}
	}

	if p.children == nil {
		p.children = []*node{NewLeaf(value)}
		p.elems = 1
	} else {
		p.children = append(p.children, NewLeaf(value))
		p.elems += 1
	}
}

func merge(src []*node, dest []*node) (*node, bool) {
	if !flatCompare(src, dest) {
		return nil, false
	}

	newChildren, err := _merge(src, dest)
	if err != nil {
		if err == ErrUnexpected {
			zap.L().Info(err.Error())
		}
		return nil, false
	}

	if len(newChildren) == 1 && !newChildren[0].IsLeaf() {
		return newChildren[0], true
	}
	return NewNode(newChildren), true
}

var (
	ErrEmpty              = errors.New("empty patterns cannot be merged")
	ErrDifferentLeaf      = errors.New("different value leaf patterns cannot be merged")
	ErrDifferentStructure = errors.New("different structure")
	ErrUnexpected         = fmt.Errorf("unexpected error in _merge")
)

func _merge(src []*node, dest []*node) ([]*node, error) {
	if len(src) == 0 || len(dest) == 0 {
		return nil, ErrEmpty
	}

	if len(src) == 1 && src[0].IsLeaf() && len(dest) == 1 && dest[0].IsLeaf() {
		if src[0].value != dest[0].value {
			return nil, ErrDifferentLeaf
		}
		return []*node{NewLeaf(src[0].value)}, nil
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
		mergedChildren, err := _merge(s, d)
		if err != nil {
			return nil, err
		}
		return []*node{NewNode(mergedChildren)}, nil
	}

	for i := 1; i <= len(src); i++ {
		for j := 1; j <= len(dest); j++ {
			if i == len(src) && j == len(dest) {
				return nil, ErrDifferentStructure
			}
			if flatCompare(src[:i], dest[:j]) {
				head, err := _merge(src[:i], dest[:j])
				if err != nil {
					return nil, err
				}
				tail, err := _merge(src[i:], dest[j:])
				if err != nil {
					return nil, err
				}
				return append(head, tail...), nil
			}
		}
	}

	zap.L().Info(fmt.Sprintf("unexpected error in _merge src:%+v dest:%+v\n", src, dest))
	return nil, ErrUnexpected
}

func flatCompare(src []*node, dest []*node) bool {
	//l1 := len(src)
	//l2 := len(dest)
	//if l1 == 0 && l2 == 0 {
	//	return true
	//}
	//if l1 == 0 {
	//	return false
	//}
	//if l2 == 0 {
	//	return false
	//}
	if src[0].First().value != dest[0].First().value {
		return false
	}
	//if src[l1-1].Last().value != dest[l2-1].Last().value {
	//	return false
	//}
	srcSize := 0
	for _, s := range src {
		srcSize += s.elems
	}
	destSize := 0
	for _, d := range dest {
		destSize += d.elems
	}
	if srcSize != destSize {
		return false
	}

	return slices.Equal(flatten(src, srcSize), flatten(dest, destSize))
}

func flatten(ps []*node, size int) []string {
	result := make([]string, size)
	i := 0
	for _, p := range ps {
		if p.IsLeaf() {
			result[i] = p.value
			i++
		} else {
			for _, v := range flatten(p.children, p.elems) {
				result[i] = v
				i++
			}
		}
	}
	return result
}

func elements(ps []*node) int {
	l := 0
	for _, p := range ps {
		if p.IsLeaf() {
			l += 1
			p.elems = 1
		} else {
			p.elems = elements(p.children)
			l += p.elems
		}
	}
	return l
}
