package genconf

import (
	"go/ast"
	"go/token"
)

type ArgNotBasicLitError struct {
	Info []*ArgAstInfo
}

type ArgAstInfo struct {
	Call     *ast.CallExpr
	CallPos  token.Position
	ArgPos   token.Position
	ArgIndex int
}

func (e *ArgNotBasicLitError) Error() string {
	return "not basic literal"
}
