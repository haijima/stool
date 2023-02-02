package stool

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/Wing924/ltsv"
	mapset "github.com/deckarep/golang-set/v2"
)

type TransitionProfiler struct {
}

func NewTransitionProfiler() *TransitionProfiler {
	return &TransitionProfiler{}
}

func (p *TransitionProfiler) Profile(in io.Reader, opt TransitionOption) (*Transition, error) {
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

	var result = map[string]map[string]int{}
	result[""] = map[string]int{}
	var lastVisit = map[string]string{}
	var sum = map[string]int{}
	endpoints := mapset.NewSet[string]()
	endpoints.Add("")
	for scanner.Scan() {
		row := make(map[string]string)
		row, err := ltsv.DefaultParser.ParseLineAsMap(scanner.Bytes(), row)
		if err != nil {
			return nil, err
		}

		uidSet := row["uidset"]
		uidGot := row["uidgot"]
		k := key(row["req"], patterns)
		if isIgnored(row["req"], ignorePatterns) {
			continue
		}

		endpoints.Add(k)
		sum[k] += 1

		if uidGot != "" && uidGot != "-" {
			// revisiting user
			if result[lastVisit[uidGot]] == nil {
				result[lastVisit[uidGot]] = map[string]int{}
			}
			result[lastVisit[uidGot]][k] += 1
			lastVisit[uidGot] = k
		} else if uidSet != "" && uidSet != "-" {
			// new user
			result[""][k] += 1
			lastVisit[uidSet] = k
		}
	}

	for _, lv := range lastVisit {
		if result[lv] == nil {
			result[lv] = map[string]int{}
		}
		result[lv][""] += 1
	}

	res := NewTransition(result, endpoints, sum)
	return res, nil
}

func isIgnored(req string, ignorePatterns []*regexp.Regexp) bool {
	splitted := strings.Split(req, " ")
	uri := strings.Split(splitted[1], "?")[0]

	for _, p := range ignorePatterns {
		if p.MatchString(uri) {
			return true
		}
	}
	return false
}

type TransitionOption struct {
	MatchingGroups []string
	IgnorePatterns []string
	TimeFormat     string
}

type Transition struct {
	Data      map[string]map[string]int
	Endpoints mapset.Set[string]
	Sum       map[string]int
}

func NewTransition(data map[string]map[string]int, endpoints mapset.Set[string], sum map[string]int) *Transition {
	return &Transition{
		Data:      data,
		Endpoints: endpoints,
		Sum:       sum,
	}
}
