package internal

import (
	"sort"
	"strings"
	"time"

	"github.com/haijima/stool/internal/log"
	"golang.org/x/exp/maps"
)

type TrendProfiler struct {
}

func NewTrendProfiler() *TrendProfiler {
	return &TrendProfiler{}
}

func (p *TrendProfiler) Profile(reader *log.LTSVReader, interval int, sortKeys []string) (*Trend, error) {
	var result = map[string]*TrendData{}
	var startTime time.Time
	step := 0

	var entry log.LogEntry
	for reader.Read() {
		_, err := reader.Parse(&entry)
		if err != nil {
			if err == log.Filtered {
				continue
			}
			return nil, err
		}

		if startTime.IsZero() {
			startTime = entry.Time
		}
		t := int(entry.Time.Sub(startTime).Seconds()) / interval
		if t+1 > step {
			step = t + 1
			for _, v := range result {
				if len(v.counts) < step {
					v.counts = append(v.counts, make([]int, step-len(v.counts))...)
				}
			}
		}
		k := entry.Key()
		if result[k] == nil {
			result[k] = &TrendData{counts: make([]int, step)}
		}
		result[k].counts[t] += 1
		result[k].sum += 1
	}

	res := NewTrend(result, interval, step)
	res.sort(sortKeys)
	return res, nil
}

type TrendOption struct {
	MatchingGroups []string
	TimeFormat     string
	Interval       int
}

type Trend struct {
	data     map[string]*TrendData
	Interval int
	Step     int
	keys     []string
	sorted   bool
}

type TrendData struct {
	Method string
	Uri    string
	counts []int
	sum    int
}

func (t *TrendData) AddCount(index int, count int) {
	if len(t.counts) <= index {
		t.counts = append(t.counts, make([]int, index-len(t.counts)+1)...)
	}
	t.counts[index] += count
	t.sum += count
}

func NewTrend(data map[string]*TrendData, interval, step int) *Trend {
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
	return m.counts
}

func (t *Trend) Endpoints() []string {
	if t.sorted {
		return t.keys
	}
	t.sort([]string{})
	return t.keys
}

func (t *Trend) sort(sortOptions []string) []string {
	if t.sorted {
		return t.keys
	}

	s := NewSortable(maps.Keys(t.data))
	s.AddMapper("method", func(i, j string) bool { return t.data[i].Method < t.data[j].Method })
	s.AddMapper("uri", func(i, j string) bool { return t.data[i].Uri < t.data[j].Uri })
	s.AddMapper("sum", func(i, j string) bool { return t.data[i].sum < t.data[j].sum })
	s.AddMapper("count0", func(i, j string) bool { return t.data[i].counts[0] < t.data[j].counts[0] })
	s.AddMapper("count1", func(i, j string) bool { return t.data[i].counts[1] < t.data[j].counts[1] })
	s.AddMapper("countN", func(i, j string) bool {
		return t.data[i].counts[len(t.data[i].counts)-1] < t.data[j].counts[len(t.data[j].counts)-1]
	})

	if len(sortOptions) == 0 {
		sortOptions = []string{"sum:desc"}
	}
	sortKeys := make([]string, 0, len(sortOptions))
	sortOrders := make([]SortOrder, 0, len(sortOptions))
	for _, k := range sortOptions {
		split := strings.Split(strings.TrimSpace(strings.ToLower(k)), ":")
		key := split[0]
		switch key {
		case "method", "uri", "sum", "count0", "count1", "countN":
		default:
			continue
		}
		order := Asc
		if len(split) > 1 {
			if split[1] == "asc" {
				order = Asc
			} else if split[1] == "desc" {
				order = Desc
			} else {
				continue
			}
		}
		sortKeys = append(sortKeys, key)
		sortOrders = append(sortOrders, order)
	}
	s.MustSetSortOption(sortKeys, sortOrders)

	sort.Sort(s)
	sortedKeys := s.Values()
	t.sorted = true
	t.keys = sortedKeys
	return sortedKeys
}
