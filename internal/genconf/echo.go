package genconf

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/haijima/analysisutil/ssautil"
)

type EchoExtractor struct{}

func (e *EchoExtractor) Extract(callInfo ssautil.CallInfo) (string, bool) {
	if c, ok := callInfo.(*ssautil.StaticMethodCall); ok {
		if slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "Any", "Static", "File"}, c.Method().Name()) {
			return extract(c, 0)
		} else if slices.Contains([]string{"Add", "Match"}, c.Method().Name()) {
			return extract(c, 1)
		}
	}
	return "", false
}

func extract(c *ssautil.StaticMethodCall, idx int) (string, bool) {
	if arg, ok := valueToStringConst(c.Arg(idx)); ok {
		switch c.Recv().Type().String() {
		case "*github.com/labstack/echo/v4.Echo":
			return arg, true
		case "*github.com/labstack/echo/v4.Group":
			paths, ok := groupPrefixes(c)
			return strings.Join(append(paths, arg), ""), ok
		}
	} else {
		slog.Warn("failed to parse path", "arg", c.Arg(idx))
	}
	return "", false
}

func groupPrefixes(staticMethodCall *ssautil.StaticMethodCall) ([]string, bool) {
	if call, ok := ssautil.ValueToCallCommon(staticMethodCall.Recv()); ok {
		if c, ok := ssautil.GetCallInfo(call).(*ssautil.StaticMethodCall); ok {
			if s, ok := valueToStringConst(c.Arg(0)); ok {
				switch c.Name() {
				case "(*github.com/labstack/echo/v4.Echo).Group":
					return []string{s}, true
				case "(*github.com/labstack/echo/v4.Group).Group":
					prefixes, ok := groupPrefixes(c)
					return append(prefixes, s), ok
				}
			} else {
				slog.Warn("failed to parse path", "arg", c.Arg(0))
			}
		}
	}
	return nil, false
}

var echoPathParamPattern = regexp.MustCompile(":([^/]+)")

func (e *EchoExtractor) RegexpPattern(path string, _ bool) string {
	return fmt.Sprintf("^%s$", echoPathParamPattern.ReplaceAllString(path, "([^/]+)"))
}
