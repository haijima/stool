package internal

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/pattern"
	"golang.org/x/exp/maps"
)

type ScenarioProfiler struct {
}

type ScenarioOption struct {
	MatchingGroups []string
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
	var endTime time.Time
	var entry log.LogEntry
	for reader.Read() {
		_, err := reader.Parse(&entry)
		if err != nil {
			if err == log.Filtered {
				continue
			}
			return nil, err
		}

		k := entry.Key()

		if startTime.IsZero() {
			startTime = entry.Time
		}
		endTime = entry.Time
		reqTimeSec := int(entry.Time.Sub(startTime).Milliseconds())

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
	period := endTime.Sub(startTime).Milliseconds()
	latestFirstCall := int64(float64(period) * 0.95)
	for uid, firstCall := range firstCalls {
		if firstCall > int(latestFirstCall) {
			delete(result, uid)
			delete(firstCalls, uid)
			delete(lastCalls, uid)
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
	ss := maps.Values(scenarios)
	slices.SortFunc(ss, func(a, b ScenarioStruct) int {
		return strings.Compare(b.Pattern.String(true), a.Pattern.String(true))
	})
	tt := make([]ScenarioStruct, 0)
	for _, s := range ss {
		match := false
		for i, t := range tt {
			if p, ok := pattern.Merge([]pattern.Node{*s.Pattern}, []pattern.Node{*t.Pattern}); ok {
				tt[i] = ScenarioStruct{
					Hash:     p.String(true),
					Count:    s.Count + t.Count,
					FirstReq: min(s.FirstReq, t.FirstReq),
					LastReq:  max(s.LastReq, t.LastReq),
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

	slices.SortFunc(tt, func(a, b ScenarioStruct) int {
		if a.FirstReq != b.FirstReq {
			return cmp.Compare(a.FirstReq, b.FirstReq)
		}
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		return strings.Compare(b.Pattern.String(true), a.Pattern.String(true))
	})

	return tt, nil
}
