package log

import (
	"bufio"
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
	var entry LogEntry
	err := ltsv.DefaultParser.ParseLine(r.r.Bytes(), func(label []byte, value []byte) error {
		l := string(label)
		v := string(value)

		switch l {
		case "req":
			entry.Req = v
			method, uri := parseReq(v, r.matchingPatterns)
			entry.Method = method
			entry.Uri = uri
			isIgnored := isIgnored(v, r.ignorePatterns)
			entry.IsIgnored = isIgnored

		case "status":
			status, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			entry.Status = status

		case "time":
			reqTime, err := time.Parse(r.timeFormat, v)
			if err != nil {
				return err
			}
			entry.Time = reqTime

		case "uidset":
			if v != "" && v != "-" {
				if i := strings.Index(v, "="); i >= 0 {
					entry.Uid = v[i+1:]
				} else {
					entry.Uid = v
				}
				entry.SetNewUid = true
			}

		case "uidgot":
			if v != "" && v != "-" {
				if i := strings.Index(v, "="); i >= 0 {
					entry.Uid = v[i+1:]
				} else {
					entry.Uid = v
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
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
