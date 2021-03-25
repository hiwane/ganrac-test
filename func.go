package ganrac

import (
	"fmt"
	"sort"
)

// 関数テーブル
var builtin_func_table = []struct {
	name     string
	min, max int
	f        func(name string, args []interface{}) (interface{}, error)
	descript string
	help     string
}{
	// sorted by name
	{"all", 2, 2, funcForAll, "([x], FOF): universal quantifier.", ""},
	{"and", 2, 2, funcAnd, "(FOF, ...): conjunction (&&)", ""},
	{"coef", 3, 3, funcCoef, "(poly, var, deg): ", ""}, // coef(F, x, 2)
	{"deg", 2, 2, funcDeg, "(poly|FOF, var): degree of a polynomial with respect to var", `
Args
========
  poly: a polynomial
  FOF : a first-order formula
  var : a variable

Returns
========
  degree of a polynomial w.r.t a variable

Examples
========
  > deg(x^2+3*y+1, x);
  2
  > deg(x^2+3*y+1, y);
  1
  > deg(x^2+3*y+1>0 && x^3+y^3==0, y);
  3
  > deg(0, y);
  -1
`}, // deg(F, x)
	{"ex", 2, 2, funcExists, "(vars, FOF): existential quantifier.", `
Args
========
  vars: list of variables
  FOF : a first-order formula

Examples
========
  > ex([x], a*x^2+b*x+c == 0);
`},
	// {"fctr", 1, 1, funcFctr, "(poly)*: factorize polynomial over the rationals.", ""},
	{"help", 0, 1, nil, "(): show help", ""},
	{"indets", 1, 1, funcIndets, "(mobj): find indeterminates of an expression", ""},
	{"init", 0, 0, nil, "(vars, ...): init variable order", ""},
	{"len", 1, 1, funcLen, "(mobj): length of an object", ""},
	{"load", 2, 2, funcLoad, "(fname): load file", ""},
	{"not", 1, 1, funcNot, "(FOF)", `
Args
========
  FOF : a first-order formula

Examples
========
  > not(x > 0);
  x <= 0
  > not(ex([x], a*x^2+b*x+c==0));
  all([x], a*x^2+b*x+c != 0)
`},
	{"or", 2, 2, funcOr, "(FOF, ...): disjunction (||)", ""},
	{"realroot", 2, 2, funcRealRoot, "(uni-poly): real root isolation", ""},
	{"rootbound", 1, 1, funcRootBound, "(uni-poly in Z[x]): root bound", `
Args
========
  poly: univariate polynomial

Examples
========
  > realroot(x^2-2);

`},
	{"save", 2, 3, funcSave, "(obj, fname): save object...", ""},
	// {"sqrt", 1, 1, funcSqrt, "(poly)*: square-free factorization", ""},
	{"subst", 1, 101, funcSubst, "(poly,x,vx,y,vy,...)", ""},
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
			if f.name == "help" {
				return funcHelp(f.name, args)
			} else {
				return f.f(f.name, args)
			}
		}
	}

	return nil, fmt.Errorf("unknown function: %s", p.str)
}

func funcNot(name string, args []interface{}) (interface{}, error) {
	f, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("not(): unsupported for %v", args[0])
	}
	return f.Not(), nil
}

func funcAnd(name string, args []interface{}) (interface{}, error) {
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

func funcOr(name string, args []interface{}) (interface{}, error) {
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

func funcExists(name string, args []interface{}) (interface{}, error) {
	return funcForEx(false, name, args)
}

func funcForAll(name string, args []interface{}) (interface{}, error) {
	return funcForEx(true, name, args)
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

func funcSubst(name string, args []interface{}) (interface{}, error) {
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

func funcDeg(name string, args []interface{}) (interface{}, error) {
	// FoF にも適用可能にする.
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
		if p.IsZero() {
			return mone, nil
		} else {
			return zero, nil
		}
	}

	return NewInt(int64(p.Deg(d.lv))), nil
}

func funcCoef(name string, args []interface{}) (interface{}, error) {
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

func funcRealRoot(name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(): expected poly: %v", name, args[0])
	}
	return p.RealRootIsolation(1)
}

func funcRootBound(name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(): expected poly: %v", name, args[0])
	}
	return p.RootBound()
}

func funcIndets(name string, args []interface{}) (interface{}, error) {
	b := make([]bool, len(varlist))
	p, ok := args[0].(Indeter)
	ret := make([]interface{}, 0, len(b))
	if !ok {
		return NewList(ret), nil
	}
	p.Indets(b)

	for i := 0; i < len(b); i++ {
		if b[i] {
			ret = append(ret, NewPolyVar(Level(i)))
		}
	}
	return NewList(ret), nil
}

func funcLoad(name string, args []interface{}) (interface{}, error) {
	return nil, fmt.Errorf("%s not implemented", name) // @TODO
}

func funcSave(name string, args []interface{}) (interface{}, error) {
	return nil, fmt.Errorf("%s not implemented", name) // @TODO
}

func funcLen(name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(Lener)
	if !ok {
		return nil, fmt.Errorf("%s(): not supported: %v", name, args[0])
	}
	return NewInt(int64(p.Len())), nil
}

func funcHelp(name string, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return funcHelps("@")
	}

	p, ok := args[0].(*String)
	if !ok {
		return nil, fmt.Errorf("%s(): required help(\"string\"):", name)
	}

	return funcHelps(p.s)
}

func funcHelps(name string) (interface{}, error) {
	if name == "@" {
		fmt.Printf("GANRAC\n")
		fmt.Printf("==========\n")
		fmt.Printf("TOKENS:\n")
		fmt.Printf("  integer      : `[0-9]+`\n")
		fmt.Printf("  string       : `\"[^\"]*\"`\n")
		fmt.Printf("  indeterminate: `[a-z][a-zA-Z0-9_]*`\n")
		fmt.Printf("  variable     : `[A-Z][a-zA-Z0-9_]*`\n")
		fmt.Printf("  true/false   : `true`/`false`\n")
		fmt.Printf("\n")
		fmt.Printf("OPERATORS:\n")
		fmt.Printf("  + - * / ^\n")
		fmt.Printf("  < <= > >= == !=\n")
		fmt.Printf("  && ||\n")
		fmt.Printf("\n")
		fmt.Printf("FUNCTIONS:\n")
		for _, fv := range builtin_func_table {
			fmt.Printf("  %s%s\n", fv.name, fv.descript)
		}
		fmt.Printf("\n")
		fmt.Printf("EXAMPLES:\n")
		fmt.Printf("  > init(x, y, z);  # init variable order.\n")
		fmt.Printf("  > F = x^2 + 2;\n")
		fmt.Printf("  > deg(F, x);\n")
		fmt.Printf("  2\n")
		fmt.Printf("  > t;\n")
		fmt.Printf("  error: undefined variable `t`\n")
		fmt.Printf("  > qe(ex([x], x+1 = 0 && x < 0));\n")
		fmt.Printf("  true\n")
		fmt.Printf("  > help(\"deg\");\n")
		return nil, nil
	}
	for _, fv := range builtin_func_table {
		if fv.name == name {
			fmt.Printf("%s%s\n", fv.name, fv.descript)
			if fv.help != "" {
				fmt.Printf("%s\n", fv.help)
			}
			return nil, nil
		}
	}

	return nil, fmt.Errorf("unknown function `%s()`\n", name)
}
