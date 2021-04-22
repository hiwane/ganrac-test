package ganrac

import (
	"fmt"
)

type cadSqfr struct {
	p *Poly
	r int8
}

func newCadSqfr(cell *Cell, p *Poly, r int8) *cadSqfr {
	sq := new(cadSqfr)
	sq.p = p.primpart()
	if cell != nil {
		sq.p = cell.reduce(sq.p).(*Poly)
	}
	if sq.p.Sign() < 0 {
		sq.p = sq.p.Neg().(*Poly)
	}
	sq.r = r
	return sq
}

func (cell *Cell) isDE() bool {
	n := 0
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly != nil && len(c.defpoly.c) > 2 {
			n++
		}
	}
	return n > 1
}

func (cad *CAD) symsex_zero_chk(p *Poly, cell *Cell) bool {
	// Simple EXtension
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly == nil {
			continue
		}
		fmt.Printf("sym_zero_chk_node(): p=%v\n", p)
		if len(c.defpoly.c) == 2 { // 1次である. 代入で消去
			deg := p.Deg(c.lv)
			if deg == 0 {
				continue
			}
			coef1 := c.defpoly.c[1].Neg()
			dens := make([]RObj, deg+1)
			dens[0] = one
			dens[1] = coef1
			for i := 2; i <= deg; i++ {
				dens[i] = dens[i-1].Mul(coef1)
			}
			q := p.subst_frac(c.defpoly.c[0], dens, c.lv)
			fmt.Printf("subst q=%v\n", q)
			switch qq := q.(type) {
			case *Poly:
				p = qq
			case NObj:
				return qq.IsZero()
			default:
				if !qq.IsNumeric() {
					panic("??") // @DEBUG
				}
			}
		} else {
			if !p.isUnivariate() {
				panic("??") // @DEBUG
			}
			r := p.reduce(c.defpoly)
			fmt.Printf("sym_zero_chk_node(): q=%v\n", p)
			fmt.Printf("sym_zero_chk_node(): r=%v\n", r)
			return r.IsZero()
		}
	}
	if true {
		panic("to-ranai")
	}
	return p.IsZero()
}

func (cad *CAD) sym_equal(ci, cj *Cell) bool {
	// @TODO 同じ次数なら主係数有理数が起点のほうがいいか.
	if len(ci.defpoly.c) > len(cj.defpoly.c) {
		return cad.sym_zero_chk(ci.defpoly, cj)
	} else {
		return cad.sym_zero_chk(cj.defpoly, ci)
	}
}

func (cad *CAD) sym_zero_chk(p *Poly, c *Cell) bool {
	if !c.parent.isDE() {
		if len(c.defpoly.c) == 2 { // 1 次
			return cad.symsex_zero_chk(p, c)
		}
	}

	ret := cad.symde_zero_chk(p, c)
	fmt.Printf("sym_zero_chk() ret=%v\n", ret)
	return ret
}

// 定義多項式の主係数を整数にする
func (cad *CAD) symde_hcrat_cells(c *Cell) {
	if c.lv > 0 {
		cad.symde_hcrat_cells(c.parent)
	}
	if c.defpoly == nil {
		return
	}
	lc := c.defpoly.c[len(c.defpoly.c)-1]
	if lc.IsNumeric() {
		return
	}

	panic("sko!")
}
func (cad *CAD) symde_normalize(p *Poly, cell *Cell) RObj {
	// 主係数が非ゼロでないか確認し，次数を下げる
	for ; cell.lv > p.lv; cell = cell.parent {
	}

	for i := len(p.c) - 1; i >= 0; i-- {
		switch c := p.c[i].(type) {
		case *Poly:
			if !cad.symde_zero_chk(c, cell) {
				if i == 0 {
					return cad.symde_normalize(c, cell)
				}
				q := NewPoly(p.lv, i+1)
				copy(q.c, p.c)
				return q
			}
		case NObj:
			if i == 0 {
				return c
			}
			if !c.IsZero() {
				q := NewPoly(p.lv, i+1)
				copy(q.c, p.c)
				return q
			}
		default:
			panic("?")
		}
	}
	return zero
}

// p が 0 か判定する
func (cad *CAD) symde_zero_chk(porg *Poly, cell *Cell) bool {
	for cell.lv > porg.lv {
		cell = cell.parent
	}

	q, s2, _ := cad.symde_gcd(cell.defpoly, porg, cell, false)
	if q == nil {
		return false
	}

	if len(q.c) < len(cell.defpoly.c) {
		// 定義多項式が分解できた
		for prec := uint(50); ; prec += uint(50) {
			fmt.Printf("input p=%v\n", porg)
			cell.Print()
			fmt.Printf("s2=%v\n", s2)
			fmt.Printf("q=%v\n", q)
			x1 := cell.subst_intv(cad, q, prec).(*Interval)          // GCD
			x2 := cell.subst_intv(cad, s2.(*Poly), prec).(*Interval) // 外

			if x1.ContainsZero() && !x2.ContainsZero() {
				cell.defpoly = q
				return true
			} else if !x1.ContainsZero() && x2.ContainsZero() {
				cell.defpoly = s2.(*Poly)
				return false
			}
			panic("!")
		}
	}
	return true
}

func lv(p RObj) Level {
	switch pp := p.(type) {
	case *Poly:
		return pp.lv
	case NObj:
		return 9
	default:
		panic("unknown")
	}
}

func deg(p RObj) int {
	switch pp := p.(type) {
	case *Poly:
		return len(pp.c) - 1
	case NObj:
		if pp.IsZero() {
			return -1
		} else {
			return 0
		}
	default:
		panic("unknown")
	}
}

func (cad *CAD) symde_gcd(porg, qorg *Poly, cell *Cell, need_t bool) (*Poly, RObj, RObj) {
	// assume: porg.lv == qorg.lv
	// returns (gcd(p, q), p/gcd(p, q), q/gcd(p,q))
	// returns (nil, nil)

	if porg.lv != qorg.lv {
		panic(fmt.Sprintf("invalid p=[%d,%d], q=[%d,%d]", porg.lv, porg.deg(), qorg.lv, qorg.deg()))
	}

	var p, q *Poly
	var s1, s2 RObj
	var t1, t2 RObj

	if len(porg.c) < len(qorg.c) {
		s1 = one
		s2 = zero
		t1 = zero
		t2 = one
		p = qorg
		q = porg
	} else {
		s1 = zero
		s2 = one
		t1 = one
		t2 = zero
		p = porg
		q = qorg
	}

	// deg(p) >= deg(q)
	for {
		fmt.Printf("gcd : p=[%d,%d], q=[%d,%d]\n", p.lv, len(p.c)-1, q.lv, len(q.c)-1)
		a, b, rr := p.pquorem(q)
		s1, s2 = s2, Sub(Mul(s1, a), Mul(s2, b))
		fmt.Printf("gcd : s1=[%d,%d], s2=[%d,%d]\n", lv(s1), deg(s1), lv(s2), deg(s2))
		if need_t {
			t1, t2 = t2, Sub(Mul(t1, a), Mul(t2, b))
			fmt.Printf("gcd : t1=[%d,%d], t2=[%d,%d]\n", lv(t1), deg(t1), lv(t2), deg(t2))
		}

		if r, ok := rr.(*Poly); ok && (r.lv != q.lv || len(r.c) < len(q.c)) {
			rr = cad.symde_normalize(r, cell)
		}
		switch r := rr.(type) {
		case *Poly:
			//			fmt.Printf("rpol[%d,%d,]=%v\n", porg.lv, r.lv, r)
			if r.lv == q.lv {
				p, q = q, r
				continue
			} else {
				c := cell
				for r.lv != c.lv {
					c = c.parent
				}
				ret := cad.symde_zero_chk(r, c) //?
				if ret {
					return q, s2, t2
				}
				return nil, porg, qorg
			}
		case NObj:
			if r.IsZero() {
				return q, s2, t2
			} else {
				return nil, porg, qorg
			}
		default:
			panic("??")
		}
	}
}

func (cad *CAD) sym_sqfr(porg *Poly, cell *Cell) []*cadSqfr {
	p := porg
	if !p.isIntPoly() {
		panic("unexpected")
	}
	pd := porg.diff(porg.lv).(*Poly)

	s0, t0, _ := cad.symde_gcd(p, pd, cell, false)
	if s0 == nil {
		return []*cadSqfr{newCadSqfr(nil, porg, 1)}
	}

	// fmt.Printf("sqfr: in: lv=%d, deg=%d\n", porg.lv, len(porg.c)-1)
	// fmt.Printf("sqfr: s0: lv=%d, deg=%d\n", s0.lv, len(s0.c)-1)
	// fmt.Printf("sqfr: t%d: lv=%d, deg=%d\n", 0, t0.(*Poly).lv, len(t0.(*Poly).c)-1)

	ret := make([]*cadSqfr, 0)
	var i int8
	for i = 1; ; i++ {
		tt, ok := t0.(*Poly)
		if !ok || tt.lv != porg.lv {
			break
		}
		if err := s0.valid(); err != nil {
			panic(err)
		}
		if err := tt.valid(); err != nil {
			panic(err)
		}

		ti, si, fi := cad.symde_gcd(s0, tt, cell, true)
		if ti == nil {
			fmt.Printf("found [%d,%d^%d]\n", tt.lv, len(tt.c)-1, i)
			ret = append(ret, newCadSqfr(cell, tt, i))
			break
		} else {
			// fmt.Printf("sqfr: t%d: lv=%d, deg=%d %v\n", i, ti.lv, len(ti.c)-1, ti)
			// fmt.Printf("sqfr: f%d: %v\n", i, fi)

			// fmt.Printf("sqfr: f%d: lv=%d, deg=%d\n", i, fi.(*Poly).lv, len(fi.(*Poly).c)-1)
			// fmt.Printf("sqfr: s%d: lv=%d, deg=%d\n", i, si.(*Poly).lv, len(si.(*Poly).c)-1)

			if ti.lv != porg.lv {
				panic("stop")
			}

			if ff, ok := fi.(*Poly); ok && ff.lv == porg.lv {
				ret = append(ret, newCadSqfr(cell, ff, i))
			}
			t0 = cell.reduce(ti)
			switch ss := si.(type) {
			case *Poly:
				if s0, ok = cell.reduce(ss).(*Poly); !ok {
					goto _END
				}
			default:
				goto _END
			}
			if s0.lv != porg.lv {
				break
			}
		}
	}
_END:
	ret = append(ret, newCadSqfr(cell, t0.(*Poly), i+1))

	for _, r := range ret {
		if !r.p.isIntPoly() {
			panic("ge!")
		}
	}
	return ret
}
