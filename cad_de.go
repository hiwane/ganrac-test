package ganrac

// H. Iwane, H. Yanami, H. Anai, K. Yokoyama
// An effective implementation of symbolic–numeric cylindrical algebraic decomposition for quantifier elimination
// symoblic numeric computation 2009.

// Wang's rational reconstruction algo.
// Monagan's maximal quotient rational reconstruction algo.

import (
	"fmt"
)

type cadSqfr struct {
	p *Poly
	r mult_t
}

func newCadSqfr(cell *Cell, p *Poly, r mult_t) *cadSqfr {
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

func (cell *Cellmod) mod_monic(cad *CAD, p Uint) bool {
	// 定義多項式の主係数を 1 にする
	if cell.parent != nil && cell.parent.lv >= 0 {
		if !cell.parent.mod_monic(cad, p) {
			return false
		}
	}
	if cell.defpoly == nil {
		return true
	}

	lc := cell.defpoly.lc().(Moder)
	if !lc.IsOne() {
		lcinv := lc.inv_mod(cell.parent, p)
		if lcinv == nil {
			// 定義多項式の主係数非ゼロは保証しているが...?
			return false
		}
		cell.defpoly = cell.defpoly.mul_mod(lcinv, p).(*Poly)
	}
	return true
}

func (cell *Cell) mod(cad *CAD, p Uint) (*Cellmod, bool) {
	// mod する. 定義多項式があるところにしか興味がない
	var cellp, cp, cold *Cellmod
	if cad.rootp == nil {
		panic("not initialized rootp")
	}
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly == nil {
			continue
		}

		cp = NewCellmod(c)
		cp.parent = cad.rootp
		if cellp == nil {
			cellp = cp
		}
		switch q := c.defpoly.mod(p).(type) {
		case *Poly:
			if q.lv != c.defpoly.lv || q.deg() != c.defpoly.deg() {
				return nil, false
			}
			cp.defpoly = q
		default:
			return nil, false
		}

		if cold != nil {
			cold.parent = cp
		}
		cold = cp
	}
	if cold == nil {
		return cad.rootp, true
	}
	cold.parent = cad.rootp
	cold.de = false

	if !cellp.mod_monic(cad, p) {
		// 定義多項式が因数分解された.
		return cellp, false
	}
	return cellp, true
}

func (cad *CAD) symsex_zero_chk(p *Poly, cell *Cell) bool {
	// Simple Extension
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly == nil {
			continue
		}
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
			return r.IsZero()
		}
	}
	if true {
		panic("to-ranai")
	}
	return p.IsZero()
}

func (cad *CAD) sym_equal(ci, cj *Cell) bool {
	fmt.Printf("    sym_equal(%v,%v) deg=(%d,%d)\n", ci.Index(), cj.Index(), ci.defpoly.deg(), cj.defpoly.deg())

	if len(ci.defpoly.c) > len(cj.defpoly.c) {
		return cad.sym_zero_chk(ci.defpoly, cj)
	} else {
		return cad.sym_zero_chk(cj.defpoly, ci)
	}
}

func (cad *CAD) sym_zero_chk(p *Poly, c *Cell) bool {
	if !c.parent.isDE() {
		if c.defpoly.deg() == 1 {
			return cad.symsex_zero_chk(p, c)
		}
	}

	ret := cad.symde_zero_chk2(p, c, 0)
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
			if !cad.symde_zero_chk2(c, cell, 0) {
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

func (cad *CAD) symde_zero_chk2(porg *Poly, cell *Cell, pi int) bool {
	for cell.lv > porg.lv {
		cell = cell.parent
	}

	q, s2 := cad.symde_gcd2(cell.defpoly, porg, cell.parent, pi)
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
			x1 := cell.subst_intv(cad, q, prec).(*Interval)  // GCD
			x2 := cell.subst_intv(cad, s2, prec).(*Interval) // 外

			if x1.ContainsZero() && !x2.ContainsZero() {
				cell.defpoly = q
				return true
			} else if !x1.ContainsZero() && x2.ContainsZero() {
				cell.defpoly = s2
				return false
			}
			panic("!")
		}
	}
	return true
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
			fmt.Printf("   defp=%v\n", cell.parent.defpoly)
			fmt.Printf("   defp=%v\n", cell.defpoly)
			fmt.Printf("     s2=%v\n", s2)
			fmt.Printf("      q=%v\n", q)

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
		if pp == nil {
			return -2
		}
		return pp.lv
	case NObj:
		return -1
	default:
		if pp == nil {
			return -2
		}
		panic("unknown")
	}
}

func deg(p RObj) int {
	switch pp := p.(type) {
	case *Poly:
		if pp == nil {
			return -2
		}
		return len(pp.c) - 1
	case NObj:
		if pp.IsZero() {
			return -1
		} else {
			return 0
		}
	default:
		if pp == nil {
			return -2
		}
		panic("unknown")
	}
}

func (cad *CAD) symde_zero_chk_mod(forg *Poly, cell *Cellmod, p Uint) (bool, bool) {
	// returns (a, b)
	//     ... a := (forg == 0)
	//     ... b := (DE)
	for cell.lv > forg.lv {
		cell = cell.parent
	}
	q, s2, _ := cad.symde_gcd_mod(cell.defpoly, forg, cell.parent, p, false)
	if s2 == nil {
		// 定義多項式が分解されてしまった
		return false, false
	}
	if q == nil {
		return false, true
	}
	if q.deg() < cell.defpoly.deg() {
		// 定義多項式が分解できた.
		ff, _ := q.monicize(cell, p)
		cell.factor1 = ff.(*Poly)
		ff, _ = cell.defpoly.divmod_poly_mod(cell.factor1, cell.parent, p)
		cell.factor2 = ff.(*Poly)
		return false, false
	}

	return true, true
}

func (cad *CAD) symde_gcd_mod(forg, gorg *Poly, cell *Cellmod, p Uint, need_t bool) (*Poly, Moder, Moder) {
	// returns (g, a, b) where g = gcd(forg, gorg), g = a * forg + b * gorg
	// returns (nil, nil, nil) .... DE
	// returns (nil, a, b) .... gcd(forg, gorg) = 1
	if forg.lv != gorg.lv {
		panic(fmt.Sprintf("invalid p=[%d,%d], q=[%d,%d]", forg.lv, forg.deg(), gorg.lv, gorg.deg()))
	}

	var f, g *Poly
	var s1, s2 Moder
	var t1, t2 Moder
	var s_1 Uint

	if len(forg.c) < len(gorg.c) {
		f = gorg
		g = forg
	} else {
		s_1 = 1
		f = forg
		g = gorg
	}

	s1 = s_1
	s2 = Uint(1) - s_1
	t1 = s2
	t2 = s1

	for {
		if err := g.valid(); err != nil {
			panic(fmt.Sprintf("invalid g %v: %v,  <%v,%v>\n", g, err, forg, gorg))
		}
		q, rr := f.divmod_poly_mod(g, cell, p)
		if q == nil {
			return nil, nil, nil
		}
		if err := rr.valid_mod(cell, p); err != nil {
			panic(fmt.Sprintf("invalid r %v: %v,  <%v,%v>\n", rr, err, forg, gorg))
		}
		if q.IsZero() && rr.IsZero() {
			// f は zero だった...
			return g, s2, t2
		}
		if r, ok := rr.(*Poly); ok && (r.lv != g.lv || len(r.c) < len(g.c)) {
			rr = r.simpl_mod(cell, p)
		}
		if err := rr.valid_mod(cell, p); err != nil {
			panic(fmt.Sprintf("invalid %v: %v,  <%v,%v>\n", rr, err, forg, gorg))
		}
		if rr.IsZero() {
			return g, s2, t2
		}

		s1, s2 = s2, s1.sub_mod(s2.mul_mod(q, p), p)
		if need_t {
			t1, t2 = t2, t1.sub_mod(t2.mul_mod(q, p), p)
		}

		switch r := rr.(type) {
		case *Poly:
			if r.lv == g.lv {
				f, g = g, r
				continue
			} else {
				c := cell
				for r.lv != c.lv {
					c = c.parent
				}
				ret, ok := cad.symde_zero_chk_mod(r, c, p)
				if !ok {
					return nil, nil, nil
				}

				if ret {
					return g, s2, t2
				} else {
					return nil, forg, gorg
				}
			}
		case Uint:
			if r == 0 {
				return g, s2, t2
			} else {
				return nil, forg, gorg
			}
		}
	}
}

func (cad *CAD) test_div(h, f *Poly, cell *Cell, pi int) (bool, *Poly) {
	// 試し割り.
	// assume: f.lc() in Z
	// return h % f == 0, h / f
	_, qq, rr := h.pquorem(f)
	q := qq.(*Poly)

	switch r := rr.(type) {
	case *Int:
		return r.IsZero(), q
	case *Poly:
		if r.lv != h.lv {
			return cad.symde_zero_chk2(r, cell, pi+1), q
		}
		for i := range r.c {
			switch c := r.c[i].(type) {
			case *Poly:
				if !cad.symde_zero_chk2(c, cell, pi+1) {
					return false, q
				}
			case *Int:
				if !c.IsZero() {
					return false, q
				}
			}
		}
		return true, q
	}
	panic("?")
}

func (cad *CAD) symde_gcd2(forg, gorg *Poly, cell *Cell, pi int) (*Poly, *Poly) {
	// assume: forg.lv == gorg.lv
	// returns (gcd(f,g), f/gcd(f,g) or nil)
	// returns (nil, forg) if gcd=1

	if forg.lv != gorg.lv {
		panic(fmt.Sprintf("invalid p=[%d,%d], q=[%d,%d]", forg.lv, forg.deg(), gorg.lv, gorg.deg()))
	}

	var f, g *Poly

	f = forg
	g = gorg

	var g_crt *Poly  // GCD
	var f1_crt *Poly // factor1
	var f2_crt *Poly // factor2

	var g_rr *Poly  // GCD
	var f1_rr *Poly // factor1
	var f2_rr *Poly // factor2

	var pm *Int
	pos := 0
	tried := false

	count := 0
	for pidx, p := range lprime_table[pi:] {
		fp, ok := f.mod(p).(*Poly)
		if !ok || fp.lv != f.lv || fp.deg() != f.deg() { // unlucky
			continue
		}
		gp, ok := g.mod(p).(*Poly)
		if !ok || gp.lv != g.lv || gp.deg() != g.deg() { // unlucky
			continue
		}
		cellp, ok := cell.mod(cad, p)
		if !ok {
			if pos == 2 { // 他の素数では，この段階で共通因子なかった
				continue
			}

			unlucky := false
			c := cellp
			for ; c != nil; c = c.parent {
				if c.factor1 != nil {
					if c.parent == nil {
						unlucky = true // 既約なはず.
					}
					break
				}
			}
			if unlucky {
				continue
			}

			pos = 1
			if f1_crt == nil || c.lv != f1_crt.lv && count >= 5 ||
				c.lv == f1_crt.lv && f1_crt.deg() > c.factor1.deg() {
				f1_crt = c.factor1
				f2_crt = c.factor2
				f1_rr = nil
				f2_rr = nil
				pm = NewInt(int64(p))
				count = 0
				tried = false
				continue
			}
			if c.lv != f1_crt.lv {
				// 前の素数でみつかったものが偽因子かも保険.
				count++
				continue
			}
			if f1_crt.deg() < c.factor1.deg() { // 偽因子
				continue
			}

			// CRT する.
			no_chg := true
			f1_crt, f1_rr, _, ok = f1_crt.crt_interpol(f1_rr, c.factor1, pm, p)
			if !ok {
				no_chg = false
			}
			f2_crt, f2_rr, pm, ok = f2_crt.crt_interpol(f2_rr, c.factor2, pm, p)
			if !ok {
				no_chg = false
			}

			if no_chg && !tried {
				// 試し割りして，うまくいったら，再帰する
				tried = true
			} else {
				tried = false
			}
			continue
		}
		if cellp == nil { // unlucky
			continue
		}

		pos = 2
		gcd, s, _ := cad.symde_gcd_mod(fp, gp, cellp, p, false)
		if s == nil {
			// 定義多項式が因数分解された.
			// @TODO 共通化したいが...
		}
		if gcd == nil {
			// 共通因子がなかった.
			return nil, forg
		}

		if g_crt == nil || g_crt.deg() > gcd.deg() {
			// 1回目
			f1_crt = nil
			f2_crt = nil
			f1_rr = nil
			f2_rr = nil

			ggg, _ := gcd.monicize(cellp, p)
			g_crt = ggg.(*Poly)
			g_rr = nil
			pm = NewInt(int64(p))
			tried = false
			continue
		}

		// 2 回目以降
		if g_crt.deg() < gcd.deg() { // 偽因子
			continue
		}

		no_chg := true
		ggg, _ := gcd.monicize(cellp, p)
		g_crt, g_rr, pm, ok = g_crt.crt_interpol(g_rr, ggg.(*Poly), pm, p)
		if !ok {
			no_chg = false
		}

		if no_chg && !tried {
			// 試し割り
			if g_rr.deg() == forg.deg() {
				return forg, nil
			}

			if ok, q := cad.test_div(forg, g_rr, cell, pi+pidx); ok {
				return g_rr, q
			}
			fmt.Printf("try failed\n")
			panic("!")
			tried = true
		} else {
			tried = false
		}
		continue
	}

	panic(fmt.Sprintf("no more prime number [%d:%d]", pi, len(lprime_table)))
}

func (cad *CAD) symde_gcd(forg, gorg *Poly, cell *Cell, need_t bool) (*Poly, RObj, RObj) {
	// assume: forg.lv == gorg.lv
	// returns (gcd(f,g), f/gcd(f,g), g/gcd(f,g))
	// returns (nil, forg, gord) if gcd=1

	if forg.lv != gorg.lv {
		panic(fmt.Sprintf("invalid p=[%d,%d], q=[%d,%d]", forg.lv, forg.deg(), gorg.lv, gorg.deg()))
	}

	var f, g *Poly
	var s1, s2 RObj
	var t1, t2 RObj

	if len(forg.c) < len(gorg.c) {
		f = gorg
		g = forg
		s1 = one
		s2 = zero
	} else {
		f = forg
		g = gorg
		s1 = zero
		s2 = one
	}
	t1 = s2
	t2 = s1

	// CRT....
	// for _, p := range lprime_table {
	// 	fp, ok := f.mod(p).(*Poly)
	// 	if !ok || fp.lv != f.lv || fp.deg() != f.deg() {
	// 		continue
	// 	}
	// 	gp, ok := g.mod(p).(*Poly)
	// 	if !ok || gp.lv != g.lv || gp.deg() != g.deg() {
	// 		continue
	// 	}
	// 	cellp, ok := cell.mod(cad, p)
	// 	if cellp == nil {
	// 		continue
	// 	}
	//
	// 	h, s, t := cad.symde_gcd_mod(fp, gp, cellp, p, s1, need_t)
	//
	//
	// }

	// deg(f) >= deg(g)
	for {
		fmt.Printf("gcd : f =[%d,%3d], g =[%d,%3d] <%3d,%3d>\n", f.lv, f.deg(), g.lv, g.deg(), forg.deg(), gorg.deg())
		a, b, rr := f.pquorem(g)
		s1, s2 = s2, Sub(Mul(s1, a), Mul(s2, b))
		fmt.Printf("gcd : s1=[%d,%3d], s2=[%d,%3d]\n", lv(s1), deg(s1), lv(s2), deg(s2))
		if need_t {
			t1, t2 = t2, Sub(Mul(t1, a), Mul(t2, b))
			fmt.Printf("gcd : t1=[%d,%3d], t2=[%d,%3d]\n", lv(t1), deg(t1), lv(t2), deg(t2))
		}

		if r, ok := rr.(*Poly); ok && (r.lv != g.lv || len(r.c) < len(g.c)) {
			rr = cad.symde_normalize(r, cell)
		}
		switch r := rr.(type) {
		case *Poly:
			//			fmt.Printf("rpol[%d,%d,]=%v\n", porg.lv, r.lv, r)
			if r.lv == g.lv {
				f, g = g, r
				continue
			} else {
				c := cell
				for r.lv != c.lv {
					c = c.parent
				}
				ret := cad.symde_zero_chk(r, c) //?
				if ret {
					fmt.Printf("gcd : f =[%d,%3d], g =[%d,%3d] <%3d,%3d> END1\n", f.lv, f.deg(), g.lv, g.deg(), forg.deg(), gorg.deg())
					return g, s2, t2
				}
				fmt.Printf("gcd : f =[%d,%3d], g =[%d,%3d] <%3d,%3d> END2\n", f.lv, f.deg(), g.lv, g.deg(), forg.deg(), gorg.deg())
				return nil, forg, gorg
			}
		case NObj:
			if r.IsZero() {
				fmt.Printf("gcd : f =[%d,%3d], g =[%d,%3d] <%3d,%3d> END3\n", f.lv, f.deg(), g.lv, g.deg(), forg.deg(), gorg.deg())
				return g, s2, t2
			} else {
				fmt.Printf("gcd : f =[%d,%3d], g =[%d,%3d] <%3d,%3d> END4\n", f.lv, f.deg(), g.lv, g.deg(), forg.deg(), gorg.deg())
				return nil, forg, gorg
			}
		default:
			panic("??")
		}
	}
}

func (cad *CAD) sym_sqfr2(porg *Poly, cell *Cell) []*cadSqfr {
	// sqfr using modular GCD
	// p65 Fundamentals of Computer Algebra
	p := porg
	if !p.isIntPoly() {
		panic("unexpected")
	}

	pd := porg.diff(porg.lv).(*Poly)
	s0, t0 := cad.symde_gcd2(p, pd, cell, 0)
	if s0 == nil { // gcd=1 => square-free
		return []*cadSqfr{newCadSqfr(nil, porg, 1)}
	}

	ret := make([]*cadSqfr, 0)
	for i := mult_t(1); t0 != nil; i++ {
		if s0 == nil {
			if t0.lv == porg.lv {
				ret = append(ret, newCadSqfr(cell, t0, i))
			}
			break
		}

		ui, si := cad.symde_gcd2(s0, t0, cell, 0)
		_, fi, _ := t0.pquorem(ui)
		if ff, ok := fi.(*Poly); ok && ff.lv == porg.lv {
			ret = append(ret, newCadSqfr(cell, ff, i))
		}
		s0 = si
		t0 = ui
	}
	d := 0
	for _, r := range ret {
		d += r.p.deg() * int(r.r)
	}
	if d != porg.deg() {
		panic(fmt.Sprintf("invalid... deg(p)=%d, sqfr=%d, len=%d", porg.deg(), d, len(ret)))
	}

	return ret
}

func (cad *CAD) sym_sqfr(porg *Poly, cell *Cell) []*cadSqfr {
	fmt.Printf("    sym_sqfr(%v) %v\n", cell.Index(), porg)
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
	var i mult_t
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
