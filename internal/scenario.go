package internal

import (
	"bufio"
	"io"
	"regexp"
	"time"

	"github.com/Wing924/ltsv"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

type ScenarioProfiler struct {
}

type ScenarioOption struct {
	MatchingGroups []string
	IgnorePatterns []string
	TimeFormat     string
}

type ScenarioStruct struct {
	Hash     string
	Count    int
	FirstReq int
	LastReq  int
	Pattern  *Node
}

func NewScenarioProfiler() *ScenarioProfiler {
	return &ScenarioProfiler{}
}

func (p *ScenarioProfiler) Profile(in io.Reader, opt ScenarioOption) ([]ScenarioStruct, error) {
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

	var result = map[string]*Node{}
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
				result[uidGot] = &Node{}
			}
			result[uidGot].Append(k)
			lastCalls[uidGot] = reqTimeSec
		} else if uidSet != "" && uidSet != "-" {
			// new user
			result[uidSet] = &Node{}
			result[uidSet].Append(k)
			firstCalls[uidSet] = reqTimeSec
			lastCalls[uidSet] = reqTimeSec
		}
	}

	scenarios := map[string]ScenarioStruct{}
	for uid, scenario := range result {
		if s, ok := scenarios[scenario.String(true)]; ok {
			s.Count += 1
			if firstCalls[uid] < s.FirstReq {
				s.FirstReq = firstCalls[uid]
			}
			if s.LastReq < lastCalls[uid] {
				s.LastReq = lastCalls[uid]
			}
			scenarios[scenario.String(true)] = s
		} else {
			scenarios[scenario.String(true)] = ScenarioStruct{
				Hash:     scenario.String(true),
				Count:    1,
				FirstReq: firstCalls[uid],
				LastReq:  lastCalls[uid],
				Pattern:  scenario,
			}
		}
	}

	// merge patterns
	ss := lo.Values(scenarios)
	tt := make([]ScenarioStruct, 0)
	for _, s := range ss {
		match := false
		for i, t := range tt {
			if p, ok := Merge([]Node{*s.Pattern}, []Node{*t.Pattern}); ok {
				tt[i] = ScenarioStruct{
					Hash:     p.String(true),
					Count:    s.Count + t.Count,
					FirstReq: lo.Min([]int{s.FirstReq, t.FirstReq}),
					LastReq:  lo.Max([]int{s.LastReq, t.LastReq}),
					Pattern:  p,
				}
				match = true
				break
			}
		}
		if !match {
			tt = append(tt, s)
		}
	}

	slices.SortFunc(tt, func(a, b ScenarioStruct) bool {
		if a.FirstReq != b.FirstReq {
			return a.FirstReq < b.FirstReq
		}
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Pattern.String(true) > b.Pattern.String(true)
	})

	return tt, nil
}
