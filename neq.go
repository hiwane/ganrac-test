package ganrac

// Quantifier elimination for inequational constraints.
// Hidenao IWANE. 2015 (Japanese)

// ex([x], f(x) != 0 && phi(x))

func is_all_neq(fof Fof, lv Level) bool {
	switch pp := fof.(type) {
	case FofQ:
		return is_all_neq(pp.Fml(), lv)
	case FofAO:
		for _, f := range pp.Fmls() {
			if !is_all_neq(f, lv) {
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

func neqQE(fof Fof, lv Level, qeopt QEopt, cond qeCond) Fof {

	if is_all_neq(fof, lv) {
		qeopt.log(cond, 3, "neq", "<%s> all %v\n", varstr(lv), fof)
		return apply_neqQE(fof, lv)
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
