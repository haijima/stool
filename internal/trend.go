package internal

import (
	"github.com/haijima/stool/internal/log"
	"sort"
	"time"

	"github.com/samber/lo"
)

type TrendProfiler struct {
}

func NewTrendProfiler() *TrendProfiler {
	return &TrendProfiler{}
}

func (p *TrendProfiler) Profile(reader *log.LTSVReader, interval int) (*Trend, error) {
	var result = map[string]map[int]int{}
	var startTime time.Time
	step := 0

	var entry log.LogEntry
	for reader.Read() {
		_, err := reader.Parse(&entry)
		if err != nil {
			return nil, err
		}

		k := entry.Key()
		if result[k] == nil {
			result[k] = map[int]int{}
		}

		if startTime.IsZero() {
			startTime = entry.Time
		}
		t := int(entry.Time.Sub(startTime).Seconds()) / interval
		step = t + 1

		result[k][t] += 1
	}

	res := NewTrend(result, interval, step)
	return res, nil
}

type TrendOption struct {
	MatchingGroups []string
	TimeFormat     string
	Interval       int
}

type Trend struct {
	data     map[string]map[int]int
	Interval int
	Step     int
}

func NewTrend(data map[string]map[int]int, interval, step int) *Trend {
	return &Trend{
		data:     data,
		Interval: interval,
		Step:     step,
	}
}

func (t *Trend) Counts(endpoint string) []int {
	m, ok := t.data[endpoint]
	if !ok {
		return []int{}
	}

	return lo.Times(t.Step, func(i int) int {
		return m[i] // if not contained stores zero
	})
}

func (t *Trend) Endpoints() []string {
	keys := lo.Keys(t.data)
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
