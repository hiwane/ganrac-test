package ganrac

// ex([x], p(x^2)) <==>
// ex([x], x >= 0 && p(x))

import (
	"fmt"
)

const (
	EVEN_NO  = 0
	EVEN_NG  = 1
	EVEN_LIN = 2
	EVEN_OK  = 4
	EVEN_OKM = 8 // 積の計算が必要
)

// FofAObase FofQbase
func (qeopt *QEopt) qe_evenq(prenex_fof Fof, cond qeCond) Fof {
	fof := prenex_fof

	bs := make([]bool, qeopt.varn)
	fof.Indets(bs)
	fqs := make([]FofQ, 0)

	for {
		// quantified var.
		if fofq, ok := fof.(FofQ); ok {
			for _, q := range fofq.Qs() {
				bs[q] = false
				if v := fofq.isEven(q); v&EVEN_NG == 0 {
					if v == EVEN_OK {
						// 単純に次数を下げればいい．
						f := fofq.Fml()
						f = f.redEven(q, v)
						if fofq.isForAll() {
							f = NewFmlOr(f, NewAtom(NewPolyVar(q), LT))
						} else {
							f = NewFmlAnd(f, NewAtom(NewPolyVar(q), GE))
						}
						f = fofq.gen(fofq.Qs(), f)
						qeopt.log(cond, 1, "qeven", "<%s,%d> %v\n", varstr(q), v, f)

						cond2 := cond
						cond2.neccon = cond2.neccon.Subst(NewPolyVar(qeopt.varn), q)
						cond2.sufcon = cond2.sufcon.Subst(NewPolyVar(qeopt.varn), q)
						qeopt.varn++
						cond2.depth++
						f = qeopt.qe(f, cond2)
						qeopt.varn--

						// 再構築
						if len(fqs) > 0 {
							for i := len(fqs) - 1; i >= 0; i-- {
								f = fqs[i].gen(fqs[i].Qs(), f)
							}
							return qeopt.qe(f, cond)
						}
						return f
					}
				}
			}
			fof = fofq.Fml()
			fqs = append(fqs, fofq)
		} else {
			break
		}
	}

	fof = prenex_fof

	// free var.  次数も
	for j, b := range bs {
		if b {
			if v := fof.isEven(Level(j)); v&EVEN_NG == 0 {
			}
		}
	}

	return nil
}

// 部品ごと. これだと (x+a)(x-a) が even 扱いでなくなる...
func (p *Atom) isEvenE(lv Level) int {
	for _, pp := range p.p {
		if !pp.hasVar(lv) {
			continue
		}
		d := pp.Deg(lv)
		if d%2 == 0 {
			for i := 1; i <= d; i += 2 {
				if !pp.Coef(lv, uint(i)).IsZero() {
					return EVEN_NG
				}
			}
			continue
		}
		if d == 1 && len(p.p) == 1 {
			c := pp.Coef(lv, 1)
			if _, ok := c.(NObj); ok {
				return EVEN_LIN
			}
		}
		return EVEN_NG
	}
	return EVEN_OK
}

func (p *Atom) isEven(lv Level) int {
	if v := p.isEvenE(lv); v != EVEN_NG {
		return v
	}
	if len(p.p) == 1 {
		return EVEN_NG
	}
	d := 0
	for _, q := range p.p {
		d += q.deg()
	}
	if d%2 != 0 {
		return EVEN_NG
	}

	m := p.p[0]
	for i := 1; i < len(p.p); i++ {
		m = m.Mul(p.p[i]).(*Poly)
	}
	for i := 1; i < len(m.c); i += 2 {
		c := m.Coef(lv, uint(i))
		if !c.IsZero() {
			return EVEN_NG
		}
	}
	return EVEN_OKM
}

func (p *FofTFbase) isEven(lv Level) int {
	return EVEN_NO
}

func (p *FmlAnd) isEven(lv Level) int {
	v := 0
	for _, f := range p.Fmls() {
		v |= f.isEven(lv)
		if v&EVEN_NG != 0 {
			return v
		}
	}
	return v
}
func (p *FmlOr) isEven(lv Level) int {
	v := 0
	for _, f := range p.Fmls() {
		v |= f.isEven(lv)
		if v&EVEN_NG != 0 {
			return v
		}
	}
	return v
}

func (p *ForAll) isEven(lv Level) int {
	return p.Fml().isEven(lv)
}
func (p *Exists) isEven(lv Level) int {
	return p.Fml().isEven(lv)
}

func (p *FofTFbase) redEven(lv Level, v int) Fof {
	return p
}

func (p *ForAll) redEven(lv Level, v int) Fof {
	return p.gen(p.Qs(), p.Fml().redEven(lv, v))
}

func (p *Exists) redEven(lv Level, v int) Fof {
	return p.gen(p.Qs(), p.Fml().redEven(lv, v))
}

func (p *FmlAnd) redEven(lv Level, v int) Fof {
	fmls := p.Fmls()
	fs := make([]Fof, len(fmls))
	for i, f := range fmls {
		fs[i] = f.redEven(lv, v)
	}
	return p.gen(fs)
}
func (p *FmlOr) redEven(lv Level, v int) Fof {
	fmls := p.Fmls()
	fs := make([]Fof, len(fmls))
	for i, f := range fmls {
		fs[i] = f.redEven(lv, v)
	}
	return p.gen(fs)
}

func (p *Atom) redEven(lv Level, v int) Fof {
	ps := make([]RObj, len(p.p))
	up := false
	for i, p := range p.p {
		d := p.Deg(lv)
		if d > 0 {
			ps[i] = p.redEven(lv)
			up = true
		} else {
			ps[i] = p
		}
	}
	if !up {
		return p
	}
	return NewAtoms(ps, p.op)
}

func (p *Poly) redEven(lv Level) *Poly {
	if p.lv < lv {
		return p
	} else if p.lv > lv {
		q := NewPoly(p.lv, len(p.c))
		for i, cc := range p.c {
			switch c := cc.(type) {
			case *Poly:
				q.c[i] = c.redEven(lv)
			default:
				q.c[i] = c
			}
		}
		return q
	}

	q := NewPoly(p.lv, p.deg()/2+1)
	for i := 0; i < len(q.c); i++ {
		q.c[i] = p.c[2*i]
	}
	if err := q.valid(); err != nil {
		panic(fmt.Sprintf("err=%v: %v", err, q))
	}
	return q
}
