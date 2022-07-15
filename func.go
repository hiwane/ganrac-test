package ganrac

import (
	"fmt"
	"os"
	"time"
)

func (g *Ganrac) setBuiltinFuncTable() {
	// 関数テーブル
	g.builtin_func_table = []func_table{
		// sorted by name
		{"all", 2, 2, funcForAll, false, "([x], FOF):\t\tuniversal quantifier.", ""},
		//		{"and", 2, 2, funcAnd, false, "(FOF, ...):\t\tconjunction (&&)", ""},
		{"cad", 1, 2, funcCAD, true, "(FOF [, proj])*", ""},
		{"cadinit", 1, 1, funcCADinit, true, "(FOF)*", ""},
		{"cadlift", 1, 10, funcCADlift, true, "(CAD)*", ""},
		{"cadproj", 1, 2, funcCADproj, true, "(CAD [, proj])*", ""},
		{"cadsfc", 1, 1, funcCADsfc, true, "(CAD)*", ""},
		{"coef", 3, 3, funcCoef, false, "(poly, var, deg)", ""}, // coef(F, x, 2)
		{"deg", 2, 2, funcDeg, false, "(poly|FOF, var)\t\tdegree of a polynomial with respect to var", `
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
		{"discrim", 2, 2, funcOXDiscrim, true, "(poly)*\t\tdiscriminant.", `
Args
========
  poly: polynomial
  var : a variable in poly

Examples
========
  > discrim(2*x^2-3*x-3, x);
  33
  > discrim(a*x^2+b*x+c, x);
  -4*c*a+b^2
  > discrim(a*x^2+b*x+c, y);
  0
`},
		{"equiv", 2, 2, funcEquiv, false, "(fof1, fof2)\t\tfof1 is equivalent to fof2", ""},
		{"ex", 2, 2, funcExists, false, "(vars, FOF):\t\texistential quantifier.", `
Args
========
  vars: list of variables
  FOF : a first-order formula

Examples
========
  > ex([x], a*x^2+b*x+c == 0);
`},
		{"example", 0, 1, funcExample, false, "([name])\t\texample.", ""},
		{"fctr", 1, 1, funcOXFctr, true, "(poly)*\t\t\tfactorize polynomial over the rationals.", ""},
		{"gb", 2, 3, funcOXGB, true, "(polys, vars)*\t\tGroebner basis", ""},
		{"help", 0, 1, nil, false, "()\t\t\tshow help", ""},
		//		{"igcd", 2, 2, funcIGCD, false, "(int1, int2)\t\tThe integer greatest common divisor", ""},
		{"impl", 2, 2, funcImpl, false, "(fof1, fof2)\t\tfof1 impies fof2", ""},
		{"indets", 1, 1, funcIndets, false, "(mobj)\t\t\tfind indeterminates of an expression", ""},
		{"intv", 1, 3, funcIntv, false, "(lb, ub [, prec])\t\tmake an interval", ""},
		{"len", 1, 1, funcLen, false, "(mobj)\t\t\tlength of an object", ""},
		{"load", 2, 2, funcLoad, false, "(fname)@\t\t\tload file", ""},
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
		{"oxfunc", 2, 100, funcOXFunc, true, "(fname, args...)*\tcall ox-function by ox-asir", `
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
		{"oxstr", 1, 1, funcOXStr, true, "(str)*\t\t\tevaluate str by ox-asir", `
Args
========
str : string

Examples
========
  > oxstr("fctr(x^2-4);");
  [[1,1],[x-2,1],[x+2,1]]
`},
		{"print", 1, 10, funcPrint, false, "(obj [, kind, ...])\tprint object", `

Examples*
========
  > print(x^10+y-3);
  y+x^10-3
  > print(x^10+y-3, "tex");
  y+x^{10}-3
  > print(ex([x], x^2>1 && y +x == 0), "tex");
  \exists x x^2-1 > 0 \land y+x = 0

  > ` + init_var_funcname + `(a,b,c,x);
  > C = cadinit(ex([x], a*x^2+b*x+c==0));
  > cadproj(C);
  > print(C, "proj");
  > print(C, "proji");
  > cadlift(C);
  > print(C, "signatures");
  > print(C, "signatures", 1);
  > print(C, "cell", 1, 1);
  > print(C, "stat");
`},
		{"psc", 4, 4, funcOXPsc, true, "(poly, poly, var, int)*\tprincipal subresultant coefficient.", ""},
		{"qe", 1, 2, funcQE, true, "(fof [, opt])\t\treal quantifier elimination", fmt.Sprintf(`
Args
========
fof: first-order formula
opt: dictionary.
  %9s: linear    virtual substitution
  %9s: quadratic virtual substitution
  %9s: linear    equational constraint (Hong93)
  %9s: quadratic equational constraint (Hong93)
  %9s: inequational constraints (Iwane15)
  %9s: simplify  even formula
  %9s: simplify  homogeneous formula

Example
=======
  > vars(x, b, c);
  > F = ex([x], x^2+b*x+c == 0);
`,
			getQEoptStr(QEALGO_VSLIN),
			getQEoptStr(QEALGO_VSQUAD),
			getQEoptStr(QEALGO_EQLIN),
			getQEoptStr(QEALGO_EQQUAD),
			getQEoptStr(QEALGO_NEQ),
			getQEoptStr(QEALGO_SMPL_EVEN),
			getQEoptStr(QEALGO_SMPL_HOMO),
		)},
		{"quit", 0, 1, funcQuit, false, "([code])\t\tbye.", ""},
		{"realroot", 2, 2, funcRealRoot, false, "(uni-poly)\t\treal root isolation", ""},
		{"rootbound", 1, 1, funcRootBound, false, "(uni-poly in Z[x])\troot bound", `
Args
========
  poly: univariate polynomial

Examples
========
  > rootbound(x^2-2);
  3
`},
		{"save", 2, 3, funcSave, false, "(obj, fname)@\t\tsave object...", ""},
		{"simpl", 1, 2, funcSimplify, true, "(Fof)\t\t\tsimplify formula FoF", ""},
		{"sleep", 1, 1, funcSleep, false, "(milisecond)\t\tzzz", ""},
		// {"sqfr", 1, 1, funcSqfr, false, "(poly)* square-free factorization", ""},
		{"sres", 4, 4, funcOXSres, true, "(poly, poly, var, int)*\tslope resultant.", ""},
		{"subst", 1, 101, funcSubst, false, "(poly|FOF|List,x,vx,y,vy,...)", ""},
		{"time", 1, 1, funcTime, false, "(expr)\t\t\trun command and system resource usage", ""},
		{init_var_funcname, 0, 0, nil, false, "(var, ...)\t\tinit variable order", `
Args
========
  var

Examples*
========
  > ` + init_var_funcname + `(x,y,z);
  > F = x^2+2;
  > F;
  > ` + init_var_funcname + `(a,b,c);
  > F;
  a^2+2
  > x;
  error: undefined variable ` + "`x`\n"},
		{"verbose", 1, 2, funcVerbose, false, "(int [, int])\t\tset verbose level", ""},
		{"vs", 1, 1, funcVS, true, "(FOF)* ", ""},
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

func funcImpl(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(1st-arg): expected a first-order formula", name)
	}
	f1, ok := args[1].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(2nd-arg): expected a first-order formula", name)
	}

	return FofImpl(f0, f1), nil
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

	return FofEquiv(f0, f1), nil
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

	err := g.ox.ExecFunction(f0.s, args[1:]...)
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

func funcOXDiscrim(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[1].(*Poly)
	if !ok || !c.isVar() {
		return nil, fmt.Errorf("%s(2nd arg): expected var: %v", name, args[1])
	}

	switch p := args[0].(type) {
	case *Poly:
		return g.ox.Discrim(p, c.lv), nil
	case NObj:
		return zero, nil
	default:
		return nil, fmt.Errorf("%s(1st arg): expected poly: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}
}

func funcOXFctr(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected poly: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}

	return g.ox.Factor(f0), nil
}

func funcOXGB(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f0, ok := args[0].(*List)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected poly-list: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}

	f1, ok := args[1].(*List)
	if !ok {
		return nil, fmt.Errorf("%s(2nd arg): expected var-list: %d:%v", name, args[1].(GObj).Tag(), args[1])
	}

	n := 0
	if len(args) == 3 {
		f2, ok := args[2].(*Int)
		if !ok || f2.Sign() < 0 || !f2.IsInt64() {
			return nil, fmt.Errorf("%s(3rd arg): expected nonnegint: %d:%v", name, args[2].(GObj).Tag(), args[2])
		}
		n = int(f2.Int64())
	}

	return g.ox.GB(f0, f1, n), nil
}

func funcOXPsc(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected poly: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}
	h, ok := args[1].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(2nd arg): expected poly: %d:%v", name, args[1].(GObj).Tag(), args[1])
	}

	x, ok := args[2].(*Poly)
	if !ok || !x.isVar() {
		return nil, fmt.Errorf("%s(3rd arg): expected var: %d:%v", name, args[2].(GObj).Tag(), args[2])
	}

	j, ok := args[3].(*Int)
	if !ok || !j.IsInt64() || j.Sign() < 0 {
		return nil, fmt.Errorf("%s(4th arg): expected nonnegint: %v", name, args[3])
	}

	return g.ox.Psc(f, h, x.lv, int32(j.Int64())), nil
}

func funcOXSres(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	f, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected poly: %d:%v", name, args[0].(GObj).Tag(), args[0])
	}
	h, ok := args[1].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(2nd arg): expected poly: %d:%v", name, args[1].(GObj).Tag(), args[1])
	}

	x, ok := args[2].(*Poly)
	if !ok || !x.isVar() {
		return nil, fmt.Errorf("%s(3rd arg): expected var: %d:%v", name, args[2].(GObj).Tag(), args[2])
	}

	j, ok := args[3].(*Int)
	if !ok || !j.IsInt64() || j.Sign() < 0 {
		return nil, fmt.Errorf("%s(4th arg): expected nonnegint: %v", name, args[3])
	}

	return g.ox.Sres(f, h, x.lv, int32(j.Int64())), nil
}

////////////////////////////////////////////////////////////
// CAD
////////////////////////////////////////////////////////////
func funcExample(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	str := ""
	if len(args) == 1 {
		c, ok := args[0].(*String)
		if !ok {
			return nil, fmt.Errorf("%s() expected string", name)
		}
		str = c.s
	}

	ex := GetExampleFof(str)
	if ex == nil {
		if str == "" {
			return nil, nil
		}
		return nil, fmt.Errorf("%s() invalid 1st arg %s", name, str)
	}

	ll := NewList()
	ll.Append(ex.Input)
	ll.Append(ex.Output)
	ll.Append(NewString(ex.Ref))
	ll.Append(NewString(ex.DOI))

	return ll, nil
}

func funcSimplify(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s() expected FOF", name)
	}

	return g.simplFof(c, trueObj, falseObj), nil
}

func funcCAD(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	fof, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s() expected FOF", name)
	}
	var algo ProjectionAlgo = PROJ_McCallum
	if len(args) > 1 {
		algoi, ok := args[1].(*Int)
		if !ok || (!algoi.IsZero() && !algoi.IsOne()) {
			return nil, fmt.Errorf("%s(2nd-arg) expected proj operator", name)
		}
		algo = ProjectionAlgo(algoi.Int64())
	}
	switch fof.(type) {
	case *AtomT, *AtomF:
		return fof, nil
	}

	cad, err := NewCAD(fof, g)
	if err != nil {
		return nil, err
	}
	_, err = cad.Projection(algo)
	if err != nil {
		return nil, err
	}
	err = cad.Lift()
	if err != nil {
		return nil, err
	}
	return cad.Sfc()
}

func funcCADinit(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s() expected FOF", name)
	}

	return NewCAD(c, g)
}

func funcCADproj(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[0].(*CAD)
	if !ok {
		return nil, fmt.Errorf("%s() expected CAD generated by cadinit()", name)
	}

	var algo ProjectionAlgo = PROJ_McCallum
	if len(args) > 1 {
		algoi, ok := args[1].(*Int)
		if !ok || (!algoi.IsZero() && !algoi.IsOne()) {
			return nil, fmt.Errorf("%s(2nd-arg) expected proj operator", name)
		}
		algo = ProjectionAlgo(algoi.Int64())
	}

	p, err := c.Projection(algo)
	return p, err
}

func funcCADlift(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[0].(*CAD)
	if !ok {
		return nil, fmt.Errorf("%s(1st-arg) expected CAD generated by cadinit()", name)
	}

	index := make([]int, len(args)-1)
	for i := 1; i < len(args); i++ {
		v, ok := args[i].(*Int)
		if !ok || !v.IsInt64() || v.Int64() > 10000000 || v.Int64() < -1 {
			return nil, fmt.Errorf("%s(%dth arg) expected index", name, i)
		}
		index[i-1] = int(v.Int64())
	}

	err := c.Lift(index...)
	return c, err
}

func funcCADsfc(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	cad, ok := args[0].(*CAD)
	if !ok {
		return nil, fmt.Errorf("%s(1st-arg) expected CAD generated by cadinit()", name)
	}

	return cad.Sfc()
}

func funcVS(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	fof, ok := args[0].(FofQ)
	if !ok || !fof.Fml().IsQff() {
		return nil, fmt.Errorf("%s() expected prenex-FOF", name)
	}

	var fml Fof
	fml = fof
	for _, q := range fof.Qs() {
		fml = vsLinear(fml, q)
	}
	return fml, nil
}

////////////////////////////////////////////////////////////
// util
////////////////////////////////////////////////////////////
func funcPrint(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	switch cc := args[0].(type) {
	case *CAD:
		return nil, cc.Print(args[1:]...)
	case fmt.Formatter:
		if len(args) > 2 {
			return nil, fmt.Errorf("invalid # of arg")
		}
		t := "org"
		if len(args) == 2 {
			s, ok := args[1].(*String)
			if !ok {
				return nil, fmt.Errorf("invalid 2nd arg")
			}
			t = s.s
		}
		switch t {
		case "org":
			fmt.Printf("%v\n", cc)
		case "tex":
			fmt.Printf("%P\n", cc)
		case "src":
			fmt.Printf("%S\n", cc)
		case "dump":
			fmt.Printf("%V\n", cc)
		case "qepcad":
			fmt.Printf("%Q\n", cc)
		default:
			fmt.Printf(t, cc)
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("print(): unsupported object is specified")
	}
}

func funcQuit(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	code := 0
	if len(args) > 0 {
		c, ok := args[0].(*Int)
		if !ok {
			return nil, fmt.Errorf("%s() expected int", name)
		}
		if c.Sign() < 0 || !c.IsInt64() || c.Int64() > 125 {
			return nil, fmt.Errorf("%s() expected integer in the range [0, 125]", name)
		}
		code = int(c.Int64())
	}
	os.Exit(code)
	return nil, nil
}

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

func funcVerbose(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	c, ok := args[0].(*Int)
	if !ok || !c.IsInt64() {
		return nil, fmt.Errorf("%s(1st arg) expected int", name)
	}

	if len(args) > 1 {
		d, ok := args[1].(*Int)
		if !ok || !c.IsInt64() {
			return nil, fmt.Errorf("%s(2nd arg) expected int", name)
		}
		if d.Sign() <= 0 {
			g.verbose_cad = 0
		} else {
			g.verbose_cad = int(d.Int64())
		}
	}

	if c.Sign() <= 0 {
		g.verbose = 0
	} else {
		g.verbose = int(c.Int64())
	}
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

		rlv[j].lv = p.lv

		v, ok := args[i+1].(RObj)
		if !ok {
			return nil, fmt.Errorf("%s() invalid %d'th arg", name, i+2)
		}
		rlv[j].r = v
		j++
	}
	rlv = rlv[:j]

	o := args[0].(GObj)
	for _, r := range rlv {
		o = gobjSubst(o, r.r, r.lv)
	}

	return o, nil
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

func funcArgBoolVal(val GObj) bool {
	switch v := val.(type) {
	case RObj:
		return !v.IsZero()
	case *String:
		return v.s == "yes" || v.s == "Y" || v.s == "y"
	default:
		return false
	}
}

func funcQE(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	fof, ok := args[0].(Fof)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected Fof: %v", name, args[0])
	}
	opt := NewQEopt()
	if len(args) > 1 {
		dic, ok := args[1].(*Dict)
		if !ok {
			return nil, fmt.Errorf("%s(2nd arg): expected Dict: %v", name, args[1])
		}
		for k, v := range dic.v {
			switch k {
			case getQEoptStr(QEALGO_EQQUAD):
				opt.SetAlgo(QEALGO_EQQUAD, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_EQLIN):
				opt.SetAlgo(QEALGO_EQLIN, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_VSQUAD):
				opt.SetAlgo(QEALGO_VSQUAD, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_VSLIN):
				opt.SetAlgo(QEALGO_VSLIN, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_NEQ):
				opt.SetAlgo(QEALGO_NEQ, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_SMPL_EVEN):
				opt.SetAlgo(QEALGO_SMPL_EVEN, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_SMPL_HOMO):
				opt.SetAlgo(QEALGO_SMPL_HOMO, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_SMPL_TRAN):
				opt.SetAlgo(QEALGO_SMPL_TRAN, funcArgBoolVal(v))
			case getQEoptStr(QEALGO_SMPL_ROTA):
				opt.SetAlgo(QEALGO_SMPL_ROTA, funcArgBoolVal(v))
			case "verbose":
				if val, ok := v.(*Int); ok && val.IsInt64() {
					opt.log_level = int(val.Int64())
				} else {
					return nil, fmt.Errorf("%s(3rd arg): invalid option value: %s: %v.", name, k, v)
				}
			default:
				return nil, fmt.Errorf("%s(3rd arg): unknown option: %s", name, k)
			}
		}
	}

	return g.QE(fof, opt), nil
}

func funcRealRoot(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	p, ok := args[0].(*Poly)
	if !ok {
		return nil, fmt.Errorf("%s(): expected poly: %v", name, args[0])
	}

	q, ok := args[1].(*Int)
	if !ok {
		return nil, fmt.Errorf("%s(): expected int: %v", name, args[1])
	}

	return p.RealRootIsolation(int(q.Int64()))
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
	if !ok {
		return NewList(), nil
	}
	p.Indets(b)

	ret := make([]interface{}, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] {
			ret = append(ret, NewPolyVar(Level(i)))
		}
	}
	return NewList(ret...), nil
}

func funcIntv(g *Ganrac, name string, args []interface{}) (interface{}, error) {
	prec := uint(53)
	if g, ok := args[0].(GObj); ok && len(args) == 1 {
		return gobjToIntv(g, prec), nil
	}

	a, ok := args[0].(NObj)
	if !ok {
		return nil, fmt.Errorf("%s(1st arg): expected number: %v", name, args[0])
	}
	b := a
	if len(args) > 1 {
		bb, ok := args[1].(NObj)
		if !ok {
			return nil, fmt.Errorf("%s(2nd arg): expected number: %v", name, args[1])
		}
		if _, ok = a.(NObj); !ok {
			return nil, fmt.Errorf("%s(1st arg): expected number: %v", name, args[0])
		}
		b = bb
	}
	if len(args) > 2 {
		pp, ok := args[2].(*Int)
		if !ok || !pp.IsInt64() || pp.Sign() <= 0 {
			return nil, fmt.Errorf("%s(3rd arg): expected int: %v", name, args[2])
		}

		prec = uint(pp.Int64())
	}
	aa := a.toIntv(prec)
	if aintv, ok := aa.(*Interval); ok {
		bintv, ok := b.toIntv(prec).(*Interval)
		if !ok {
			return nil, fmt.Errorf("%s(2nd arg): expected number: %v", name, args[1])
		}
		u := newInterval(prec)
		u.inf = aintv.inf
		u.sup = bintv.sup
		return u, nil
	} else {
		return aa, nil
	}
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
		fmt.Printf("  > %s(x, y, z);  # init variable order.\n", init_var_funcname)
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
