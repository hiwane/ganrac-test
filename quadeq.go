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
// 核
/////////////////////////////////////////////////
func qe_lineq(a *Atom, param interface{}) Fof {
	t := param.(*fof_quad_eq)
	if !a.hasVar(t.lv) {
		return a
	}
	res := make([]RObj, len(a.p))
	for i, p := range a.p {
		res[i] = t.g.ox.Resultant(p, t.p, t.lv)
	}
	op := a.op
	if t.sgn_lcp < 0 && a.Deg(t.lv)%2 != 0 {
		op = op.neg()
	}
	return NewAtoms(res, a.op)
}

func qe_quadeq(a *Atom, param interface{}) Fof {
	u := param.(*fof_quad_eq)
	if !a.hasVar(u.lv) {
		return a
	}
	f := u.p
	r := make([]RObj, len(a.p))
	for i, p := range a.p {
		r[i] = u.g.ox.Resultant(p, f, u.lv)
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
