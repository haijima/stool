package stool

import (
	"bufio"
	"fmt"
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

func NewScenarioProfiler() *ScenarioProfiler {
	return &ScenarioProfiler{}
}

func (p *ScenarioProfiler) Profile(in io.Reader, opt ScenarioOption) (*node, error) {
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

	var result = map[string]*node{}
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
				result[uidGot] = &node{}
			}
			//result[uidGot].Append(k)
			result[uidGot].Append(fmt.Sprintf("%d", endpointToInt[k]))
			lastCalls[uidGot] = reqTimeSec
		} else if uidSet != "" && uidSet != "-" {
			// new user
			result[uidSet] = &node{}
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
		pattern  *node
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
			if p, ok := merge([]node{*s.pattern}, []node{*t.pattern}); ok {
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
	fmt.Println("first call[s],last call[s],count,scenario node")
	for _, s := range tt {
		fmt.Printf("%d,%d,%d,%s\n", s.firstReq, s.lastReq, s.count, s.pattern.String())
	}

	return nil, nil
}
