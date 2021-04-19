package ganrac

import (
	"fmt"
	"os"
)

type cadSqfr struct {
	p *Poly
	r int
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
			deg := p.Deg(Level(c.lv))
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
			q := p.subst_frac(c.defpoly.c[0], dens, Level(c.lv))
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
	for ; Level(cell.lv) > p.lv; cell = cell.parent {
	}

	for i := len(p.c) - 1; i >= 0; i++ {
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
	// cad.symde_hcrat_cells(cell)

	// GCD.
	p := porg
	q := cell.defpoly

	var s1, s2 RObj
	if len(p.c) < len(cell.defpoly.c) {
		r := p
		p = cell.defpoly
		q = r
		s1 = zero
		s2 = one
	} else {
		s1 = one
		s2 = zero
	}

	// deg(p) >= deg(q)
	for {
		a, b, rr := p.pquorem(q)
		s1, s2 = s2, Sub(Mul(s1, a), Mul(s2, b))

		if r, ok := rr.(*Poly); ok && (r.lv != q.lv || len(r.c) < len(q.c)) {
			rr = cad.symde_normalize(r, cell)
		}
		switch r := rr.(type) {
		case *Poly:
			fmt.Printf("rpol[%d,%d,]=%v\n", porg.lv, r.lv, r)
			if r.lv == q.lv {
				p, q = q, r
				continue
			} else {
				c := cell
				for r.lv != Level(c.lv) {
					c = c.parent
				}
				ret := cad.symde_zero_chk(r, c) //?
				if ret {
					goto _GCD
				}
				return ret
			}
		case NObj:
			fmt.Printf("rnum[%d]=%v\n", porg.lv, r)
			if r.IsZero() {
				goto _GCD
			} else {
				return false
			}
		default:
			panic("??")
		}
	}
_GCD: // 共通因子がみつかった
	if len(q.c) < len(cell.defpoly.c) {
		// 定義多項式が分解できた
		for prec := uint(50); ; prec += uint(50) {
			fmt.Printf("input p=%v\n", porg)
			cell.Print(os.Stdout)
			fmt.Printf("s1=%v\n", s1)
			fmt.Printf("s2=%v\n", s2)
			fmt.Printf("p=%v\n", p)
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

func (cad *CAD) sym_sqfr(porg *Poly, cell *Cell) []cadSqfr {
	ret := make([]cadSqfr, 0)

	// p := porg
	// pd := porg.diff(porg.lv)
	//
	// gcd := cad.sym_gcd(p, pd, cell)
	// flat := cad.sym_div(p, gcd, cell)

	return ret
}
