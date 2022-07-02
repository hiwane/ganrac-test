package ganrac

// Quantifier elimination for inequational constraints.
// Hidenao IWANE. 2015 (Japanese)

// ex([x], f(x) != 0 && phi(x))

func neqQE(fof Fof, lv Level) Fof {
	var fml Fof
	switch pp := fof.(type) {
	case *ForAll:
		fml = pp.fml.Not()
	case *Exists:
		fml = pp.fml
	default:
		return fof
	}

	atom, ok := fml.(*Atom)
	if ok {
		if atom.op != NE {
			return fof
		}
		// ex([x], f(x) != 0)
		var ret Fof = falseObj
		for _, p := range atom.p {
			for d := p.Deg(lv); d >= 0; d-- {
				ret = NewFmlOr(ret, NewAtom(p.Coef(lv, uint(d)), NE))
			}
		}
		return ret
	}

	return fof
}

func (qeopt QEopt) qe_neq(fof FofQ, cond qeCond) Fof {
	for _, q := range fof.Qs() {
		qeopt.log(cond, 2, "neq", "<%s> %v\n", varstr(q), fof)
		ff := neqQE(fof, q)
		if ff != fof {
			return ff
		}
	}
	return nil
}
