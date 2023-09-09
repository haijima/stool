package genconf

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/ast/astutil"
)

var pathParamPattern = regexp.MustCompile(":([^/]+)")

func GenMatchingGroupFromEchoV4(fileName string, echoPkgName string, captureGroupName bool) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.AllErrors)
	if err != nil {
		return []string{}, err
	}

	// Check e := echo.New()
	echoRcvrName := "e"
	astutil.Apply(node, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.AssignStmt:
			if len(x.Lhs) > 0 && len(x.Rhs) > 0 {
				rh, ok := x.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}
				f, ok := rh.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := f.X.(*ast.Ident)
				if !ok {
					return true
				}
				method := f.Sel.Name
				if ident.Name == echoPkgName && method == "New" {
					lh, ok := x.Lhs[0].(*ast.Ident)
					if !ok {
						return true
					}
					echoRcvrName = lh.Name
				}
			}
		}
		return true
	})

	errInfo := make([]*ArgAstInfo, 0)

	// Check e.Group()
	type Group struct {
		Ident     string
		PathValue string
	}
	groups := make([]Group, 0)
	astutil.Apply(node, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.AssignStmt:
			if len(x.Lhs) > 0 && len(x.Rhs) > 0 {
				rh, ok := x.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}
				f, ok := rh.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := f.X.(*ast.Ident)
				if !ok {
					return true
				}
				method := f.Sel.Name
				if ident.Name == echoRcvrName && method == "Group" {
					lh, ok := x.Lhs[0].(*ast.Ident)
					if !ok {
						return true
					}
					arg, ok := rh.Args[0].(*ast.BasicLit)
					if !ok {
						// TODO: Error
						errInfo = append(errInfo, &ArgAstInfo{
							Call:     rh,
							CallPos:  fset.Position(rh.Pos()),
							ArgPos:   fset.Position(arg.Pos()),
							ArgIndex: 0,
						})
					}
					groupPath, err := strconv.Unquote(arg.Value)
					if err != nil {
						return true
					}
					groups = append(groups, Group{Ident: lh.Name, PathValue: groupPath})
				}
			}
		}
		return true
	})

	// Check e.GET(), e.POST(), e.PUT(), ...
	targetFuncNames := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "Static", "CONNECT", "HEAD", "OPTIONS", "TRACE", "RouteNotFound", "Any", "Match"}
	endpoints := make([]string, 0)
	astutil.Apply(node, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.CallExpr:
			selector, ok := x.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			ok = false
			if ident.Name == echoRcvrName {
				ok = true
			} else {
				for _, group := range groups {
					if group.Ident == ident.Name {
						ok = true
						break
					}
				}
			}
			if !ok {
				return true
			}
			methodName := selector.Sel.Name
			if !slices.Contains(targetFuncNames, methodName) {
				return true
			}
			if len(x.Args) == 0 {
				return true
			}
			switch arg := x.Args[0].(type) {
			case *ast.BasicLit:
				unquote, err := strconv.Unquote(arg.Value)
				if err != nil {
					return true
				}
				for _, group := range groups {
					if group.Ident == ident.Name {
						unquote = group.PathValue + unquote
						break
					}
				}
				if captureGroupName {
					replaced := pathParamPattern.ReplaceAllString(unquote, "(?P<$1>[^/]+)")
					if methodName == "Static" {
						endpoints = append(endpoints, fmt.Sprintf("^%s/(?P<filepath>.+)", replaced))
					} else {
						endpoints = append(endpoints, fmt.Sprintf("^%s$", replaced))
					}
				} else {
					replaced := pathParamPattern.ReplaceAllString(unquote, "([^/]+)")
					if methodName == "Static" {
						endpoints = append(endpoints, fmt.Sprintf("^%s/(.+)", replaced))
					} else {
						endpoints = append(endpoints, fmt.Sprintf("^%s$", replaced))
					}
				}
			default:
				// add error info
				errInfo = append(errInfo, &ArgAstInfo{
					Call:     x,
					CallPos:  fset.Position(x.Pos()),
					ArgPos:   fset.Position(arg.Pos()),
					ArgIndex: 0,
				})
			}
		}
		return true
	})

	set := make(map[string]interface{})
	for _, e := range endpoints {
		set[e] = nil
	}
	patterns := maps.Keys(set)
	slices.Sort(patterns)
	if len(errInfo) > 0 {
		return patterns, &ArgNotBasicLitError{Info: errInfo}
	}
	return patterns, nil
}
