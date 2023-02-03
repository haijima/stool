package stool

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/Wing924/ltsv"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type ScenarioProfiler struct {
}

type ScenarioOption struct {
	MatchingGroups []string
	IgnorePatterns []string
	TimeFormat     string
}

func NewScenarioProfiler() *ScenarioProfiler {
	return &ScenarioProfiler{}
}

func (p *ScenarioProfiler) Profile(in io.Reader, opt ScenarioOption) (*Pattern, error) {
	patterns := make([]*regexp.Regexp, len(opt.MatchingGroups))
	for i, mg := range opt.MatchingGroups {
		p, err := regexp.Compile(mg)
		if err != nil {
			return nil, err
		}
		patterns[i] = p
	}
	ignorePatterns := make([]*regexp.Regexp, len(opt.IgnorePatterns))
	for i, ip := range opt.IgnorePatterns {
		p, err := regexp.Compile(ip)
		if err != nil {
			return nil, err
		}
		ignorePatterns[i] = p
	}

	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	var result = map[string]*Pattern{}
	endpoints := mapset.NewSet[string]()
	intToEndpoint := map[int]string{}
	endpointToInt := map[string]int{}
	firstCalls := map[string]int{}
	lastCalls := map[string]int{}
	i := 0
	var startTime time.Time
	for scanner.Scan() {
		row := make(map[string]string)
		row, err := ltsv.DefaultParser.ParseLineAsMap(scanner.Bytes(), row)
		if err != nil {
			return nil, err
		}

		reqTime, err := time.Parse(opt.TimeFormat, row["time"])
		if err != nil {
			return nil, err
		}
		if startTime.IsZero() {
			startTime = reqTime
		}
		reqTimeSec := int(reqTime.Sub(startTime).Seconds())
		uidSet := row["uidset"]
		uidGot := row["uidgot"]
		k := key(row["req"], patterns)
		if isIgnored(row["req"], ignorePatterns) {
			continue
		}

		added := endpoints.Add(k)
		if added {
			intToEndpoint[i] = k
			endpointToInt[k] = i
			i++
		}

		if uidGot != "" && uidGot != "-" {
			// revisiting user
			if _, ok := result[uidGot]; !ok {
				result[uidGot] = &Pattern{}
			}
			//result[uidGot].Append(k)
			result[uidGot].Append(fmt.Sprintf("%d", endpointToInt[k]))
			lastCalls[uidGot] = reqTimeSec
		} else if uidSet != "" && uidSet != "-" {
			// new user
			result[uidSet] = &Pattern{}
			//result[uidSet].Append(k)
			result[uidSet].Append(fmt.Sprintf("%d", endpointToInt[k]))
			firstCalls[uidSet] = reqTimeSec
			lastCalls[uidSet] = reqTimeSec
		}
	}

	fmt.Println("====================")
	ks := lo.Keys(intToEndpoint)
	slices.Sort(ks)
	for _, i := range ks {
		fmt.Printf("%2d %s\n", i, intToEndpoint[i])
	}
	fmt.Println("====================")

	fmt.Printf("uids: %d\n", len(result))

	type scenarioStruct struct {
		hash     string
		count    int
		firstReq int
		lastReq  int
		pattern  *Pattern
	}
	scenarios := map[string]scenarioStruct{}
	for uid, scenario := range result {
		if s, ok := scenarios[scenario.String()]; ok {
			s.count += 1
			if firstCalls[uid] < s.firstReq {
				s.firstReq = firstCalls[uid]
			}
			if s.lastReq < lastCalls[uid] {
				s.lastReq = lastCalls[uid]

			}
			scenarios[scenario.String()] = s
		} else {
			scenarios[scenario.String()] = scenarioStruct{
				hash:     scenario.String(),
				count:    1,
				firstReq: firstCalls[uid],
				lastReq:  lastCalls[uid],
				pattern:  scenario,
			}
		}
	}

	// merge patterns
	ss := lo.Values(scenarios)
	tt := make([]scenarioStruct, 0)
	for _, s := range ss {
		match := false
		for i, t := range tt {
			if p, ok := merge([]Pattern{*s.pattern}, []Pattern{*t.pattern}); ok {
				p.rep = s.pattern.rep || t.pattern.rep
				tt[i] = scenarioStruct{
					hash:     p.String(),
					count:    tt[i].count + s.count,
					firstReq: lo.Min([]int{tt[i].firstReq, s.firstReq}),
					lastReq:  lo.Max([]int{tt[i].lastReq, s.lastReq}),
					pattern:  p,
				}
				match = true
				break
			}
		}
		if !match {
			tt = append(tt, s)
		}
	}

	slices.SortFunc(tt, func(a, b scenarioStruct) bool {
		if a.firstReq != b.firstReq {
			return a.firstReq < b.firstReq
		}
		if a.count != b.count {
			return a.count > b.count
		}
		return a.pattern.String() > b.pattern.String()
	})
	fmt.Println("first call[s],last call[s],count,scenario Pattern")
	for _, s := range tt {
		fmt.Printf("%d,%d,%d,%s\n", s.firstReq, s.lastReq, s.count, s.pattern.String())
	}

	return nil, nil
}

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
