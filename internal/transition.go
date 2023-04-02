package internal

import (
	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/util"
)

type TransitionProfiler struct {
}

func NewTransitionProfiler() *TransitionProfiler {
	return &TransitionProfiler{}
}

func (p *TransitionProfiler) Profile(reader *log.LTSVReader) (*Transition, error) {
	var result = map[string]map[string]int{}
	result[""] = map[string]int{}
	var lastVisit = map[string]string{}
	var sum = map[string]int{}
	endpoints := map[string]struct{}{}
	endpoints[""] = struct{}{}

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

		endpoints[k] = struct{}{}
		sum[k] += 1

		if entry.Uid != "" {
			lv := lastVisit[entry.Uid]
			if result[lv] == nil {
				result[lv] = map[string]int{}
			}
			result[lv][k] += 1
			lastVisit[entry.Uid] = k
		}
	}

	for _, lv := range lastVisit {
		if result[lv] == nil {
			result[lv] = map[string]int{}
		}
		result[lv][""] += 1
	}

	res := NewTransition(result, util.Keys(endpoints), sum)
	return res, nil
}

type TransitionOption struct {
	MatchingGroups []string
	TimeFormat     string
}

type Transition struct {
	Data      map[string]map[string]int
	Endpoints []string
	Sum       map[string]int
}

func NewTransition(data map[string]map[string]int, endpoints []string, sum map[string]int) *Transition {
	return &Transition{
		Data:      data,
		Endpoints: endpoints,
		Sum:       sum,
	}
}
