package log

import (
	"errors"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type FilterExpr struct {
	program *vm.Program
}

func NewFilterExpr(code string) (*FilterExpr, error) {
	if code == "" {
		code = "true"
	}
	program, err := expr.Compile(code, expr.Env(&ExprEnv{}), expr.AsBool(),
		expr.Operator("=", "TimeStringEqual"),
		expr.Operator("!=", "TimeStringNotEqual"),
		expr.Operator(">", "TimeStringGreaterThan"),
		expr.Operator(">=", "TimeStringGreaterThanEqual"),
		expr.Operator("<", "TimeStringLessThan"),
		expr.Operator("<=", "TimeStringLessThanEqual"),
		expr.Operator("=", "StringTimeEqual"),
		expr.Operator("!=", "StringTimeNotEqual"),
		expr.Operator(">", "StringTimeGreaterThan"),
		expr.Operator(">=", "StringTimeGreaterThanEqual"),
		expr.Operator("<", "StringTimeLessThan"),
		expr.Operator("<=", "StringTimeLessThanEqual"),
	)
	if err != nil {
		return nil, err
	}

	return &FilterExpr{
		program: program,
	}, nil
}

func (f *FilterExpr) Run(entry LogEntry) (bool, error) {
	run, err := expr.Run(f.program, ExprEnv{
		Req:                        entry.Req,
		Method:                     entry.Method,
		Uri:                        entry.Uri,
		Status:                     entry.Status,
		Time:                       entry.Time,
		Uid:                        entry.Uid,
		SetNewUid:                  entry.SetNewUid,
		TimeStringEqual:            TimeStringEqual,
		TimeStringNotEqual:         TimeStringNotEqual,
		TimeStringGreaterThan:      TimeStringGreaterThan,
		TimeStringGreaterThanEqual: TimeStringGreaterThanEqual,
		TimeStringLessThan:         TimeStringLessThan,
		TimeStringLessThanEqual:    TimeStringLessThanEqual,
		StringTimeEqual:            StringTimeEqual,
		StringTimeNotEqual:         StringTimeNotEqual,
		StringTimeGreaterThan:      StringTimeGreaterThan,
		StringTimeGreaterThanEqual: StringTimeGreaterThanEqual,
		StringTimeLessThan:         StringTimeLessThan,
		StringTimeLessThanEqual:    StringTimeLessThanEqual,
	})
	if err != nil {
		return false, err
	}
	b, ok := run.(bool)
	if !ok {
		return false, errors.New("not a bool")
	}
	return b, nil
}

type ExprEnv struct {
	Req                        string    `expr:"req"`
	Method                     string    `expr:"method"`
	Uri                        string    `expr:"uri"`
	Status                     int       `expr:"status"`
	Time                       time.Time `expr:"time"`
	Uid                        string    `expr:"uid"`
	SetNewUid                  bool      `expr:"setNewUid"`
	TimeStringEqual            func(l time.Time, r string) bool
	TimeStringNotEqual         func(l time.Time, r string) bool
	TimeStringGreaterThan      func(l time.Time, r string) bool
	TimeStringGreaterThanEqual func(l time.Time, r string) bool
	TimeStringLessThan         func(l time.Time, r string) bool
	TimeStringLessThanEqual    func(l time.Time, r string) bool
	StringTimeEqual            func(l string, r time.Time) bool
	StringTimeNotEqual         func(l string, r time.Time) bool
	StringTimeGreaterThan      func(l string, r time.Time) bool
	StringTimeGreaterThanEqual func(l string, r time.Time) bool
	StringTimeLessThan         func(l string, r time.Time) bool
	StringTimeLessThanEqual    func(l string, r time.Time) bool
}

var timeLayout = "2006-01-02 15:04:05"

func toTime(s string) time.Time {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TimeStringEqual(l time.Time, r string) bool            { return l.Equal(toTime(r)) }
func TimeStringNotEqual(l time.Time, r string) bool         { return !l.Equal(toTime(r)) }
func TimeStringGreaterThan(l time.Time, r string) bool      { return l.After(toTime(r)) }
func TimeStringGreaterThanEqual(l time.Time, r string) bool { return !l.Before(toTime(r)) }
func TimeStringLessThan(l time.Time, r string) bool         { return l.Before(toTime(r)) }
func TimeStringLessThanEqual(l time.Time, r string) bool    { return !l.After(toTime(r)) }
func StringTimeEqual(l string, r time.Time) bool            { return toTime(l).Equal(r) }
func StringTimeNotEqual(l string, r time.Time) bool         { return !toTime(l).Equal(r) }
func StringTimeGreaterThan(l string, r time.Time) bool      { return toTime(l).After(r) }
func StringTimeGreaterThanEqual(l string, r time.Time) bool { return !toTime(l).Before(r) }
func StringTimeLessThan(l string, r time.Time) bool         { return toTime(l).Before(r) }
func StringTimeLessThanEqual(l string, r time.Time) bool    { return !toTime(l).After(r) }
