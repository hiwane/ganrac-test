package ganrac

// Quantifier elimination for inequational constraints.
// Hidenao IWANE. 2015 (Japanese)

// ex([x], f(x) != 0 && phi(x))

func is_neq_only(fof Fof, lv Level) bool {
	switch pp := fof.(type) {
	case FofQ:
		return is_neq_only(pp.Fml(), lv)
	case FofAO:
		for _, f := range pp.Fmls() {
			if !is_neq_only(f, lv) {
				return false
			}
		}
		return true
	case *Atom:
		if pp.op == NE {
			return true
		}
		return !pp.hasVar(lv)
	}
	return false
}

func is_strict_only(fof Fof, lv Level) bool {
	switch pp := fof.(type) {
	case FofQ:
		return is_strict_only(pp.Fml(), lv)
	case FofAO:
		for _, f := range pp.Fmls() {
			if !is_strict_only(f, lv) {
				return false
			}
		}
		return true
	case *Atom:
		if pp.op == NE || pp.op == LT || pp.op == GT {
			return true
		}
		return !pp.hasVar(lv)
	}
	return false
}

/*
 * 非等式制約部分とそれ以外で分割する
 */
func divide_neq(finput Fof, lv Level, qeopt QEopt) (Fof, Fof) {

	switch fof := finput.(type) {
	case *Atom:
		if qeopt.assert && fof.Deg(lv) == 0 {
			panic("lv not found")
		}
		if fof.op == NE {
			return fof, trueObj
		} else {
			return trueObj, fof
		}
	case *FmlAnd:
		fne := make([]Fof, 0, len(fof.Fmls()))
		fot := make([]Fof, 0, len(fof.Fmls()))
		for _, f := range fof.Fmls() {
			if is_neq_only(f, lv) {
				fne = append(fne, f)
			} else {
				fot = append(fot, f)
			}
		}
		if len(fne) == 0 {
			return trueObj, fof
		} else if len(fot) == 0 {
			return fof, trueObj
		} else {
			return newFmlAnds(fne...), newFmlAnds(fot...)
		}
	}
	return nil, nil
}

func apply_neqQE(fof Fof, lv Level) Fof {
	switch pp := fof.(type) {
	case FofQ:
		return pp.gen(pp.Qs(), apply_neqQE(pp.Fml(), lv))
	case FofAO:
		fmls := pp.Fmls()
		ret := make([]Fof, len(fmls))
		for i, f := range fmls {
			ret[i] = apply_neqQE(f, lv)
		}
		return pp.gen(ret)
	case *Atom:
		if pp.op != NE || !pp.hasVar(lv) {
			return pp
		}
		var ret Fof = trueObj
		for _, p := range pp.p {
			var r Fof = falseObj
			for d := p.Deg(lv); d >= 0; d-- {
				r = NewFmlOr(r, NewAtom(p.Coef(lv, uint(d)), NE))
			}
			ret = NewFmlAnd(ret, r)
		}
		return ret

	}
	return nil
}

/*
 * fof: inequational constraints
 * atom: f <= 0 or f >= 0: f is univariate.
 *
 * Returns: qff which is equivalent to ex([x], f <= 0 && fof_neq)
 :        : nil if fails
*/
func apply_neqQE_atom_univ(fof, qffneq Fof, atom *Atom, lv Level, qeopt QEopt, cond qeCond) Fof {
	// atom.p は univariate
	// qffneq := apply_neqQE(fof, lv)
	p := atom.getPoly()

	// ex([x], sgn * f >= 0) がわかった
	if p.Sign() > 0 && atom.op == GE || p.Sign() < 0 && atom.op == LE {
		return qffneq
	}

	bak_deg := p.Deg(lv)
	if bak_deg % 2 != 0 {
		return qffneq
	}

	ps := qeopt.g.ox.Factor(p)
	evens := make([]*Poly, 0, ps.Len())

	for i := 1; i < ps.Len(); i++ {
		_fr, _ := ps.Geti(i) // f^r
		fr := _fr.(*List)
		f := fr.getiPoly(0)
		r := fr.getiInt(1)

		rr, _ := f.RealRootIsolation(1)
		if rr.Len() == 0 {
			// ゼロ点がない => 符号一定
			continue
		}

		if r.Bit(0) == 0 { // r % 2 == 0
			evens = append(evens, f)
			continue // 有限個のゼロ点以外は符号を変えないので無視
		} else {
			// 符号が変化する区間があることが確定
			return qffneq
		}
	}

	if len(evens) == 0 {
		return falseObj
	}

	// 有限個のゼロ点でのみ条件を満たすことがわかった => 等式制約 QE へ.
	var ret Fof = falseObj
	for _, z := range evens {
		f := z
		ret = NewFmlOr(ret, qeopt.qe(NewQuantifier(false, []Level{lv},
			NewFmlAnd(NewAtom(f, EQ), fof)), cond))
	}

	return ret
}

/*
 * fof: inequational constraints
 * atom: f <= 0 or f >= 0
 *
 * Returns: qff which is equivalent to ex([x], f <= 0 && fof_neq)
 :        : nil if fails
*/
func apply_neqQE_atom(fof Fof, atom *Atom, lv Level, qeopt QEopt, cond qeCond) Fof {
	if atom.op == EQ {
		return fof
	}
	if qeopt.assert && atom.op != GE && atom.op != LT {
		panic("unexpected op")
	}

	var ret Fof = falseObj
	poly := atom.getPoly()
	for {
		qffneq := apply_neqQE(fof, lv)

		deg := poly.Deg(lv)
		lc := poly.Coef(lv, uint(deg))
		if lc.IsNumeric() && (atom.op == LE && lc.Sign() < 0 ||
			atom.op == GE && lc.Sign() > 0) {
			return qffneq

		} else if deg%2 != 0 {
			ret = NewFmlOr(ret, NewFmlAnd(NewAtom(lc, NE), qffneq))
		} else if deg == 0 {
			return NewFmlOr(ret, NewFmlAnd(NewAtom(lc, atom.op), qffneq))
		} else if deg == 2 {
			discrim := poly.discrim2(lv)
			op := LT
			if atom.op == GE {
				op = GT
			}
			// ex([x], ax^2+bx+c >= 0)
			// <==>
			// infinite: a > 0 || b^2-4ac > 0 || (a=0 && b=0 && c >= 0)
			// __finite: b^2-4ac=0 /\ a !=0
			c1 := poly.Coef(lv, 1)
			c0 := poly.Coef(lv, 0)
			ret = NewFmlOr(ret, NewFmlAnd(NewAtom(lc, op), qffneq))
			ret = NewFmlOr(ret, NewFmlAnd(NewAtom(discrim, GT), qffneq))
			ret = NewFmlOr(ret, newFmlAnds(NewAtom(lc, EQ), NewAtom(c1, EQ), NewAtom(c0, atom.op), qffneq))
			ret = NewFmlOr(ret, newFmlAnds(
				NewAtom(lc, NE),
				NewAtom(discrim, EQ),
				qeopt.qe(NewQuantifier(false, []Level{lv}, NewFmlAnd(fof,
					NewAtom(NewInt(2).Mul(lc).Add(c1), EQ))), cond)))
			return ret

		} else if poly.isUnivariate() {
			r := apply_neqQE_atom_univ(fof, qffneq, NewAtom(poly, atom.op).(*Atom), lv, qeopt, cond)
			if r == nil {
				return nil
			}
			ret = NewFmlOr(ret, r)
			return ret
		} else {
			return nil
		}

		fof = NewFmlAnd(fof, NewAtom(lc, EQ))
		switch pp := poly.Sub(lc).(type) {
		case *Poly:
			poly = pp
		default:
			// 定数になった...
			ret = NewFmlOr(ret, NewFmlAnd(NewAtom(pp, atom.op), qffneq))
			return ret
		}
	}
}

func neqQE(fof Fof, lv Level, qeopt QEopt, cond qeCond) Fof {

	fne, fot := divide_neq(fof, lv, qeopt)
	if fot == trueObj {
		qeopt.log(cond, 3, "neq", "<%s> all %v\n", varstr(lv), fof)
		return apply_neqQE(fof, lv)
	}
	if fne == trueObj {
		return fof
	}
	if is_strict_only(fot, lv) {
		fstrict := NewQuantifier(false, []Level{lv}, fot)
		return NewFmlAnd(apply_neqQE(fne, lv), qeopt.qe(fstrict, cond))
	}
	if atom, ok := fot.(*Atom); ok {
		return apply_neqQE_atom(fne, atom, lv, qeopt, cond)
	}

	return fof
}

func (qeopt QEopt) qe_neq(fof FofQ, cond qeCond) Fof {

	var fml Fof
	not := false
	switch pp := fof.(type) {
	case *ForAll:
		fml = pp.fml.Not()
		not = true
	case *Exists:
		fml = pp.fml
	default:
		return fof
	}

	for _, q := range fof.Qs() {
		qeopt.log(cond, 2, "neq", "<%s> %v\n", varstr(q), fof)
		ff := neqQE(fml, q, qeopt, cond)
		if ff != fof {
			if not {
				return ff.Not()
			} else {
				return ff
			}
		}
	}
	return nil
}
