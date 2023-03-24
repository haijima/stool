package log

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Wing924/ltsv"
	"golang.org/x/exp/maps"
)

type LTSVReader struct {
	r                *bufio.Scanner
	timeFormat       string
	matchingPatterns []regexp.Regexp
	ignorePatterns   []regexp.Regexp
	labels           map[string]string
	line             int
}

type LTSVReadOpt struct {
	MatchingGroups []string
	IgnorePatterns []string
	TimeFormat     string
	Labels         map[string]string
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

	labels := maps.Clone(defaultLabels)
	for k, v := range opt.Labels {
		if _, ok := labels[k]; ok {
			labels[k] = v
		}
	}

	return &LTSVReader{
		r:                scanner,
		matchingPatterns: matchingregexps,
		ignorePatterns:   ignoreRegexps,
		timeFormat:       timeFormat,
		labels:           labels,
	}, nil
}

var defaultLabels = map[string]string{
	"req":    "req",
	"status": "status",
	"time":   "time",
	"uidset": "uidset",
	"uidgot": "uidgot",
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
	scanned := r.r.Scan()
	if scanned {
		r.line++
	}
	return scanned
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
		case r.labels["req"]:
			entry.Req = string(value)
			method, uri := parseReq(string(value), r.matchingPatterns)
			entry.Method = method
			entry.Uri = uri
			entry.IsIgnored = isIgnored(uri, r.ignorePatterns)

		case r.labels["status"]:
			status, err := strconv.Atoi(string(value))
			if err != nil {
				return err
			}
			entry.Status = status

		case r.labels["time"]:
			reqTime, err := time.Parse(r.timeFormat, string(value))
			if err != nil {
				return err
			}
			entry.Time = reqTime

		case r.labels["uidset"]:
			if string(value) != "" && string(value) != "-" {
				if i := strings.Index(string(value), "="); i >= 0 {
					entry.Uid = string(value)[i+1:]
				} else {
					entry.Uid = string(value)
				}
				entry.SetNewUid = true
			}

		case r.labels["uidgot"]:
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

	if entry.Req == "" {
		return nil, fmt.Errorf("\"%s\" field is not found on line %d", r.labels["req"], r.line)
	} else if entry.Status == 0 {
		return nil, fmt.Errorf("\"%s\" field is not found on line %d", r.labels["status"], r.line)
	} else if entry.Time.IsZero() {
		return nil, fmt.Errorf("\"%s\" field　is not found on line %d", r.labels["time"], r.line)
	} else if entry.Uid == "" {
		return nil, fmt.Errorf("\"%s\" or \"%s\" field is not found on line %d", r.labels["uidset"], r.labels["uidgot"], r.line)
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
			return method, p.String()
		}
	}
	return method, uri
}

func isIgnored(uri string, ignorePatterns []regexp.Regexp) bool {
	for _, p := range ignorePatterns {
		if p.MatchString(uri) {
			return true
		}
	}
	return false
}
