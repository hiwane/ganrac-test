package ganrac

import (
	"fmt"
)

// 関数テーブル
var func_table = []struct {
	name     string
	min, max int
	f        func(args []interface{}) (interface{}, error)
}{
	// sorted by name
	{"not", 1, 1, funcNot},
	{"and", 2, 2, funcAnd},
	{"or", 2, 2, funcOr},
	{"ex", 2, 2, funcExists},
	{"all", 2, 2, funcForAll},
}

func (p *pNode) callFunction(args []interface{}) (interface{}, error) {
	// とりあえず素朴に
	for _, f := range func_table {
		if f.name == p.str {
			if len(args) < f.min {
				return nil, fmt.Errorf("too few argument: function %s()", p.str)
			}
			if len(args) > f.max {
				return nil, fmt.Errorf("too many argument: function %s()", p.str)
			}
			return f.f(args)
		}
	}

	return nil, fmt.Errorf("unknown function: %s", p.str)
}

func funcNot(args []interface{}) (interface{}, error) {
	f, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("not(): unsupported for %v", args[0])
	}
	return f.Not(), nil
}

func funcAnd(args []interface{}) (interface{}, error) {
	f0, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("and(): unsupported for %v", args[0])
	}
	f1, ok := args[1].(Fof)
	if !ok {
		return nil, fmt.Errorf("and(): unsupported for %v", args[1])
	}
	return NewFmlAnd(f0, f1), nil
}

func funcOr(args []interface{}) (interface{}, error) {
	f0, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("or(): unsupported for %v", args[0])
	}
	f1, ok := args[1].(Fof)
	if !ok {
		return nil, fmt.Errorf("or(): unsupported for %v", args[1])
	}
	return NewFmlOr(f0, f1), nil
}

func funcExists(args []interface{}) (interface{}, error) {
	return funcForEx(false, "ex", args)
}

func funcForAll(args []interface{}) (interface{}, error) {
	return funcForEx(true, "all", args)
}

func funcForEx(forex bool, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(*List)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected list: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}

	lv := make([]Level, len(f0.v))
	for i, qq := range f0.v {
		q, ok := qq.(*Poly)
		if !ok || len(q.c) != 2 || !q.c[0].IsZero() || !q.c[1].IsOne() {
			return nil, fmt.Errorf("%s(1st arg:%d): expected var-list", name, i)
		}
		lv[i] = q.lv
	}

	f1, ok := args[1].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(2nd arg): expected formula", name)
	}
	return NewQuantifier(forex, lv, f1), nil
}
