package ganrac

import (
	"fmt"
	"sort"
	"time"
)

func (g *Ganrac) setBuiltinFuncTable() {
	// 関数テーブル
	g.builtin_func_table = []func_table{
		// sorted by name
		{"all", 2, 2, funcForAll, false, "([x], FOF): universal quantifier.", ""},
		{"and", 2, 2, funcAnd, false, "(FOF, ...): conjunction (&&)", ""},
		{"coef", 3, 3, funcCoef, false, "(poly, var, deg): ", ""}, // coef(F, x, 2)
		{"deg", 2, 2, funcDeg, false, "(poly|FOF, var): degree of a polynomial with respect to var", `
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
		{"equiv", 2, 2, funcEquiv, false, "(fof1, fof2): fof1 is equivalent to fof2", ""},
		{"ex", 2, 2, funcExists, false, "(vars, FOF): existential quantifier.", `
Args
========
  vars: list of variables
  FOF : a first-order formula

Examples
========
  > ex([x], a*x^2+b*x+c == 0);
`},
		// {"fctr", 1, 1, funcFctr, false, "(poly)* factorize polynomial over the rationals.", ""},
		{"help", 0, 1, nil, false, "(): show help", ""},
		{"igcd", 2, 2, funcIGCD, false, "(int1, int2): The integer greatest common divisor", ""},
		{"impl", 2, 2, funcImpl, false, "(fof1, fof2): fof1 impies fof2", ""},
		{"indets", 1, 1, funcIndets, false, "(mobj): find indeterminates of an expression", ""},
		{"init", 0, 0, nil, false, "(vars, ...): init variable order", ""},
		{"len", 1, 1, funcLen, false, "(mobj): length of an object", ""},
		{"load", 2, 2, funcLoad, false, "(fname): load file", ""},
		{"not", 1, 1, funcNot, false, "(FOF)", `
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
		{"or", 2, 2, funcOr, false, "(FOF, ...): disjunction (||)", ""},
		{"oxfunc", 2, 100, funcOXFunc, true, "(fname, args...)* call ox-function by ox-asir", `
Args
========
fname : string, function name of ox-server
args  : arguments of the function

Examples
========
  > oxfunc("deg", x^2-1, x);
  2
  > oxfunc("igcd", 8, 12);
  4
`},
		{"oxstr", 1, 1, funcOXStr, true, "(str)* evaluate str by ox-asir", `
Args
========
str : string

Examples
========
  > oxstr("fctr(x^2-4);");
  [[1,1],[x-2,1],[x+2,1]]
`},
		{"realroot", 2, 2, funcRealRoot, false, "(uni-poly): real root isolation", ""},
		{"rootbound", 1, 1, funcRootBound, false, "(uni-poly in Z[x]): root bound", `
Args
========
  poly: univariate polynomial

Examples
========
  > rootbound(x^2-2);
  3
`},
		{"save", 2, 3, funcSave, false, "(obj, fname): save object...", ""},
		{"sleep", 1, 1, funcSleep, false, "(milisecond): zzz", ""},
		// {"sqrt", 1, 1, funcSqrt, false, "(poly)* square-free factorization", ""},
		{"subst", 1, 101, funcSubst, false, "(poly,x,vx,y,vy,...):", ""},
		{"time", 1, 1, funcTime, false, "(expr)@ run command and system resource usage", ""},
	}
}

// func (p *pNode) callFunction(args []interface{}) (interface{}, error) {
func (g *Ganrac) callFunction(funcname string, args []interface{}) (interface{}, error) {
	// とりあえず素朴に
	for _, f := range g.builtin_func_table {
		if f.name == funcname {
			if len(args) < f.min {
				return nil, fmt.Errorf("too few argument: function %s()", funcname)
			}
			if len(args) > f.max {
				return nil, fmt.Errorf("too many argument: function %s()", funcname)
			}
			if f.ox && g.ox == nil {
				return nil, fmt.Errorf("required OX server: function %s()", funcname)
			}
			if f.name == "help" {
				return funcHelp(g.builtin_func_table, f.name, args)
			} else {
				return f.f(g, f.name, args)
			}
		}
	}

	return nil, fmt.Errorf("unknown function: %s", funcname)
}

////////////////////////////////////////////////////////////
// 論理式
////////////////////////////////////////////////////////////
func funcNot(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("not(): unsupported for %v", args[0])
	}
	return f.Not(), nil
}

func funcAnd(g *Ganrac, name string, args []interface{}) (interface{}, error) {
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

func funcOr(g *Ganrac, name string, args []interface{}) (interface{}, error) {
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

func funcImpl(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(1st-arg): expected a first-order formula", name)
	}
	f1, ok := args[1].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(2nd-arg): expected a first-order formula", name)
	}

	return NewFmlOr(f0.Not(), f1), nil
}

func funcEquiv(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(1st-arg): expected a first-order formula", name)
	}
	f1, ok := args[1].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(2nd-arg): expected a first-order formula", name)
	}

	return NewFmlAnd(NewFmlOr(f0.Not(), f1), NewFmlOr(f0, f1.Not())), nil
}

func funcExists(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	return funcForEx(false, name, args)
}

func funcForAll(g *Ganrac, name string, args []interface{}) (interface{}, error) {
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

////////////////////////////////////////////////////////////
// OpenXM
////////////////////////////////////////////////////////////
func funcOXStr(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(*String)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected string: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}
	g.ox.PushOxCMO(f0.s)
	g.ox.PushOXCommand(SM_executeStringByLocalParser)
	s, err := g.ox.PopCMO()
	if err != nil {
		return nil, fmt.Errorf("%s(): popCMO failed %w", name, err)
	}
	gob := g.ox.toGObj(s)

	return gob, nil
}

func funcOXFunc(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(*String)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected string: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}

	err := g.ox.ExecFunction(f0.s, args[1:])
	if err != nil {
		return nil, fmt.Errorf("%s(): required OX server", name)
	}
	s, err := g.ox.PopCMO()
	if err != nil {
		return nil, fmt.Errorf("%s(): popCMO failed %w", name, err)
	}
	gob := g.ox.toGObj(s)

	return gob, nil
}

////////////////////////////////////////////////////////////
// OpenXM
////////////////////////////////////////////////////////////
func funcSleep(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[0].(*Int)
	if !ok {
		return nil, fmt.Errorf("%s() expected int", name)
	}
	if c.Sign() <= 0 {
		return nil, nil
	}

	v := c.Int64()
	time.Sleep(time.Millisecond * time.Duration(v))
	return nil, nil
}

func funcTime(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	return nil, nil
}

////////////////////////////////////////////////////////////
// poly
////////////////////////////////////////////////////////////
func funcSubst(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	if len(args)%2 != 1 {
		return nil, fmt.Errorf("%s() invalid args", name)
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
			return nil, fmt.Errorf("%s() invalid %d'th arg: %v", name, i+1, args[i])
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
			return nil, fmt.Errorf("%s() invalid %d'th arg", name, i+2)
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

	return nil, fmt.Errorf("%s() unsupported 1st arg", name)

}

func funcDeg(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	d, ok := args[1].(*Poly)
	if !ok || !d.isVar() {
		return nil, fmt.Errorf("%s(2nd arg): expected var: %v", name, args[1])
	}

	var deg int
	switch p := args[0].(type) {
	case Fof:
		deg = p.Deg(d.lv)
	case *Poly:
		deg = p.Deg(d.lv)
	case RObj:
		if p.IsZero() {
			deg = -1
		} else {
			deg = 0
		}
	default:
		return nil, fmt.Errorf("%s(1st arg): expected poly or FOF: %v", name, args[0])
	}

	return NewInt(int64(deg)), nil
}

func funcCoef(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	_, ok := args[0].(RObj)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected RObj: %v", name, args[0])
	}

	c, ok := args[1].(*Poly)
	if !ok || !c.isVar() {
		return nil, fmt.Errorf("%s(2nd arg): expected var: %v", name, args[1])
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

func funcRealRoot(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(): expected poly: %v", name, args[0])
	}
	return p.RealRootIsolation(1)
}

func funcRootBound(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(): expected poly: %v", name, args[0])
	}
	return p.RootBound()
}

func funcIndets(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	b := make([]bool, len(varlist))
	p, ok := args[0].(indeter)
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

////////////////////////////////////////////////////////////
// integer
////////////////////////////////////////////////////////////
func funcIGCD(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	a, ok := args[0].(*Int)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected int: %v", name, args[0])
	}
	b, ok := args[1].(*Int)
	if !ok {
		return nil, fmt.Errorf("%s(2nd arg): expected int: %v", name, args[1])
	}

	return a.Gcd(b), nil
}

////////////////////////////////////////////////////////////
// system
////////////////////////////////////////////////////////////

func funcLoad(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	return nil, fmt.Errorf("%s not implemented", name) // @TODO
}

func funcSave(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	return nil, fmt.Errorf("%s not implemented", name) // @TODO
}

func funcLen(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(lener)
	if !ok {
		return nil, fmt.Errorf("%s(): not supported: %v", name, args[0])
	}
	return NewInt(int64(p.Len())), nil
}

func funcHelp(builtin_func_table []func_table, name string, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return funcHelps(builtin_func_table, "@")
	}

	p, ok := args[0].(*String)
	if !ok {
		return nil, fmt.Errorf("%s(): required help(\"string\"):", name)
	}

	return funcHelps(builtin_func_table, p.s)
}

func funcHelps(builtin_func_table []func_table, name string) (interface{}, error) {
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
		fmt.Printf(" * ... required OX-server\n")
		fmt.Printf(" @ ... not implemented\n")
		fmt.Printf("\n")
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
