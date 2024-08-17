package endpoint

import (
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"

	"github.com/haijima/analysisutil/ssautil"
	"golang.org/x/tools/go/ssa"
)

type WebFramework int

type Endpoint struct {
	Method     string
	Path       string
	FuncName   string
	Comment    string
	DeclarePos *Pos
	FuncPos    *Pos
}

var pathDirRegex = regexp.MustCompile(`([^/]+)/`)

type Pos struct {
	Func *ssa.Function
	Pos  []token.Pos
}

func NewPos(fn *ssa.Function, pos ...token.Pos) *Pos {
	return &Pos{Func: fn, Pos: pos}
}

func (m *Pos) Package() *types.Package {
	if m.Func == nil || m.Func.Pkg == nil {
		return &types.Package{}
	}
	return m.Func.Pkg.Pkg
}

func (p *Pos) PackagePath() string {
	return pathDirRegex.ReplaceAllStringFunc(p.Package().Path(), func(m string) string { return m[:1] + "/" })
}

func (m *Pos) Position() token.Position {
	if m.Func == nil {
		return token.Position{}
	}
	return ssautil.GetPosition(m.Func.Pkg, m.Pos...)
}

func (p *Pos) FLC() string {
	return filepath.Base(p.Position().String())
}

type Extractor interface {
	Extract(callInfo ssautil.CallInfo, parent *ssa.Function, pos token.Pos) (*Endpoint, bool)
}

func FindEndpoints(dir, pattern string, ext Extractor) ([]*Endpoint, error) {
	ssaProgs, err := ssautil.LoadBuildSSAs(dir, pattern)
	if err != nil {
		return nil, err
	}

	result := make([]*Endpoint, 0)
	for _, ssaProg := range ssaProgs {
		for _, fn := range ssaProg.SrcFuncs {
			for _, b := range fn.Blocks {
				for _, instr := range b.Instrs {
					if call, ok := instr.(*ssa.Call); ok {
						if p, ok := ext.Extract(ssautil.GetCallInfo(&call.Call), call.Parent(), call.Pos()); ok {
							result = append(result, p)
						}
					}
				}
			}
		}
	}

	return result, nil
}
