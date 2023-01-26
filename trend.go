package stool

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Wing924/ltsv"
	"github.com/samber/lo"
)

type TrendProfiler struct {
}

func NewTrendProfiler() *TrendProfiler {
	return &TrendProfiler{}
}

func (p *TrendProfiler) Profile(in io.Reader, opt TrendOption) (*Trend, error) {
	var patterns []*regexp.Regexp
	patterns = make([]*regexp.Regexp, len(opt.MatchingGroups))
	for i, mg := range opt.MatchingGroups {
		p, err := regexp.Compile(mg)
		if err != nil {
			return nil, err
		}
		patterns[i] = p
	}

	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	var result = map[string]map[int]int{}
	var startTime time.Time
	step := 0
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

		k := key(row["req"], patterns)
		if result[k] == nil {
			result[k] = map[int]int{}
		}

		if startTime.IsZero() {
			startTime = reqTime
		}
		t := int(reqTime.Sub(startTime).Seconds()) / opt.Interval
		step = t + 1

		result[k][t] += 1
	}

	res := NewTrend(result, opt.Interval, step)
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

func key(req string, patterns []*regexp.Regexp) string {
	splitted := strings.Split(req, " ")
	method := splitted[0]
	uri := strings.Split(splitted[1], "?")[0]

	for _, p := range patterns {
		if p.MatchString(uri) {
			uri = p.String()
		}
	}
	return fmt.Sprintf("%s %s", method, uri)
}
