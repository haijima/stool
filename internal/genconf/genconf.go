package genconf

import (
	"slices"
	"strings"

	"github.com/haijima/analysisutil"
	"github.com/haijima/analysisutil/ssautil"
	"golang.org/x/tools/go/ssa"
)

type WebFramework int

const (
	EchoV4 WebFramework = iota
	None
)

func CheckImportedFramework(dir, pattern string) (WebFramework, error) {
	pkgs, err := analysisutil.LoadPackages(dir, pattern)
	if err != nil {
		return None, err
	}

	for _, pkg := range pkgs {
		for _, p := range pkg.Imports {
			if p.PkgPath == "github.com/labstack/echo/v4" {
				return EchoV4, nil
			}
		}
	}
	return None, nil
}

type APIPathPatternExtractor interface {
	Extract(callInfo ssautil.CallInfo) (string, bool)
	RegexpPattern(path string, captureGroupName bool) string
}

func GenMatchingGroup(dir, pattern string, ext APIPathPatternExtractor, captureGroupName bool) ([]string, error) {
	ssaProgs, err := ssautil.LoadBuildSSAs(dir, pattern)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, ssaProg := range ssaProgs {
		for _, fn := range ssaProg.SrcFuncs {
			for _, b := range fn.Blocks {
				for _, instr := range b.Instrs {
					if call, ok := instr.(*ssa.Call); ok {
						if p, ok := ext.Extract(ssautil.GetCallInfo(&call.Call)); ok {
							result = append(result, p)
						}
					}
				}
			}
		}
	}

	slices.Sort(result) // Sort before compact
	result = slices.Compact(result)
	for i := range result {
		result[i] = ext.RegexpPattern(result[i], captureGroupName)
	}
	slices.SortFunc(result, func(i, j string) int { return strings.Compare(j, i) }) // Descending order
	return result, nil
}

func valueToStringConst(v ssa.Value) (string, bool) {
	if ss, ok := ssautil.ValueToStrings(v); ok && len(ss) == 1 {
		return ss[0], true
	}
	return "", false
}
