package genconf

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/haijima/analysisutil/ssautil"
)

type NetHttpExtractor struct{}

func (e *NetHttpExtractor) Extract(callInfo ssautil.CallInfo) (string, bool) {
	if callInfo.Match("(*net/http.ServeMux).Handle") ||
		callInfo.Match("(*net/http.ServeMux).HandleFunc") ||
		callInfo.Match("net/http.Handle") ||
		callInfo.Match("net/http.HandleFunc") {
		if arg, ok := valueToStringConst(callInfo.Arg(0)); ok {
			s := strings.Split(arg, " ")
			return s[len(s)-1], true
		}
		slog.Warn("failed to parse path", "arg", callInfo.Arg(0))
	}
	return "", false
}

var netHttpPathPatternEndsWithSlash = regexp.MustCompile("/$")
var netHttpPathPatternEndsWithDollar = regexp.MustCompile("\\{\\$\\}$")
var netHttpPathPatternEndsWithDotsParam = regexp.MustCompile("\\{([a-zA-Z0-9_-]+)\\.\\.\\.\\}$")
var netHttpPathParamPattern = regexp.MustCompile("\\{([a-zA-Z0-9_-]+)\\}")

func (e *NetHttpExtractor) RegexpPattern(path string, _ bool) string {
	path = netHttpPathPatternEndsWithSlash.ReplaceAllString(path, "/(.*)")
	path = netHttpPathPatternEndsWithDollar.ReplaceAllString(path, "")
	path = netHttpPathPatternEndsWithDotsParam.ReplaceAllString(path, "(.+)")
	path = netHttpPathParamPattern.ReplaceAllString(path, "([^/]+)")
	return fmt.Sprintf("^%s$", path)
}
