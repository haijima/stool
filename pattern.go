package stool

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type node struct {
	children []node
	value    string
	root     bool
}

func NewRoot() *node {
	return &node{root: true}
}

func NewNode(children []node) *node {
	return &node{children: children}
}

func NewLeaf(value string) *node {
	return &node{value: value}
}

func (p *node) Children() []node {
	return p.children
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

	childrenStr := strings.Join(lo.Map(p.children, func(child node, i int) string {
		return child.String()
	}), " -> ")
	if p.root {
		return childrenStr
	}
	return fmt.Sprintf("(%s)*", childrenStr)
}

func (p *node) Append(value string) {
	l := len(p.children)
	for i := l - 1; i >= 0; i-- {
		if p.children[i].Last().value != value {
			continue
		}

		s := make([]node, l-i)
		copy(s, p.children[i+1:l])
		s[l-i-1] = *NewLeaf(value)

		for j := i; j >= 0; j-- {
			if newPattern, ok := merge(s, p.children[j:i+1]); ok {
				p.children = p.children[:j]
				p.appendChild(*newPattern)
				return
			} else if i == j && !p.children[i].IsLeaf() {
				ll := len(p.children[i].children)
				for k := ll - 1; k >= 0; k-- {
					if newPattern, ok := merge(s, p.children[i].children[k:]); ok {
						p.children[i].children = p.children[i].children[:k]
						p.children[i].appendChild(*newPattern)
						p.children = p.children[:i+1]
						return
					}
				}
			}
		}
	}

	if p.children == nil {
		p.children = make([]node, 0)
	}
	p.appendChild(*NewLeaf(value))
}

func (p *node) appendChild(n node) {
	if p.children == nil {
		p.children = make([]node, 0)
	}
	p.children = append(p.children, n)
}

func merge(src []node, dest []node) (*node, bool) {
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
		return &newChildren[0], true
	}
	return NewNode(newChildren), true
}

var (
	ErrEmpty              = errors.New("empty patterns cannot be merged")
	ErrDifferentLeaf      = errors.New("different value leaf patterns cannot be merged")
	ErrDifferentStructure = errors.New("different structure")
	ErrUnexpected         = fmt.Errorf("unexpected error in _merge")
)

func _merge(src []node, dest []node) ([]node, error) {
	if len(src) == 0 || len(dest) == 0 {
		return nil, ErrEmpty
	}

	if len(src) == 1 && src[0].IsLeaf() && len(dest) == 1 && dest[0].IsLeaf() {
		if src[0].value != dest[0].value {
			return nil, ErrDifferentLeaf
		}
		return []node{*NewLeaf(src[0].value)}, nil
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
		return []node{*NewNode(mergedChildren)}, nil
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

func flatten(ps []node) []string {
	result := make([]string, 0)
	for _, p := range ps {
		if p.IsLeaf() {
			result = append(result, p.value)
		} else {
			result = append(result, flatten(p.Children())...)
		}
	}
	return result
}

func flatCompare(src []node, dest []node) bool {
	return slices.Equal(flatten(src), flatten(dest))
}
