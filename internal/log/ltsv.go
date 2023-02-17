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

// Parse parses one line of log file into LogEntry struct
// For reducing memory allocation, you can pass a LogEntry to record to reuse the given one.
func (r *LTSVReader) Parse(entry *LogEntry) (*LogEntry, error) {
	if entry == nil {
		entry = &LogEntry{}
	}
	entry.Req = ""
	entry.Method = ""
	entry.Uri = ""
	entry.IsIgnored = false
	entry.Status = 0
	entry.Time = time.Time{}
	entry.Uid = ""
	entry.SetNewUid = false

	err := ltsv.DefaultParser.ParseLine(r.r.Bytes(), func(label, value []byte) error {
		switch string(label) {
		case "req":
			entry.Req = string(value)
			method, uri := parseReq(string(value), r.matchingPatterns)
			entry.Method = method
			entry.Uri = uri
			entry.IsIgnored = isIgnored(string(value), r.ignorePatterns)

		case "status":
			status, err := strconv.Atoi(string(value))
			if err != nil {
				return err
			}
			entry.Status = status

		case "time":
			reqTime, err := time.Parse(r.timeFormat, string(value))
			if err != nil {
				return err
			}
			entry.Time = reqTime

		case "uidset":
			if string(value) != "" && string(value) != "-" {
				if i := strings.Index(string(value), "="); i >= 0 {
					entry.Uid = string(value)[i+1:]
				} else {
					entry.Uid = string(value)
				}
				entry.SetNewUid = true
			}

		case "uidgot":
			if string(value) != "" && string(value) != "-" {
				if i := strings.Index(string(value), "="); i >= 0 {
					entry.Uid = string(value)[i+1:]
				} else {
					entry.Uid = string(value)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func parseReq(req string, patterns []regexp.Regexp) (string, string) {
	var method string
	var uri string
	i := strings.Index(req, " ")
	if i >= 0 {
		method = req[:i]
		uri = req[i+1:]
	} else {
		return "", ""
	}
	i = strings.Index(uri, " ")
	if i >= 0 {
		uri = uri[:i]
	}
	i = strings.Index(uri, "?")
	if i >= 0 {
		uri = uri[:i]
	}
	for _, p := range patterns {
		if p.MatchString(uri) {
			uri = p.String()
			return method, uri
		}
	}
	return method, uri
}

func isIgnored(req string, ignorePatterns []regexp.Regexp) bool {
	uri := req
	i := strings.Index(req, " ")
	if i >= 0 {
		uri = req[i+1:]
	}
	i = strings.Index(req, "?")
	if i >= 0 {
		uri = uri[:i]
	}

	for _, p := range ignorePatterns {
		if p.MatchString(uri) {
			return true
		}
	}
	return false
}
