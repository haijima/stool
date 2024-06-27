package log

import (
	"errors"

	"github.com/google/cel-go/cel"
)

type FilterExpr struct {
	program cel.Program
}

func NewFilterExpr(code string) (*FilterExpr, error) {
	if code == "" {
		code = "true"
	}
	env, err := cel.NewEnv(
		cel.Variable("req", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("uri", cel.StringType),
		cel.Variable("status", cel.IntType),
		cel.Variable("time", cel.TimestampType),
		cel.Variable("uid", cel.StringType),
		cel.Variable("set_new_uid", cel.BoolType),
	)
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(code)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	return &FilterExpr{program: prg}, nil
}

func (f *FilterExpr) Run(entry LogEntry) (bool, error) {
	out, _, err := f.program.Eval(map[string]any{
		"req":         entry.Req,
		"method":      entry.Method,
		"uri":         entry.Uri,
		"status":      entry.Status,
		"time":        entry.Time,
		"uid":         entry.Uid,
		"set_new_uid": entry.SetNewUid,
	})
	if err != nil {
		return false, err
	}

	b, ok := out.Value().(bool)
	if !ok {
		return false, errors.New("not a bool")
	}
	return b, nil
}
