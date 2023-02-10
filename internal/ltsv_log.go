package internal

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Wing924/ltsv"
)

type LTSVReader struct {
	r                *bufio.Scanner
	timeFormat       string
	matchingPatterns []regexp.Regexp
	ignorePatterns   []regexp.Regexp
}

type LTSVReadOpt struct {
	MatchingGroups []string
	IgnorePatterns []string
	TimeFormat     string
}

const defaultTimeFormat = "02/Jan/2006:15:04:05 -0700"

func NewLTSVReader(r io.Reader, opt LTSVReadOpt) (*LTSVReader, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	timeFormat := opt.TimeFormat
	if opt.TimeFormat == "" {
		timeFormat = defaultTimeFormat
	}

	matchingregexps := make([]regexp.Regexp, 0, len(opt.MatchingGroups))
	for _, pattern := range opt.MatchingGroups {
		p, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		matchingregexps = append(matchingregexps, *p)
	}
	ignoreRegexps := make([]regexp.Regexp, 0, len(opt.IgnorePatterns))
	for _, pattern := range opt.IgnorePatterns {
		p, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		ignoreRegexps = append(ignoreRegexps, *p)
	}

	return &LTSVReader{
		r:                scanner,
		matchingPatterns: matchingregexps,
		ignorePatterns:   ignoreRegexps,
		timeFormat:       timeFormat,
	}, nil
}

type LogEntry struct {
	Req       string
	Method    string
	Uri       string
	Status    int
	Uid       string
	SetNewUid bool
	Time      time.Time
	IsIgnored bool
}

func (e LogEntry) Key() string {
	return e.Method + " " + e.Uri
}

func (r *LTSVReader) Read() bool {
	return r.r.Scan()
}

func (r *LTSVReader) Parse() (*LogEntry, error) {
	row := make(map[string]string)
	row, err := ltsv.DefaultParser.ParseLineAsMap(r.r.Bytes(), row)
	if err != nil {
		return nil, err
	}

	var entry LogEntry

	req, ok := row["req"]
	if !ok {
		return nil, fmt.Errorf("\"req\" field is not found on each log entry")
	}
	entry.Req = req
	method, uri := parseReq(req, r.matchingPatterns)
	entry.Method = method
	entry.Uri = uri
	isIgnored := isIgnored(req, r.ignorePatterns)
	entry.IsIgnored = isIgnored
	status, ok := row["status"]
	if !ok {
		return nil, fmt.Errorf("\"status\" field is not found on each log entry")
	}
	statusInt, err := strconv.Atoi(status)
	if err != nil {
		return nil, err
	}
	entry.Status = statusInt
	t, ok := row["time"]
	if !ok {
		return nil, fmt.Errorf("\"time\" field is not found on each log entry")
	}
	reqTime, err := time.Parse(r.timeFormat, t)
	if err != nil {
		return nil, err
	}
	entry.Time = reqTime
	uidSet, ok := row["uidset"]
	if !ok {
		return nil, fmt.Errorf("\"uidset\" field is not found on each log entry")
	}
	if uidSet != "" && uidSet != "-" {
		entry.Uid = strings.Split(uidSet, "=")[1]
		entry.SetNewUid = true
	}
	uidGot, ok := row["uidgot"]
	if !ok {
		return nil, fmt.Errorf("\"uidgot\" field is not found on each log entry")
	}
	if entry.Uid == "" && uidGot != "" && uidGot != "-" {
		entry.Uid = strings.Split(uidGot, "=")[1]
	}

	return &entry, nil
}

func parseReq(req string, patterns []regexp.Regexp) (string, string) {
	splitted := strings.Split(req, " ")
	method := splitted[0]
	uri := strings.Split(splitted[1], "?")[0]

	for _, p := range patterns {
		if p.MatchString(uri) {
			uri = p.String()
			return method, uri
		}
	}
	return method, uri
}

func isIgnored(req string, ignorePatterns []regexp.Regexp) bool {
	splitted := strings.Split(req, " ")
	uri := strings.Split(splitted[1], "?")[0]

	for _, p := range ignorePatterns {
		if p.MatchString(uri) {
			return true
		}
	}
	return false
}
