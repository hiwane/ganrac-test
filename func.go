package ganrac

import (
	"fmt"
	"sort"
)

// 関数テーブル
var builtin_func_table = []struct {
	name     string
	min, max int
	f        func(args []interface{}) (interface{}, error)
	help     string
}{
	// sorted by name
	{"all", 2, 2, funcForAll, ""},
	{"and", 2, 2, funcAnd, ""},
	{"coef", 3, 3, funcCoef, ""}, // coef(F, x, 2)
	{"deg", 2, 2, funcDeg, ""},   // deg(F, x)
	{"ex", 2, 2, funcExists, ""},
	{"help", 0, 1, funcHelp, ""},
	{"init", 0, 0, funcHelp, ""},
	{"not", 1, 1, funcNot, ""},
	{"or", 2, 2, funcOr, ""},
	{"subst", 1, 101, funcSubst, ""},
}

func (p *pNode) callFunction(args []interface{}) (interface{}, error) {
	// とりあえず素朴に
	for _, f := range builtin_func_table {
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
		if !ok || !q.isVar() {
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

func funcSubst(args []interface{}) (interface{}, error) {
	if len(args)%2 != 1 {
		return nil, fmt.Errorf("subst() invalid args")
	} else if len(args) == 1 {
		return args[0], nil
	}

	rlv := make([]struct {
		r  RObj
		lv Level
	}, (len(args)-1)/2)

	j := 0
	for i := 1; i < len(args); i += 2 {
		p, ok := args[i].(*Poly)
		if !ok || !p.isVar() {
			return nil, fmt.Errorf("subst() invalid %d'th arg: %v", i+1, args[i])
		}
		// 重複を除去
		used := false
		for k := 0; k < j; k++ {
			if rlv[k].lv == p.lv {
				used = true
				break
			}
		}
		if used {
			continue
		}

		rlv[j].lv = p.lv

		v, ok := args[i+1].(RObj)
		if !ok {
			return nil, fmt.Errorf("subst() invalid %d'th arg", i+2)
		}
		rlv[j].r = v
		j += 1
	}
	rlv = rlv[:j]

	sort.SliceStable(rlv, func(i, j int) bool {
		return rlv[i].lv < rlv[j].lv
	})

	rr := make([]RObj, len(rlv))
	lv := make([]Level, len(rlv))
	for i := 0; i < j; i++ {
		rr[i] = rlv[i].r
		lv[i] = rlv[i].lv
	}

	switch f := args[0].(type) {
	case Fof:
		return f.Subst(rr, lv), nil
	case RObj:
		return f.Subst(rr, lv, 0), nil
	}

	return nil, fmt.Errorf("subst() invalid 1st arg")

}

func funcDeg(args []interface{}) (interface{}, error) {
	// FoF にも適用可能にする.
	name := "deg"
	_, ok := args[0].(RObj)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected poly: %v", name, args[0])
	}

	d, ok := args[1].(*Poly)
	if !ok || !d.isVar() {
		return nil, fmt.Errorf("%s(2st arg): expected var: %v", name, args[1])
	}

	p, ok := args[0].(*Poly)
	if !ok {
		return zero, nil
	}

	return NewInt(int64(p.Deg(d.lv))), nil
}

func funcCoef(args []interface{}) (interface{}, error) {
	name := "coef"
	_, ok := args[0].(RObj)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected RObj: %v", name, args[0])
	}

	c, ok := args[1].(*Poly)
	if !ok || !c.isVar() {
		return nil, fmt.Errorf("%s(2st arg): expected var: %v", name, args[1])
	}

	d, ok := args[2].(*Int)
	if !ok {
		return nil, fmt.Errorf("%s(3rd arg): expected int: %v", name, args[2])
	}

	if d.Sign() < 0 {
		return zero, nil
	}
	rr, ok := args[0].(*Poly)
	if !ok {
		if d.Sign() == 0 {
			return args[0], nil
		} else {
			return zero, nil
		}
	}
	if !d.n.IsUint64() {
		return zero, nil
	}

	return rr.Coef(c.lv, uint(d.n.Uint64())), nil
}

func funcHelp(args []interface{}) (interface{}, error) {
	return zero, nil
}