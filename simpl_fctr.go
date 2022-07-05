package ganrac

/////////////////////////////////////
// 因数分解するよ
// @OX: oxサーバが必要
//
// simplification of quantifier-free formulas over ordered firlds
// A. Dolzmann, T. Sturm
/////////////////////////////////////

func (p *AtomT) simplFctr(g *Ganrac) Fof {
	return p
}

func (p *AtomF) simplFctr(g *Ganrac) Fof {
	return p
}

// @TODO p.factorized を使っていない
func (p *Atom) simplFctr(g *Ganrac) Fof {
	pp := [][]*Poly{make([]*Poly, 0), make([]*Poly, 0)}
	sgn := 1
	up := false
	for _, p := range p.p {
		fctr := g.ox.Factor(p)
		fctrn, _ := fctr.Geti(0)
		cont, _ := fctrn.(*List).Geti(0)
		sgn *= cont.(RObj).Sign()
		if !cont.(RObj).IsOne() && !cont.(RObj).IsMinusOne() {
			up = true
		}
		for i := fctr.Len() - 1; i > 0; i-- {
			fctrn, _ = fctr.Geti(i)
			ei, _ := fctrn.(*List).Geti(1)
			e := ei.(*Int).Int64()
			pi, _ := fctrn.(*List).Geti(0)
			pp[e%2] = append(pp[e%2], pi.(*Poly))
			if e > 1 {
				up = true
			}
		}
	}
	if !up && len(pp[0]) == 0 && len(pp[1]) == len(p.p) {
		return p
	}
	var ret Fof
	switch p.op {
	case EQ:
		ret = falseObj
		for _, q := range pp[0] {
			ret = NewFmlOr(ret, NewAtom(q, p.op))
		}
		for _, q := range pp[1] {
			ret = NewFmlOr(ret, NewAtom(q, p.op))
		}
		return ret
	case NE:
		ret = trueObj
		for _, q := range pp[0] {
			ret = NewFmlAnd(ret, NewAtom(q, p.op))
		}
		for _, q := range pp[1] {
			ret = NewFmlAnd(ret, NewAtom(q, p.op))
		}
		return ret
	}
	op := p.op
	if sgn < 0 {
		op = op.neg()
	}

	if (op & EQ) != 0 { // LE || GE
		if len(pp[1]) == 0 && op == GE {
			return trueObj
		}
		ret = falseObj
		for _, q := range pp[0] {
			ret = NewFmlOr(ret, NewAtom(q, EQ))
		}
		if len(pp[1]) > 0 {
			qq := make([]RObj, len(pp[1]))
			for i := 0; i < len(qq); i++ {
				qq[i] = pp[1][i]
			}
			ret = NewFmlOr(ret, NewAtoms(qq, op))
		}
	} else if len(pp[1]) == 0 && op == LT {
		return falseObj
	} else if len(pp[1]) == 0 && op == GE {
		return trueObj
	} else { // LT || GT
		ret = trueObj
		for _, q := range pp[0] {
			ret = NewFmlAnd(ret, NewAtom(q, NE))
		}
		qq := make([]RObj, len(pp[1]))
		for i := 0; i < len(qq); i++ {
			qq[i] = pp[1][i]
		}
		ret = NewFmlAnd(ret, NewAtoms(qq, op))
	}

	return ret
}

func (p *FmlAnd) simplFctr(g *Ganrac) Fof {
	fml := make([]Fof, len(p.fml))
	for i := 0; i < len(fml); i++ {
		fml[i] = p.fml[i].simplFctr(g)
	}

	return newFmlAnds(fml...)
}

func (p *FmlOr) simplFctr(g *Ganrac) Fof {
	fml := make([]Fof, len(p.fml))
	for i := 0; i < len(fml); i++ {
		fml[i] = p.fml[i].simplFctr(g)
	}

	return newFmlOrs(fml...)
}

func (p *ForAll) simplFctr(g *Ganrac) Fof {
	fml := p.fml.simplFctr(g)
	if fml == p.fml {
		return p
	}
	return NewQuantifier(true, p.q, fml)
}

func (p *Exists) simplFctr(g *Ganrac) Fof {
	fml := p.fml.simplFctr(g)
	if fml == p.fml {
		return p
	}
	return NewQuantifier(false, p.q, fml)
}
