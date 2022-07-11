package ganrac

import (
	"fmt"
)

// Quantifier Elimination for Formulas Constrained by Quadratic Equations via Slope Resultants
// Hoon Hong, The computer J., 1993

type fof_quad_eqer interface {
	// ax+b=0 && sgn a > 0 && fof
	qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof
}

type fof_quad_eq struct {
	p       *Poly
	sgn_lcp int // sign of lc(p)
	sgn_s   int // for quadratic case
	lv      Level
	g       *Ganrac
}

/////////////////////////////////////////////////
//
/////////////////////////////////////////////////
func quadeq_isEven(f Fof, lv Level) bool {
	stack := make([]Fof, 1)
	stack[0] = f
	for len(stack) > 0 {
		f = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch ff := f.(type) {
		case *Atom:
			if ff.op == EQ || ff.op == NE {
				continue
			}
			if ff.Deg(lv)%2 != 0 {
				return false
			}
		case FofAO:
			stack = append(stack, ff.Fmls()...)
		case FofQ:
			stack = append(stack, ff.Fml())
		}
	}
	return true
}

/////////////////////////////////////////////////
// 核
/////////////////////////////////////////////////
func qe_lineq(a *Atom, param interface{}) Fof {
	t := param.(*fof_quad_eq)
	if !a.hasVar(t.lv) {
		return a
	}
	res := make([]RObj, len(a.p))
	for i, p := range a.p {
		res[i] = t.g.ox.Resultant(t.p, p, t.lv)
	}
	op := a.op
	if t.sgn_lcp < 0 && a.Deg(t.lv)%2 != 0 {
		op = op.neg()
	}
	return NewAtoms(res, op)
}

func qe_quadeq(a *Atom, param interface{}) Fof {
	u := param.(*fof_quad_eq)
	if !a.hasVar(u.lv) {
		return a
	}
	f := u.p
	r := make([]RObj, len(a.p))
	for i, p := range a.p {
		r[i] = u.g.ox.Resultant(f, p, u.lv)
	}
	g := a.getPoly()
	aop := a.op
	if a.op == GE || a.op == LT {
		aop = aop.neg()
		g = g.Neg().(*Poly)
	}
	t := u.g.ox.Sres(f, g, u.lv, 0)
	s := u.g.ox.Sres(f, g, u.lv, 1)
	switch aop {
	case EQ, NE:
		opt := LE
		ops := LE
		if u.sgn_s < 0 {
			ops = GE
		}
		ret := NewFmlAnd(NewAtoms(r, EQ),
			NewFmlOr(
				NewFmlAnd(NewAtom(s, ops), NewAtom(t, opt.neg())),
				NewFmlAnd(NewAtom(t, opt), NewAtom(s, ops.neg()))))
		if a.op == NE {
			ret = ret.Not()
		}
		return ret
	case GT, LE:
		op := aop
		if u.sgn_lcp < 0 && a.Deg(u.lv)%2 != 0 {
			op = op.neg()
		}
		if a.op&EQ != 0 {
			op = op.not()
		}
		ops := op
		if u.sgn_s < 0 {
			ops = ops.neg()
		}
		sa := NewAtom(s, ops)
		ta := NewAtom(t, op)
		ret := newFmlOrs(
			NewFmlAnd(NewAtoms(r, op), ta),
			NewFmlAnd(NewAtoms(r, op.neg()), sa),
			NewFmlAnd(ta, sa))
		if a.op&EQ != 0 {
			ret = ret.Not()
		}
		return ret
	default:
		panic(fmt.Sprintf("op=%d", a.op))
	}
}

/////////////////////////////////////////////////
// 共通部分
/////////////////////////////////////////////////

func (fof *AtomT) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	return fof
}

func (fof *AtomF) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	return fof
}

func (fof *Atom) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	return fm(fof, p)
}

func (fof *FmlAnd) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	ret := make([]Fof, fof.Len())
	for i, q := range fof.Fmls() {
		ret[i] = q.qe_quadeq(fm, p)
	}
	return fof.gen(ret)
}

func (fof *FmlOr) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	ret := make([]Fof, fof.Len())
	for i, q := range fof.Fmls() {
		ret[i] = q.qe_quadeq(fm, p)
	}
	return fof.gen(ret)
}

func (fof *ForAll) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	fmt.Printf("forall %v\n", fof)
	panic("invalid......... qe_quadeq(forall)")
}

func (fof *Exists) qe_quadeq(fm func(a *Atom, p interface{}) Fof, p interface{}) Fof {
	fmt.Printf("exists %v\n", fof)
	panic("invalid......... qe_quadeq(exists)")
}

func (qeopt QEopt) qe_quadeq(fof FofQ, cond qeCond) Fof {
	var op OP
	if _, ok := fof.(*Exists); ok {
		op = EQ
	} else {
		op = NE
	}

	dmin := 1
	dmax := 2
	if qeopt.Algo&QEALGO_EQLIN == 0 {
		dmin = 2
	}
	if qeopt.Algo&QEALGO_EQQUAD == 0 {
		dmax = 1
	}

	fff, ok := fof.Fml().(FofAO)
	if !ok {
		return nil
	}
	var minatom struct {
		lv  Level
		a   *Atom
		p   *Poly
		z   RObj
		idx int
		deg int
		lc  bool // lc(a.p) is constant?
		uni bool // p is univariate
	}

	for ii, fffi := range fff.Fmls() {
		if atom, ok := fffi.(*Atom); ok && atom.op == op {
			poly := atom.getPoly()
			for _, q := range fof.Qs() {
				d := poly.Deg(q)
				if d < dmin || d > dmax {
					continue
				}
				z := poly.Coef(q, uint(d))
				_, lc := z.(NObj)
				univ := poly.isUnivariate()

				// 次数が低いか，主係数が定数なものを選択する
				if minatom.a == nil || minatom.deg > d ||
					(minatom.deg == d && univ) ||
					(minatom.deg == d && !minatom.uni && lc) ||
					(minatom.deg == d && !minatom.lc) {
					minatom.lv = q
					minatom.a = atom
					minatom.p = poly
					minatom.z = z
					minatom.idx = ii
					minatom.deg = d
					minatom.lc = lc
					minatom.uni = univ
				}
			}
		}
	}

	if minatom.a == nil {
		return nil
	}

	if op == NE {
		fff = fff.Not().(FofAO)
	}

	tbl := new(fof_quad_eq)
	tbl.g = qeopt.g
	tbl.p = minatom.p
	tbl.lv = minatom.lv

	if minatom.deg == 2 {
		// minatom.deg == 2
		even := quadeq_isEven(fff, minatom.lv)
		discrim := NewAtom(qeopt.g.ox.Discrim(minatom.p, minatom.lv), GE)
		qeopt.log(cond, 2, "eq2", "%v [%v] discrim=%v\n", fof, minatom.p, discrim)
		var o Fof = falseObj
		for _, sgns := range []struct {
			sgn_s int // 2つの根のうち，大きい方なら正.
			op    OP  // 主係数の符号
			skip  bool
		}{
			{+1, GT, false},
			{+1, LT, even},
			{-1, GT, false},
			{-1, LT, even},
		} {
			if sgns.skip {
				continue
			}
			tbl.sgn_s = sgns.sgn_s
			if sgns.op == GT {
				tbl.sgn_lcp = 1
			} else {
				tbl.sgn_lcp = -1
			}
			opp := newFmlAnds(fff.qe_quadeq(qe_quadeq, tbl), NewAtom(minatom.z, sgns.op), discrim)
			o = NewFmlOr(o, opp)
		}

		if _, ok := minatom.z.(NObj); ok {
			if op == NE {
				o = o.Not()
			}
			return o
		}

		eq := NewAtom(minatom.z, EQ)
		fml := qeopt.g.simplFof(fff, eq, falseObj) // 等式制約で簡単化
		fml = NewFmlAnd(fml, eq)
		o = NewFmlOr(o, fml)
		if op == NE {
			o = o.Not()
		}
		return o
	}

	qeopt.log(cond, 2, "eq1", "%v [%v]\n", fff, minatom.p)
	tbl.sgn_lcp = 1
	opos := NewFmlAnd(fff.qe_quadeq(qe_lineq, tbl), NewAtom(minatom.z, GT))

	tbl.sgn_lcp = -1
	oneg := NewFmlAnd(fff.qe_quadeq(qe_lineq, tbl), NewAtom(minatom.z, LT))

	fs := make([]Fof, len(fff.Fmls())+1)
	copy(fs, fff.Fmls())
	c := minatom.p.Coef(tbl.lv, 0)
	fs[minatom.idx] = NewAtom(c, EQ)
	fs[len(fs)-1] = NewFmlAnd(fs[minatom.idx], NewAtom(minatom.z, EQ))

	ret := newFmlOrs(opos, oneg, fff.gen(fs))
	if op == NE {
		ret = ret.Not()
	}
	return ret
}
