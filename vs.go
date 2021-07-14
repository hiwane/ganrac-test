package ganrac

// Applying Linear Quantifier Elimination
// R. Loos, V. Weispfenning 1993

// A Generalized Framework for Virtual Substitution
// M. Kosta, T. Sturm 2015

import (
	"fmt"
)

type fof_vser interface {
	apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof
}

type vslin_sample_point struct {
	num    RObj
	den    []RObj // [den^0, den, den^2, den^3, ...]
	densgn int
	lv     Level
}

type vs_elimination_set struct {
	equ []*Poly // <=, >=, ==
	ine []*Poly // <, >, !=
	lv  Level
}

func newVsEliminationSet(lv Level) *vs_elimination_set {
	p := new(vs_elimination_set)
	p.equ = make([]*Poly, 0)
	p.ine = make([]*Poly, 0)
	p.lv = lv
	return p
}

func (es *vs_elimination_set) exists(atom *Atom, pset []*Poly) bool {
	if err := atom.valid(); err != nil {
		fmt.Printf("atom=%v\n", atom)
		panic("invalid atom")
	}
	for i, p := range atom.p {
		for _, qq := range pset {
			if qq.Equals(p) {
				atom.p[i] = qq
				return true
			}
		}
	}
	return false
}

func (es *vs_elimination_set) addAtom(atom *Atom) {
	for _, pol := range atom.p {
		if pol.hasVar(es.lv) {
			if (atom.op & EQ) == 0 {
				if !es.exists(atom, es.ine) {
					es.ine = append(es.ine, pol)
				}
			} else {
				if !es.exists(atom, es.equ) {
					es.equ = append(es.equ, pol)
				}
			}
		}
	}
}

func (fof *AtomT) apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof {
	return fof
}
func (fof *AtomF) apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof {
	return fof
}

func (fof *Atom) apply_vs(fm func(atom *Atom, p interface{}) Fof, ptt interface{}) Fof {
	pt := ptt.(*vslin_sample_point)
	if fof.hasVar(pt.lv) {
		return fm(fof, pt)
	} else {
		return fof

	}
}

func (fof *FmlAnd) apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof {
	var r Fof = trueObj

	for _, f := range fof.fml {
		r = NewFmlAnd(r, f.apply_vs(fm, p))
	}

	return r
}

func (fof *FmlOr) apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof {
	var r Fof = falseObj

	for _, f := range fof.fml {
		r = NewFmlOr(r, f.apply_vs(fm, p))
	}

	return r
}

func (fof *ForAll) apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof {
	fmt.Printf("forall %v\n", fof)
	panic("invalid......... apply_vs(forall)")
}

func (fof *Exists) apply_vs(fm func(atom *Atom, p interface{}) Fof, p interface{}) Fof {
	fmt.Printf("exists %v\n", fof)
	panic("invalid......... apply_vs(exists)")
}

func gen_sample_vslin(p *Poly, lv Level, fml Fof) *vslin_sample_point {
	d := fml.Deg(lv)
	sp := new(vslin_sample_point)
	sp.lv = lv
	sp.num = p.Coef(lv, 0)
	den := p.Coef(lv, 1)
	if den.IsNumeric() {
		sp.densgn = 1
		if den.Sign() < 0 {
			den = den.Neg()
		} else {
			sp.num = sp.num.Neg()
		}
	} else {
		if den.Sign() < 0 {
			den = den.Neg()
		} else {
			sp.num = sp.num.Neg()
		}
		sp.densgn = 0
	}

	sp.den = make([]RObj, d+1)
	sp.den[0] = one
	sp.den[1] = den
	for i := 2; i <= d; i++ {
		sp.den[i] = sp.den[i-1].Mul(den)
	}
	return sp
}

func virtual_subst_lin(atom *Atom, ptt interface{}) Fof {
	// 線形なサンプル点の代入
	// 分母は非ゼロと仮定し，
	// その符号は pt.densgn
	pt := ptt.(*vslin_sample_point)
	pp := make([]RObj, len(atom.p))

	op := atom.op
	if pt.densgn < 0 && atom.Deg(pt.lv)%2 != 0 {
		op = op.neg()
	}
	for i, p := range atom.p {
		d := p.Deg(pt.lv)
		pp[i] = p.subst_frac(pt.num, pt.den[:d+1], pt.lv)
		if err := pp[i].valid(); err != nil {
			panic(err)
		}
		//			fmt.Printf("pp[%d]=%v\n", i, pp[i])
	}
	//		fmt.Printf("virtual_subst_lin() %v -> %v<%d>::%v\n", atom, pp, atom.op, NewAtoms(pp, atom.op))
	return NewAtoms(pp, op)
}

func vs_nu(polys []*Poly, op OP, pt *vslin_sample_point) Fof {
	// pt + epsilon を代入する
	d := 0
	if pt.densgn < 0 {
		for _, p := range polys {
			d += p.Deg(pt.lv)
		}
	}

	var f1 Fof
	if d%2 == 0 || true {
		// 偶数次数か，分母の符号が正ならそのまま.
		f1 = virtual_subst_lin(newAtoms(polys, op), pt)
	} else {
		f1 = virtual_subst_lin(newAtoms(polys, op.neg()), pt)
	}
	// fmt.Printf("vsnu(): f1=%v\n", f1)
	if err := f1.valid(); err != nil { // debug
		fmt.Printf("%V\ninvalid f1 %v\n", f1, f1)
		panic(err.Error())
	}
	var f2 Fof = virtual_subst_lin(newAtoms(polys, EQ), pt)
	// fmt.Printf("vsnu(): f2=%v\n", f2)
	if err := f2.valid(); err != nil { // debug
		fmt.Printf("%V\ninvalid f2 %v\n", f1, f1)
		panic(err.Error())
	}

	ps := make([]*Poly, 0, len(polys))
	var pmul *Poly = nil
	for i, p := range polys {
		if !p.hasVar(pt.lv) {
			ps = append(ps, p)
		} else if pmul == nil {
			pmul = polys[i]
		} else {
			pmul = pmul.Mul(polys[i]).(*Poly)
		}
	}

	var v RObj
	if pmul == nil {
		v = one
	} else {
		v = pmul.diff(pt.lv)
	}
	// fmt.Printf("vsnu(): v=%v: %v\n", v, v.IsNumeric())

	if v.IsNumeric() {
		rs := make([]RObj, len(ps))
		for i := 0; i < len(ps); i++ {
			rs[i] = ps[i]
		}
		// fmt.Printf("vsnu(): rs=%v, %d\n", rs, v.Sign())
		if v.Sign() > 0 {
			return NewFmlOr(f1, NewFmlAnd(f2, NewAtoms(rs, op)))
		} else {
			return NewFmlOr(f1, NewFmlAnd(f2, NewAtoms(rs, op.neg())))
		}
	} else {
		ps = append(ps, v.(*Poly))
		sfml := vs_nu(ps, op, pt)
		// fmt.Printf("vsnu(): sfml=%v\n", sfml)
		return NewFmlOr(f1, NewFmlAnd(f2, sfml))
	}
}

func virtual_subst_lin_e(atom *Atom, ptt interface{}) Fof {
	// ptt+ infinitesimal を代入する

	if atom.op == EQ {
		var ret Fof = falseObj
		pt := ptt.(*vslin_sample_point)
		for _, p := range atom.p {
			var pi Fof = trueObj
			d := p.Deg(pt.lv)
			for i := 0; i <= d; i++ {
				c := p.Coef(pt.lv, uint(i))
				pi = NewFmlAnd(pi, NewAtom(c, EQ))
			}
			ret = NewFmlOr(ret, pi)
		}
		return ret
	} else if atom.op == NE {
		return virtual_subst_lin_e(newAtoms(atom.p, EQ), ptt).Not()
	} else if atom.op == LT || atom.op == GT {
		pt := ptt.(*vslin_sample_point)
		if err := atom.valid(); err != nil {
			fmt.Printf("%V\ninvalid atom %v\n", atom, atom)
			panic(err.Error())
		}
		ret := vs_nu(atom.p, atom.op, pt)
		return ret
	} else if atom.op == LE || atom.op == GE {
		return virtual_subst_lin_e(newAtoms(atom.p, atom.op.not()), ptt).Not()
	} else {
		panic("invalid op")
	}
}

func vs_mu(polys []*Poly, pt *vslin_sample_point) Fof {
	// polys < 0 の atom に対して -infty を代入する

	pmul := polys[0]
	for i := 1; i < len(polys); i++ {
		pmul = pmul.Mul(polys[i]).(*Poly)
	}

	d := pmul.Deg(pt.lv)

	coeffs := make([]RObj, d+1)
	for i := 0; i <= d; i++ {
		coeffs[i] = pmul.Coef(pt.lv, uint(i))
	}

	var ret Fof = falseObj
	for i := 0; i <= d; i++ {
		var u Fof
		if i%2 == 0 {
			u = NewAtom(coeffs[i], LT)
		} else {
			u = NewAtom(coeffs[i], GT)
		}
		for j := i + 1; j <= d; j++ {
			u = NewFmlAnd(u, NewAtom(coeffs[j], EQ))
		}
		ret = NewFmlOr(ret, u)
	}
	return ret
}

func virtual_subst_lin_i(atom *Atom, ptt interface{}) Fof {
	// -inf を代入する.
	switch atom.op {
	case EQ:
		var ret Fof = falseObj
		pt := ptt.(*vslin_sample_point)
		for _, p := range atom.p {
			var pi Fof = trueObj
			d := p.Deg(pt.lv)
			for i := 0; i <= d; i++ {
				c := p.Coef(pt.lv, uint(i))
				pi = NewFmlAnd(pi, NewAtom(c, EQ))
			}
			ret = NewFmlOr(ret, pi)
		}
		return ret
	case NE:
		return virtual_subst_lin_i(newAtoms(atom.p, EQ), ptt).Not()
	case LT:
		pt := ptt.(*vslin_sample_point)
		return vs_mu(atom.p, pt)
	case LE:
		return NewFmlOr(
			virtual_subst_lin_i(newAtoms(atom.p, EQ), ptt),
			virtual_subst_lin_i(newAtoms(atom.p, LT), ptt))
	case GT:
		return NewFmlOr(
			virtual_subst_lin_i(newAtoms(atom.p, EQ), ptt),
			virtual_subst_lin_i(newAtoms(atom.p, LT), ptt)).Not()
	case GE:
		return virtual_subst_lin_i(newAtoms(atom.p, LT), ptt).Not()
	default:
		panic("invalid op")
	}
}

func get_vs_polys(p Fof, lv Level) *vs_elimination_set {
	// algorithm 2 is not implemented
	// p に含まれる lv変数を含む多項式のリスト.
	peqlt := newVsEliminationSet(lv)

	pstack := make([]Fof, 1)
	pstack[0] = p
	for len(pstack) > 0 {
		qq := pstack[len(pstack)-1]
		switch q := qq.(type) {
		case *ForAll:
			pstack[len(pstack)-1] = q.fml
		case *Exists:
			pstack[len(pstack)-1] = q.fml
		case *Atom:
			pstack = pstack[:len(pstack)-1]
			peqlt.addAtom(q)
		case *FmlAnd:
			pstack[len(pstack)-1] = q.fml[0]
			for i := 1; i < len(q.fml); i++ {
				pstack = append(pstack, q.fml[i])
			}
		case *FmlOr:
			pstack[len(pstack)-1] = q.fml[0]
			for i := 1; i < len(q.fml); i++ {
				pstack = append(pstack, q.fml[i])
			}
		}
	}
	return peqlt
}

func vsLinear(fof Fof, lv Level) Fof {
	var fml Fof
	switch pp := fof.(type) {
	case *ForAll:
		fml = pp.fml.Not()
	case *Exists:
		fml = pp.fml
	default:
		return fof
	}

	if fml.Deg(lv) != 1 {
		return fof
	}
	if !fml.IsQff() {
		return fof
	}

	var ret Fof = falseObj
	elset := get_vs_polys(fml, lv)
	// fmt.Printf("vsLin=%v\n", fml)
	// fmt.Printf("peq[%d]=%v\n", len(elset.equ), elset.equ)
	// fmt.Printf("plt[%d]=%v\n", len(elset.ine), elset.ine)
	required_zero := true
	for _, pp := range elset.equ {
		pt := gen_sample_vslin(pp, lv, fml)
		sgn := pt.densgn
		if sgn != 0 {
			required_zero = false
		}

		if sgn >= 0 {
			pt.densgn = 1
			sfml := fml.apply_vs(virtual_subst_lin, pt)
			// fmt.Printf("add2:+[%v]/[%v]: %v\n", pt.num, pt.den[1], sfml)
			if err := sfml.valid(); err != nil {
				panic(err)
			}
			ret = NewFmlOr(ret, NewFmlAnd(sfml, NewAtom(pt.den[1], GT)))
		}

		if sgn <= 0 {
			pt.densgn = -1
			sfml := fml.apply_vs(virtual_subst_lin, pt)
			// fmt.Printf("add3:-[%v]/[%v]: %v\n", pt.num, pt.den[1], sfml)
			if err := sfml.valid(); err != nil {
				panic(err)
			}
			ret = NewFmlOr(ret, NewFmlAnd(sfml, NewAtom(pt.den[1], LT)))
		}
	}
	if len(elset.ine) > 0 {
		var pt *vslin_sample_point
		required_minf := true
		for _, pp := range elset.ine {
			pt = gen_sample_vslin(pp, lv, fml)
			sgn := pt.densgn
			if sgn <= 0 {
				required_minf = true
			}
			if sgn != 0 {
				required_zero = false
			}

			if sgn >= 0 {
				pt.densgn = 1
				// fmt.Printf("add5:+e[%v]/[%v]: >>>>\n", pt.num, pt.den[1])
				sfml := fml.apply_vs(virtual_subst_lin_e, pt)
				// fmt.Printf("add5:+e[%v]/[%v]: %v\n", pt.num, pt.den[1], sfml)
				if err := sfml.valid(); err != nil {
					panic(err)
				}
				ret = NewFmlOr(ret, NewFmlAnd(sfml, NewAtom(pt.den[1], GT)))
				if err := ret.valid(); err != nil {
					panic(err)
				}
			}
			if sgn <= 0 {
				pt.densgn = -1
				// fmt.Printf("add6:-e[%v]/[%v]: >>>>\n", pt.num, pt.den[1])
				sfml := fml.apply_vs(virtual_subst_lin_e, pt)
				// fmt.Printf("add6:-e[%v]/[%v]: %v\n", pt.num, pt.den[1], sfml)
				if err := sfml.valid(); err != nil {
					panic(err)
				}
				ret = NewFmlOr(ret, NewFmlAnd(sfml, NewAtom(pt.den[1], LT)))
				if err := ret.valid(); err != nil {
					panic(err)
				}
			}
		}
		if err := ret.valid(); err != nil {
			panic(err)
		}
		if required_minf {
			sfml := fml.apply_vs(virtual_subst_lin_i, pt)
			// fmt.Printf("-inf] %v\n", sfml)
			if err := sfml.valid(); err != nil {
				panic(err)
			}
			// fmt.Printf("before\nret%x= %v\nsfm%x= %v\n", ret.fofTag(), ret, sfml.fofTag(), sfml)
			ret = NewFmlOr(ret, sfml)
			if err := ret.valid(); err != nil {
				fmt.Printf("ret=%v\n", ret)
				ppp, ok := ret.(*FmlOr)
				if ok {
					fmt.Printf("len=%d\n", ppp.Len())
				}

				panic(err)
			}
		}
	}
	if required_zero {
		ret = NewFmlOr(ret, fml.Subst(zero, lv))
		if err := ret.valid(); err != nil {
			panic(err)
		}
	}

	if q, ok := fof.(*ForAll); ok {
		ret = ret.Not()
		ret = NewQuantifier(true, q.q, ret)
	} else if q, ok := fof.(*Exists); ok {
		ret = NewQuantifier(false, q.q, ret)
	}

	// fmt.Printf("LinearVS ret=%v\n", ret)
	return ret
}

func (qeopt QEopt) qe_vslin(fof FofQ, cond qeCond) Fof {
	for _, q := range fof.Qs() {
		qeopt.log(cond, 2, "qevs1", "<%s> %v\n", varstr(q), fof)
		ff := vsLinear(fof, q)
		if ff != fof {
			return ff
		}
	}
	return nil
}
