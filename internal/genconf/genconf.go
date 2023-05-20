package genconf

import (
	"go/parser"
	"go/token"
	"strconv"
)

type WebFramework int

const (
	EchoV4 WebFramework = iota
	None
)

type WebFrameworkInfo struct {
	PkgName string
	Kind    WebFramework
}

func CheckImportedFramework(fileName string) (*WebFrameworkInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	// Check echo import
	echoPkgName := "echo"
	var useEcho bool
	for _, x := range node.Imports {
		if path, err := strconv.Unquote(x.Path.Value); err == nil && path == "github.com/labstack/echo/v4" {
			if x.Name != nil {
				echoPkgName = x.Name.Name
			}
			useEcho = true
			break
		}
	}
	if useEcho {
		return &WebFrameworkInfo{PkgName: echoPkgName, Kind: EchoV4}, nil
	} else {
		return &WebFrameworkInfo{Kind: None}, nil
	}
}
