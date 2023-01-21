package stool

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Wing924/ltsv"
	"github.com/samber/lo"
)

type TrendOption struct {
	MatchingGroups []string
	File           string
	TimeFormat     string
	Interval       int
}

type TrendResult struct {
	data     map[string]map[int]int
	interval int
	Start    time.Time
	End      time.Time
}

func NewTrendResult(data map[string]map[int]int, interval int, start, end time.Time) TrendResult {
	return TrendResult{
		data:     data,
		interval: interval,
		Start:    start,
		End:      end,
	}
}

func (t *TrendResult) Counts(endpoint string) []int {
	m, ok := t.data[endpoint]
	if !ok {
		return []int{}
	}

	return lo.Times(t.Step(), func(i int) int {
		return m[i] // if not contained stores zero
	})
}

func (t *TrendResult) Endpoints() []string {
	return lo.Keys(t.data)
}

func (t *TrendResult) Step() int {
	return int(t.End.Sub(t.Start).Seconds())/t.interval + 1
}

func (t *TrendResult) Interval() int {
	return t.interval
}

func Trend(opt TrendOption) (TrendResult, error) {
	var patterns []*regexp.Regexp
	patterns = make([]*regexp.Regexp, len(opt.MatchingGroups))
	for i, mg := range opt.MatchingGroups {
		p, err := regexp.Compile(mg)
		if err != nil {
			return TrendResult{}, err
		}
		patterns[i] = p
	}

	f, err := os.Open(opt.File)
	if err != nil {
		return TrendResult{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var result = map[string]map[int]int{}
	var startTime time.Time
	var endTime time.Time
	for scanner.Scan() {
		row := make(map[string]string)
		row, err = ltsv.DefaultParser.ParseLineAsMap(scanner.Bytes(), row)
		if err != nil {
			return TrendResult{}, err
		}

		reqTime, err := time.Parse(opt.TimeFormat, row["time"])
		if err != nil {
			return TrendResult{}, err
		}

		if startTime.IsZero() {
			startTime = reqTime
		}
		endTime = reqTime

		k := key(row["req"], patterns)
		if result[k] == nil {
			result[k] = map[int]int{}
		}
		t := int(reqTime.Sub(startTime).Seconds()) / opt.Interval
		result[k][t] += 1
	}

	res := NewTrendResult(result, opt.Interval, startTime, endTime)
	return res, nil
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
