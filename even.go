package ganrac

// ex([x], p(x^2)) <==>
// ex([x], x >= 0 && p(x))

import (
	"fmt"
)

const (
	EVEN_NO   = 0x00
	EVEN_NG   = 0x01
	EVEN_LIN1 = 0x02 // 線形，ただし，主係数は定数
	EVEN_LIN2 = 0x04 // 線形で，主係数が変数
	EVEN_OK   = 0x08 // atom の因数分解した因子すべてが even 例：(x^2+1) * (x^4+3*x^2+1)
	EVEN_OKM  = 0x10 // atom.getPoly() が even 例：(x-1)*(x+1)
	EVEN_LIN  = EVEN_LIN1 | EVEN_LIN2
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
			fqs = append(fqs, fofq)
			for _, q := range fofq.Qs() {
				bs[q] = false
				v := fofq.isEven(q)
				if v&EVEN_NG != 0 {
					continue
				}
				if v&EVEN_OK != 0 {
					// 単純に次数を下げればいい．
					qeopt.log(cond, 1, "evenI", "<%s,%#x> %v\n", varstr(q), v, fofq)

					var ret Fof = falseObj
					qff := fofq.Fml()
					if fofq.isForAll() {
						qff = qff.Not()
					}

					for _, sgn := range []int{1, -1} {
						if sgn < 0 && v&(EVEN_LIN) == 0 {
							break
						}
						f := qff
						f = f.redEven(q, v, sgn)
						f = NewFmlAnd(f, NewAtom(NewPolyVar(q), GE))
						f = NewExists(fofq.Qs(), f)
						qeopt.log(cond, 1, "evenM", "<%s,%#x,%d> %v\n", varstr(q), v, sgn, f)

						varn := qeopt.varn
						cond2 := cond
						cond2.neccon = cond2.neccon.Subst(NewPolyVar(varn), q)
						cond2.sufcon = cond2.sufcon.Subst(NewPolyVar(varn), q)

						qeopt.varn++
						ret = NewFmlOr(ret, qeopt.qe(f, cond2))
						qeopt.varn--
					}
					if fofq.isForAll() {
						ret = ret.Not()
					}
					ret = qeopt.reconstruct(fqs, ret, cond)
					return qeopt.qe(ret, cond)
				}
			}
			fof = fofq.Fml()
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
				return EVEN_LIN1
			} else {
				return EVEN_LIN2
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
	if p.Deg(lv)%2 != 0 {
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

func (p *FofTFbase) redEven(lv Level, v, sgn int) Fof {
	return p
}

func (p *ForAll) redEven(lv Level, v, sgn int) Fof {
	return p.gen(p.Qs(), p.Fml().redEven(lv, v, sgn))
}

func (p *Exists) redEven(lv Level, v, sgn int) Fof {
	return p.gen(p.Qs(), p.Fml().redEven(lv, v, sgn))
}

func (p *FmlAnd) redEven(lv Level, v, sgn int) Fof {
	fmls := p.Fmls()
	fs := make([]Fof, len(fmls))
	for i, f := range fmls {
		fs[i] = f.redEven(lv, v, sgn)
	}
	return p.gen(fs)
}
func (p *FmlOr) redEven(lv Level, v, sgn int) Fof {
	fmls := p.Fmls()
	fs := make([]Fof, len(fmls))
	for i, f := range fmls {
		fs[i] = f.redEven(lv, v, sgn)
	}
	return p.gen(fs)
}

/*
 * p.Deg() == 1
// (a + b*sqrt(x)) / d == 0 <=> ab <= 0 && a^2 == b^2*x
// (a + b*sqrt(x)) / d <= 0 <=> ad <= 0 && a^2 >= b^2*x || bd <= 0 && a^2 <= b^2*x
// (a + b*sqrt(x)) / d <  0 <=> ad <  0 && a^2 >  b^2*x || bd <= 0 && (a*d < 0 || a^2 < b^2*x)
*/
func (q *Atom) redEvenLin(lv Level, v, sgn int) Fof {
	if q.Deg(lv) != 1 {
		panic(fmt.Sprintf("why? %v", q))
	}
	p := q.getPoly()

	b := p.Coef(lv, 1)
	if sgn < 0 {
		b = b.Neg()
	}

	a := p.Coef(lv, 0)
	x := NewPolyVar(lv)

	// a^2 - b^2*x
	abx := Sub(Mul(a, a), Mul(Mul(b, b), x))

	switch q.op {
	case EQ:
		return NewFmlAnd(NewAtom(Mul(a, b), LE), NewAtom(abx, EQ))
	case NE:
		return NewFmlOr(NewAtom(Mul(a, b), GT), NewAtom(abx, NE))
	case LE:
		return NewFmlOr(
			NewFmlAnd(
				NewAtom(a, LE),
				NewAtom(abx, GE)),
			NewFmlAnd(
				NewAtom(b, LE),
				NewAtom(abx, LE)))
	case GE:
		return NewFmlAnd(
			NewFmlOr(
				NewAtom(a, GE),
				NewAtom(abx, LE)),
			NewFmlOr(
				NewAtom(b, GE),
				NewAtom(abx, GE)))
	case LT:
		return newFmlOrs(
			NewFmlAnd(
				NewAtom(a, LT),
				NewAtom(abx, GT)),
			NewFmlAnd(
				NewAtom(b, LE),
				NewAtom(a, LT)),
			NewFmlAnd(
				NewAtom(b, LE),
				NewAtom(abx, LT)))
	case GT:
		return newFmlAnds(
			NewFmlOr(
				NewAtom(a, GT),
				NewAtom(abx, LT)),
			NewFmlOr(
				NewAtom(b, GE),
				NewAtom(a, GT)),
			NewFmlOr(
				NewAtom(b, GE),
				NewAtom(abx, GT)))
	default:
		panic(fmt.Sprintf("op=%d: %v", q.p, q))
	}
}

func (q *Atom) redEven(lv Level, v, sgn int) Fof {

	m := q.isEven(lv)
	pp := q.p
	if m == EVEN_OKM {
		pp = []*Poly{q.getPoly()}
	} else if m == EVEN_LIN1 || m == EVEN_LIN2 {
		return q.redEvenLin(lv, v, sgn)
	}

	ps := make([]RObj, len(pp))
	up := false
	for i, p := range pp {
		d := p.Deg(lv)
		if d > 0 {
			ps[i] = p.redEven(lv)
			up = true
		} else {
			ps[i] = p
		}
	}
	if !up {
		return q
	}
	return NewAtoms(ps, q.op)
}

/*
 * p is a even polynomial w.r.t. lv
 *
 * return q(x) s.t. q(x^2) = p(x)
 */
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
