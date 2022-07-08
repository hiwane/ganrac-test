package ganrac

import (
	"fmt"
	// "sort"
)

/////////////////////////////////////
// 論理式の簡単化.
// 等式制約を利用して，atom を簡単化する
/////////////////////////////////////

func newReduceInfo() *reduce_info {
	inf := new(reduce_info)
	inf.eqns = NewList()
	return inf
}

func (src *reduce_info) Clone() *reduce_info {
	dest := new(reduce_info)
	*dest = *src
	dest.q = make([]Level, len(src.q))
	copy(dest.q, src.q)
	dest.varb = make([]bool, len(src.varb))
	copy(dest.varb, src.varb)
	dest.depth = src.depth + 1
	dest.eqns = NewList()
	for _, p := range src.eqns.Iter() {
		dest.eqns.Append(p)
	}
	return dest
}

func (inf *reduce_info) isQ(lv Level) bool {
	for _, q := range inf.q {
		if q == lv {
			return true
		}
	}
	return false
}

func (inf *reduce_info) Reduce(g *Ganrac, p *Poly) (RObj, bool) {
	n := 0 // 一時的に変数リストを増やす
	if int(p.lv) >= len(inf.varb) {
		b := make([]bool, p.lv+1)
		copy(b, inf.varb)
		inf.varb = b
	}

	for lv := p.lv; lv >= Level(0); lv-- {
		if !inf.varb[lv] && p.hasVar(lv) {
			inf.vars.Append(NewPolyVar(lv))
			n++
		}
	}

	r, neg := g.ox.Reduce(p, inf.eqns, inf.vars, inf.qn+n)
	inf.vars.v = inf.vars.v[:inf.vars.Len()-n] // 元に戻す

	return r, neg
}

func (inf *reduce_info) GB(g *Ganrac, lvmax Level) *List {
	quan := make([]Level, 0, lvmax)
	free := make([]Level, 0, lvmax)
	varb := make([]bool, lvmax+1)
	for lv := Level(0); lv <= lvmax; lv++ {
		for i := inf.eqns.Len() - 1; i >= 0; i-- {
			e := inf.eqns.getiPoly(i)
			if e.hasVar(lv) {
				if inf.isQ(lv) {
					quan = append(quan, lv)
				} else {
					free = append(free, lv)
				}
				varb[lv] = true
				break
			}
		}
	}

	vars := NewList()
	for _, lv := range free {
		vars.Append(NewPolyVar(lv))
	}
	for _, lv := range quan {
		vars.Append(NewPolyVar(lv))
	}
	inf.vars = vars
	inf.varb = varb
	inf.qn = len(quan)

	g.log(8, "GBi=%v\n", inf.eqns)
	gb := g.ox.GB(inf.eqns, vars, inf.qn)
	g.log(8, "GBo=%v\n", gb)
	return gb
}

func (p *AtomT) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	return p
}

func (p *AtomF) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	return p
}

func (p *Atom) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	if inf.eqns.Len() == 0 {
		return p
	}
	q := p.getPoly()
	r, neg := inf.Reduce(g, q)
	if !q.Equals(r) {
		g.log(3, "simplReduce(Atom) %v => %v\n", q, r)
		var a Fof
		if neg {
			a = NewAtom(r, p.op.neg())
		} else {
			a = NewAtom(r, p.op)
		}
		return a.simplFctr(g)
	}
	return p
}

func simplReduceAO(g *Ganrac, inf *reduce_info, p FofAO, op OP) Fof {
	update := false
	fmls := make([]Fof, len(p.Fmls()))
	for i, fml := range p.Fmls() {
		fmls[i] = fml.simplReduce(g, inf)
		if fmls[i] != fml {
			update = true
		}
	}

	n := 0
	b := make([]int, len(fmls))
	for i, fml := range fmls {
		switch f := fml.(type) {
		case *Atom:
			if f.op != op {
				continue
			}
			b[i] = +1
			n++
		case FofAO:
			noeq := false
			for _, g := range f.Fmls() {
				if a, ok := g.(*Atom); !ok || a.op != op {
					noeq = true
					break
				}
			}
			if noeq {
				continue
			}
			// 積が等式制約
			b[i] = -1
			n++
		}
	}

	// eqcon の更新
	if n > 0 {
		// 今の等式制約で簡約... 定数になったら?

		// 新たに等式制約を追加して GB 計算
		inf = inf.Clone()
		for i, fml := range fmls {
			if b[i] > 0 {
				inf.eqns.Append(fml.(*Atom).getPoly())
			} else if b[i] < 0 {
				var mul RObj
				for j, f := range fml.(FofAO).Fmls() {
					if j == 0 {
						mul = f.(*Atom).getPoly()
					} else {
						mul = Mul(mul, f.(*Atom).getPoly())
					}
				}

				inf.eqns.Append(mul.(*Poly))
			}
		}

		maxvar := p.maxVar()
		for _, eq := range inf.eqns.Iter() {
			mv := eq.(*Poly).maxVar()
			if mv > maxvar {
				maxvar = mv
			}
		}

		inf.eqns = inf.GB(g, maxvar)
		if v, ok := inf.eqns.geti(0).(NObj); ok {
			if v.Sign() == 0 {
				panic("????")
			}
			return NewAtom(v, op)
		}

		for i, fml := range fmls {
			if b[i] != 0 {
				continue
			}
			fmls[i] = fml.simplReduce(g, inf)
			if fmls[i] != fml {
				update = true
			}
		}
	}
	if update {
		return p.gen(fmls)
	} else {
		return p
	}
}

func (p *FmlAnd) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	return simplReduceAO(g, inf, p, EQ)
}

func (p *FmlOr) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	return simplReduceAO(g, inf, p, NE)
}

func simplReduceQ(g *Ganrac, inf *reduce_info, p FofQ) Fof {
	// inf に p.Qs() な変数が含まれていたら，
	// それは別の変数扱いなので，除去が必要
	qs := make([]Level, 0, len(p.Qs()))
	for _, q := range p.Qs() {
		if int(q) < len(inf.varb) && inf.varb[q] {
			qs = append(qs, q)
		}
	}

	if len(qs) > 0 {
		// qs に含まれる変数を, 等式制約から消去する
		infb := inf
		inf = infb.Clone()
		inf.vars = NewList()
		for _, q := range qs {
			inf.vars.Append(NewPolyVar(q))
		}
		for _, v := range infb.vars.Iter() {
			flg := true
			for _, q := range qs {
				if q == v.(*Poly).lv {
					flg = false
					break
				}
			}
			if flg {
				inf.vars.Append(v)
			}
		}
		if infb.vars.Len() != inf.vars.Len() {
			fmt.Printf("old=%v\n", infb.vars)
			fmt.Printf("new=%v\n", inf.vars)
			fmt.Printf("qs =%v\n", qs)
			panic("?")
		}
		inf.eqns = g.ox.GB(inf.eqns, inf.vars, inf.vars.Len()-len(qs))

		gb := NewList()
		for _, p := range inf.eqns.Iter() {
			b := true
			for _, q := range qs {
				if p.(*Poly).hasVar(q) {
					b = false
					break
				}
			}
			if b {
				gb.Append(p)
			}
		}

		// inf.varb, inf.qn を更新
		inf.eqns = gb
		inf.varb = make([]bool, len(inf.varb))
		for lv := 0; lv < len(inf.varb); lv++ {
			for _, _p := range gb.Iter() {
				p := _p.(*Poly)
				if p.hasVar(Level(lv)) {
					inf.varb[lv] = true
					break
				}
			}
		}
	}

	fml := p.Fml().simplReduce(g, inf)
	if fml == p.Fml() {
		return p
	}
	return p.gen(p.Qs(), fml)
}

func (p *ForAll) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	return simplReduceQ(g, inf, p)
}

func (p *Exists) simplReduce(g *Ganrac, inf *reduce_info) Fof {
	return simplReduceQ(g, inf, p)
}
