package stool

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type Pattern struct {
	children []Pattern
	endpoint string
	rep      bool
}

func (p *Pattern) String() string {
	if p.IsEmpty() {
		if p.rep {
			return fmt.Sprintf("%s *", p.endpoint)
		}
		return p.endpoint
	}

	if len(p.children) == 1 {
		return p.children[0].String()
	}

	childrenStr := strings.Join(lo.Map(p.children, func(child Pattern, i int) string {
		return child.String()
	}), " -> ")
	if p.rep {
		return fmt.Sprintf("(%s)*", childrenStr)
	}
	return childrenStr
}

func (p *Pattern) IsEmpty() bool {
	return p.children == nil || len(p.children) == 0
}

func (p *Pattern) LastPattern() *Pattern {
	if p.IsEmpty() {
		return p
	}
	return p.children[len(p.children)-1].LastPattern()
}

func (p *Pattern) Append(endpoint string) {
	l := len(p.children)
	for i := l - 1; i >= 0; i-- {
		if p.children[i].LastPattern().endpoint != endpoint {
			continue
		}
		s := make([]Pattern, l-i)
		copy(s, p.children[i+1:l])
		s[l-i-1] = Pattern{endpoint: endpoint}

		for j := i; j >= 0; j-- {
			if newPattern, ok := merge(s, p.children[j:i+1]); ok {
				p.children = p.children[:j]
				p.children = append(p.children, *newPattern)
				return
			} else if i == j && !p.children[i].IsEmpty() {
				ll := len(p.children[i].children)
				for k := ll - 1; k >= 0; k-- {
					if newPattern, ok := merge(s, p.children[i].children[k:]); ok {
						if k == ll-1 {
							p.children[i].children[ll-1].rep = true
						} else {
							p.children[i].children = p.children[i].children[:k]
							p.children[i].children = append(p.children[i].children, *newPattern)
						}
						p.children = p.children[:i+1]
						return
					}
				}
			}
		}
	}

	if p.children == nil {
		p.children = make([]Pattern, 0)
	}
	newPattern := Pattern{endpoint: endpoint}
	p.children = append(p.children, newPattern)
}

func merge(src []Pattern, dest []Pattern) (*Pattern, bool) {
	if !flatCompare(src, dest) {
		return nil, false
	}

	newChildren, err := _merge(src, dest)
	if err != nil {
		zap.L().Info(err.Error())
		return nil, false
	}

	if len(newChildren) == 1 {
		newChildren[0].rep = true
		return &newChildren[0], true
	}

	newPattern := Pattern{
		children: newChildren,
		rep:      true,
	}

	return &newPattern, true
}

func _merge(src []Pattern, dest []Pattern) ([]Pattern, error) {
	if len(src) == 0 || len(dest) == 0 {
		return nil, errors.New("empty patterns cannot be merged")
	}

	if len(src) == 1 && src[0].IsEmpty() && len(dest) == 1 && dest[0].IsEmpty() {
		if src[0].endpoint != dest[0].endpoint {
			return nil, errors.New("different endpoint leaf patterns cannot be merged")
		}
		return []Pattern{{
			endpoint: src[0].endpoint,
			rep:      src[0].rep || dest[0].rep,
		}}, nil
	}

	if len(src) == 1 && !src[0].IsEmpty() && len(dest) == 1 && !dest[0].IsEmpty() {
		mergedChildren, err := _merge(src[0].children, dest[0].children)
		if err != nil {
			return nil, err
		}
		return []Pattern{{
			children: mergedChildren,
			rep:      src[0].rep || dest[0].rep,
		}}, nil
	}

	if len(src) == 1 && !src[0].IsEmpty() {
		mergedChildren, err := _merge(src[0].children, dest)
		if err != nil {
			return nil, err
		}
		return []Pattern{{
			children: mergedChildren,
			rep:      true,
		}}, nil
	}
	if len(dest) == 1 && !dest[0].IsEmpty() {
		mergedChildren, err := _merge(src, dest[0].children)
		if err != nil {
			return nil, err
		}
		return []Pattern{{
			children: mergedChildren,
			rep:      true,
		}}, nil
	}

	for i := 1; i <= len(src); i++ {
		for j := 1; j <= len(dest); j++ {
			if i == len(src) && j == len(dest) {
				return nil, errors.New("different structure")
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

	return nil, fmt.Errorf("unexpected error in _merge src:%+v dest:%+v\n", src, dest)

}

func flatten(ps []Pattern) []string {
	result := make([]string, 0)
	for _, p := range ps {
		if p.IsEmpty() {
			result = append(result, p.endpoint)
		} else {
			result = append(result, flatten(p.children)...)
		}
	}
	return result
}

func flatCompare(src []Pattern, dest []Pattern) bool {
	return slices.Equal(flatten(src), flatten(dest))
}
