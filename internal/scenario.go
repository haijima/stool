package internal

import (
	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/pattern"
	"time"

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
	Pattern  *pattern.Node
}

func NewScenarioProfiler() *ScenarioProfiler {
	return &ScenarioProfiler{}
}

func (p *ScenarioProfiler) Profile(reader *log.LTSVReader) ([]ScenarioStruct, error) {
	var result = map[string]*pattern.Node{}
	endpoints := map[string]struct{}{}
	intToEndpoint := map[int]string{}
	endpointToInt := map[string]int{}
	firstCalls := map[string]int{}
	lastCalls := map[string]int{}

	i := 0
	var startTime time.Time
	for reader.Read() {
		entry, err := reader.Parse()
		if err != nil {
			return nil, err
		}
		if entry.IsIgnored {
			continue
		}

		k := entry.Key()

		if startTime.IsZero() {
			startTime = entry.Time
		}
		reqTimeSec := int(entry.Time.Sub(startTime).Seconds())

		_, exists := endpoints[k]
		if !exists {
			endpoints[k] = struct{}{}
			intToEndpoint[i] = k
			endpointToInt[k] = i
			i++
		}

		if entry.Uid != "" {
			if entry.SetNewUid {
				result[entry.Uid] = &pattern.Node{}
				firstCalls[entry.Uid] = reqTimeSec
			}
			result[entry.Uid].Append(k)
			lastCalls[entry.Uid] = reqTimeSec
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
	slices.SortFunc(ss, func(a, b ScenarioStruct) bool {
		return a.Pattern.String(true) > b.Pattern.String(true)
	})
	tt := make([]ScenarioStruct, 0)
	for _, s := range ss {
		match := false
		for i, t := range tt {
			if p, ok := pattern.Merge([]pattern.Node{*s.Pattern}, []pattern.Node{*t.Pattern}); ok {
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
