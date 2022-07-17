package ganrac

import (
	"fmt"
	"math/big"
)

// go test -bench  BenchmarkModularMulPoly -run BenchmarkModularMulPoly -o prof/ganrac -count=N
// 3x20= 34.222
// 4x20= 30.380
// 5x20= 31.947
const KARATSUBA_DEG_MOD = 4

type Cellmod struct {
	cell    *Cell // 元
	defpoly *Poly
	factor1 *Poly
	factor2 *Poly
	lv      Level
	de      bool
	parent  *Cellmod
}

type Moder interface {
	// modular arithmetic
	RObj
	add_mod(g Moder, p Uint) Moder
	sub_mod(g Moder, p Uint) Moder
	mul_mod(g Moder, p Uint) Moder
	mul_uint_mod(g Uint, p Uint) Moder
	neg_mod(p Uint) Moder
	inv_mod(cell *Cellmod, p Uint) Moder
	simpl_mod(cell *Cellmod, p Uint) Moder
	valid_mod(cell *Cellmod, p Uint) error
	deg() int
}

func NewCellmod(cell *Cell) *Cellmod {
	c := new(Cellmod)
	c.cell = cell
	c.lv = cell.lv
	c.de = true
	return c
}

func (q Uint) deg() int {
	if q == 0 {
		return -1
	} else {
		return 0
	}
}

func (q Uint) valid_mod(cell *Cellmod, p Uint) error {
	if q >= p {
		return fmt.Errorf("q=%v, p=%v", q, p)
	}
	return nil
}

func (q *Poly) valid_mod(cell *Cellmod, p Uint) error {
	// 次数は定義多項式以下
	// 係数は p 未満
	for cell != nil && cell.lv > q.lv {
		cell = cell.parent
	}

	if cell != nil && q.lv <= cell.lv {
		if q.deg() >= cell.defpoly.deg() {
			return fmt.Errorf("deg invalid. lv=%d, deg=%d, defpoly=%d", q.lv, q.deg(), cell.defpoly.deg())
		}
	}

	for i, c := range q.c {
		cm, ok := c.(Moder)
		if !ok {
			return fmt.Errorf("coef[lv=%d, deg=%d] not moder %v", q.lv, i, q)
		}
		err := cm.valid_mod(cell, p)
		if err != nil {
			return fmt.Errorf("coef[lv=%d, deg=%d] %w\n", q.lv, i, err)
		}
	}

	return q.valid()
}

func (p Uint) String() string {
	return fmt.Sprintf("%d", p)
}

func (p Uint) Tag() uint {
	return TAG_NUM
}

func (p Uint) IsNumeric() bool {
	return true
}

func (p Uint) valid() error {
	return nil
}

func (p Uint) Format(s fmt.State, format rune) {
	fmt.Fprintf(s, "%d", uint32(p))
}

func (p Uint) IsZero() bool {
	return p == 0
}
func (p Uint) IsOne() bool {
	return p == 1
}
func (p Uint) IsMinusOne() bool {
	return false
}
func (p Uint) mul_2exp(m uint) RObj {
	panic("unsupported")
}
func (p Uint) toIntv(prec uint) RObj {
	panic("unsupported")
}

func (p Uint) Sign() int {
	if p == 0 {
		return 0
	} else {
		return 1
	}
}

func (u Uint) simpl_mod(cell *Cellmod, p Uint) Moder {
	return u
}

func (u Uint) inv_mod(cell *Cellmod, p Uint) Moder {
	// p: prime number
	// return v where (u*v == 1 mod p)
	s1 := Uint(1)
	s0 := Uint(0)

	f0 := p
	f1 := u

	for {
		q := f0 / f1
		f1, f0 = f0%f1, f1
		if f1 == 0 {
			return s1
		}
		s1, s0 = s0.sub_mod(s1.mul_uint_mod(q, p), p).(Uint), s1
	}
}

func (p Uint) Add(x RObj) RObj                   { panic("no!") }
func (p Uint) Sub(x RObj) RObj                   { panic("no!") }
func (p Uint) Mul(x RObj) RObj                   { panic("no!") }
func (p Uint) Div(x NObj) RObj                   { panic("no!") }
func (p Uint) Pow(x *Int) RObj                   { panic("no!") }
func (p Uint) Subst(x RObj, lv Level) RObj       { panic("no!") }
func (p Uint) Neg() RObj                         { panic("no!") }
func (p Uint) numTag() uint                      { panic("no!") }
func (p Uint) Float() float64                    { panic("no!") }
func (p Uint) Cmp(x NObj) int                    { panic("no!") }
func (p Uint) CmpAbs(x NObj) int                 { panic("no!") }
func (p Uint) Abs() NObj                         { return p }
func (p Uint) subst_poly(f *Poly, lv Level) RObj { panic("no!") }

func (p Uint) Equals(v interface{}) bool {
	if vv, ok := v.(Uint); ok {
		return vv == p
	}
	return false
}

func (f *Poly) mod(p Uint) Moder {
	pp := big.NewInt(int64(p))
	return f.mod_pp(pp, new(big.Int))
}

func (f *Poly) mod_pp(pp, wk *big.Int) Moder {
	g := NewPoly(f.lv, len(f.c))
	for i, cc := range f.c {
		switch c := cc.(type) {
		case *Poly:
			g.c[i] = c.mod_pp(pp, wk)
		case *Int:
			wk.Mod(c.n, pp)
			g.c[i] = Uint(wk.Int64())
		default:
			panic(fmt.Sprintf("internal error: [%d] %v", i, cc))
		}
	}
	return g.normalize().(Moder)
}

// d'th coefficient for Moder
func (f *Poly) mcoef(d int) Moder {
	return f.c[d].(Moder)
}

func (f Uint) add_mod(gg Moder, p Uint) Moder {
	switch g := gg.(type) {
	case Uint:
		if f.IsZero() {
			return gg
		}
		r := f + g
		if r >= p {
			r -= p
		}
		return r
	case *Poly:
		z := g.Clone()
		z.c[0] = f.add_mod(g.mcoef(0), p)
		return z
	default:
		panic("internal error")
	}
}

func (f *Poly) add_mod(gg Moder, p Uint) Moder {
	switch g := gg.(type) {
	case Uint:
		z := f.Clone()
		z.c[0] = g.add_mod(z.mcoef(0), p)
		return z
	case *Poly:
		if f.lv < g.lv {
			z := g.Clone()
			z.c[0] = f.add_mod(z.mcoef(0), p)
			return z
		} else if f.lv > g.lv {
			z := f.Clone()
			z.c[0] = g.add_mod(z.mcoef(0), p)
			return z
		} else {
			var dmin int
			var q *Poly
			if len(f.c) < len(g.c) {
				dmin = len(f.c)
				q = g
			} else {
				dmin = len(g.c)
				q = f
			}
			z := q.Clone()
			for i := 0; i < dmin; i++ {
				switch fc := f.c[i].(type) {
				case *Poly:
					z.c[i] = fc.add_mod(g.mcoef(i), p)
				case Uint:
					z.c[i] = fc.add_mod(g.mcoef(i), p)
				default:
					panic("un")
				}
			}
			return z.normalize().(Moder)
		}
	default:
		panic("")
	}
}

func (f Uint) sub_mod(gg Moder, p Uint) Moder {
	switch g := gg.(type) {
	case *Poly:
		z := NewPoly(g.lv, len(g.c))
		z.c[0] = f.sub_mod(g.mcoef(0), p)
		for i := 1; i < len(g.c); i++ {
			z.c[i] = g.mcoef(i).neg_mod(p)
		}
		return z
	case Uint:
		if f < g {
			return p + f - g
		} else {
			return f - g
		}
	default:
		panic("internal error")
	}
}

func (f *Poly) sub_mod(gg Moder, p Uint) Moder {
	switch g := gg.(type) {
	case *Poly:
		if f.lv < g.lv {
			z := NewPoly(g.lv, len(g.c))
			z.c[0] = f.sub_mod(g.mcoef(0), p)
			for i := 1; i < len(g.c); i++ {
				z.c[i] = g.c[i].(Moder).neg_mod(p)
			}
			return z
		} else if f.lv > g.lv {
			z := f.Clone()
			z.c[0] = f.mcoef(0).sub_mod(g, p)
			return z
		} else {
			var dmin int
			var z *Poly
			if len(f.c) <= len(g.c) {
				dmin = len(f.c)
				z = NewPoly(g.lv, len(g.c))
				for i := dmin; i < len(g.c); i++ {
					z.c[i] = g.mcoef(i).neg_mod(p)
				}
			} else {
				dmin = len(g.c)
				z = NewPoly(f.lv, len(f.c))
				copy(z.c[dmin:], f.c[dmin:])
			}
			for i := 0; i < dmin; i++ {
				z.c[i] = f.mcoef(i).sub_mod(g.mcoef(i), p)
			}
			return z.normalize().(Moder)
		}
	case Uint:
		if g.IsZero() {
			return f
		}
		z := f.Clone()
		z.c[0] = z.mcoef(0).sub_mod(g, p)
		return z
	default:
		panic("internal error")
	}
}

func (f Uint) neg_mod(p Uint) Moder {
	if f == 0 {
		return f
	}
	return p - f
}

func (f *Poly) neg_mod(p Uint) Moder {
	z := f.Clone()
	for i, c := range z.c {
		z.c[i] = c.(Moder).neg_mod(p)
	}
	return z
}

func (f Uint) mul_uint_mod(g Uint, p Uint) Moder {
	return Uint((uint64(f) * uint64(g)) % uint64(p))
}

func (f Uint) mul_mod(gg Moder, p Uint) Moder {
	if f == 0 {
		return f
	} else if f == 1 {
		return gg
	}

	switch g := gg.(type) {
	case Uint:
		return g.mul_uint_mod(f, p)
	case *Poly:
		return g.mul_uint_mod(f, p)
	default:
		panic("internal error")
	}
}

func (f *Poly) mul_uint_mod(g Uint, p Uint) Moder {
	if g == 0 {
		return g
	} else if g == 1 {
		return f
	}
	z := NewPoly(f.lv, len(f.c))
	for i, cc := range f.c {
		z.c[i] = cc.(Moder).mul_uint_mod(g, p)
	}
	return z
}

func (f *Poly) karatsuba_divide_mod(d int) (Moder, Moder) {
	f1, f0 := f.karatsuba_divide(d)
	var p1, p0 Moder
	if x, ok := f1.(Moder); ok {
		p1 = x
	} else if f1 == zero {
		p1 = Uint(0)
	} else {
		panic(fmt.Sprintf("internal error %v", f1))
	}
	if x, ok := f0.(Moder); ok {
		p0 = x
	} else if f0 == zero {
		p0 = Uint(0)
	} else {
		panic(fmt.Sprintf("internal error %v", f0))
	}
	return p1, p0
}

func (f *Poly) karatsuba_mod(g *Poly, p Uint) Moder {
	// returns f*g mod p
	// assert f.lv = g.lv
	// assert len(f.c) > KARATSUBA_DEG_MOD
	// assert len(g.c) > KARATSUBA_DEG_MOD
	var d int
	if len(f.c) > len(g.c) {
		d = len(f.c) / 2
	} else {
		d = len(g.c) / 2
	}
	f1, f0 := f.karatsuba_divide_mod(d)
	g1, g0 := g.karatsuba_divide_mod(d)

	f1g1 := f1.mul_mod(g1, p)
	f0g0 := f0.mul_mod(g0, p)
	f10 := f1.sub_mod(f0, p)
	g10 := g0.sub_mod(g1, p)
	fg := f10.mul_mod(g10, p)
	fg = fg.add_mod(f1g1, p)
	fg = fg.add_mod(f0g0, p)

	d2 := 2 * d
	var cf1g1 []RObj
	if ptmp, ok := f1g1.(*Poly); ok && ptmp.lv == f.lv {
		d2 += len(ptmp.c)
		cf1g1 = ptmp.c
	} else {
		cf1g1 = []RObj{f1g1}
	}
	var cf0g0 []RObj
	if ptmp, ok := f0g0.(*Poly); ok && ptmp.lv == f.lv {
		cf0g0 = ptmp.c
	} else {
		cf0g0 = []RObj{f0g0}
	}
	dx := -1
	if q, ok := fg.(*Poly); ok && q.lv == f.lv {
		dx = len(q.c)
	}

	dd := maxint(2*d+len(cf1g1), d+dx)
	ret := NewPoly(f.lv, dd)
	for i := 0; i < len(ret.c); i++ {
		ret.c[i] = Uint(0)
	}
	copy(ret.c, cf0g0)
	copy(ret.c[2*d:], cf1g1)

	if q, ok := fg.(*Poly); ok && q.lv == f.lv {
		for i := 0; i < len(q.c); i++ {
			ret.c[i+d] = q.c[i].(Moder).add_mod(ret.c[i+d].(Moder), p)
		}
	} else {
		ret.c[d] = fg.add_mod(ret.c[d].(Moder), p)
	}

	return ret
}

func (f *Poly) mul_poly_mod(g *Poly, p Uint) Moder {
	if f.lv < g.lv {
		z := NewPoly(g.lv, len(g.c))
		for i := range z.c {
			z.c[i] = f.mul_mod(g.mcoef(i), p)
		}
		if err := z.valid(); err != nil {
			panic(fmt.Sprintf("invalid 1 %v\nz=%v\nf=%v\ng=%v\n", err, z, f, g))
		}
		return z
	} else if f.lv > g.lv {
		return g.mul_poly_mod(f, p)
	}
	if len(f.c) > KARATSUBA_DEG_MOD && len(g.c) > KARATSUBA_DEG_MOD {
		return f.karatsuba_mod(g, p)
	} else {
		return f.mul_poly_mod_basic(g, p)
	}
}

func (f *Poly) mul_poly_mod_basic(g *Poly, p Uint) Moder {
	z := NewPoly(f.lv, len(f.c)+len(g.c)-1)
	for i := range z.c {
		z.c[i] = Uint(0)
	}
	for i, c := range f.c {
		if c.IsZero() {
			continue
		}
		fig := g.mul_mod(c.(Moder), p).(*Poly)
		for j := range fig.c {
			c := fig.c[j].(Moder)
			z.c[i+j] = c.add_mod(z.mcoef(i+j), p)
		}
	}

	if err := z.valid(); err != nil {
		panic(fmt.Sprintf("invalid 3 %v\nz=%v\nf=%v\ng=%v\n", err, z, f, g))
	}
	return z
}

func (f *Poly) mul_mod(gg Moder, p Uint) Moder {
	switch g := gg.(type) {
	case Uint:
		return f.mul_uint_mod(g, p)
	case *Poly:
		return f.mul_poly_mod(g, p)
	default:
		panic("")
	}
}

func (gorg *Poly) monicize(cell *Cellmod, p Uint) (Moder, Moder, Moder) {
	// returns (c.g, c, g) where c.g is monic && lv(c) < lv(g)
	// g は数だったかもしれないし，適切に簡略化されたもの.
	g := gorg
	if err := gorg.valid(); err != nil {
		panic(fmt.Sprintf("monicize: %v: %v\n", gorg, err))
	}
	switch lc := g.lc().(type) {
	case Uint:
		if !lc.IsOne() {
			if lc == 0 {
				panic(fmt.Sprintf("not normalized: %v\n", gorg))
			}
			inv := lc.inv_mod(cell, p)
			g = g.mul_mod(inv, p).(*Poly)
			return g, inv, gorg
		} else {
			return g, Uint(1), gorg
		}
	case *Poly:
		inv := lc.inv_mod(cell, p)
		if inv == nil {
			// 定義多項式が因数分解されてしまった?
			return nil, nil, gorg
		}
		if inv.IsZero() {
			// 主係数は 0 だった.
			g.c[g.deg()] = Uint(0)
			switch gg := g.normalize().(type) {
			case *Poly:
				if err := gg.valid(); err != nil {
					panic(fmt.Sprintf("monicize: %v: %v\n", gg, err))
				}
				if gg.lv == gorg.lv {
					return gg.monicize(cell, p)
				}
				// 数になった.
				return gg, Uint(1), gg
			case Uint:
				return gg, Uint(1), gg
			}
		}
		gg := g.mul_mod(inv, p).simpl_mod(cell, p)
		if err := gorg.valid(); err != nil {
			panic(fmt.Sprintf("monicize: %v: %v\n", gorg, err))
		}
		return gg, inv, gorg
	}
	panic("?")
}

func (f *Poly) divmod_poly_mod(gorg *Poly, cell *Cellmod, p Uint) (Moder, Moder, Moder) {
	// returns (q, r, g) where f = q * g + r && deg(r) < deg(g)
	// assume: f.lv == gorg.lv

	if err := f.valid(); err != nil {
		panic(fmt.Sprintf("invalid f %v: %v %v\n", f, err, []int{deg(f), deg(gorg)}))
	}
	if err := gorg.valid(); err != nil {
		panic(fmt.Sprintf("invalid g %v: %v %v\n", gorg, err, []int{deg(f), deg(gorg)}))
	}

	if len(f.c) < len(gorg.c) {
		return Uint(0), f, gorg
	}

	///////////////////////////
	// g を monic にする
	///////////////////////////
	g := gorg
	_g, _inv, gc := g.monicize(cell, p)
	switch gg := _g.(type) {
	case *Poly:
		if gg.lv != gorg.lv {
			return Uint(0), f, gc
		}
		g = gg
	case Uint:
		if gg == 0 {
			return gg, gg, gc
		}
		return Uint(0), f, gc
	default:
		// 定義多項式が分解された
		return nil, nil, gc
	}

	// g に inv かけたから, f にもかけよう
	switch inv := _inv.(type) {
	case Uint:
		if inv != 1 {
			f = f.mul_mod(inv, p).(*Poly)
		}
	case *Poly:
		switch ff := f.mul_mod(inv, p).simpl_mod(cell, p).(type) {
		case *Poly:
			f = ff
			if f.lv != g.lv {
				return Uint(0), f, gc
			}
		case Uint:
			return Uint(0), ff, gc
		}
	}

	// 以下では復帰前に剰余に対して gorg.lc() をかける必要あり
	// F = f*inv, G = g*inv
	//     F = Q * G + R
	// inv*f = Q * inv*g + R
	//     f = Q * g + R*lc
	q := NewPoly(f.lv, len(f.c)-len(g.c)+1)
	for i := range q.c {
		q.c[i] = Uint(0)
	}
	for j := len(f.c); f.lv == g.lv && len(f.c) >= len(g.c); j-- {
		dd := len(f.c) - len(g.c)
		q.c[dd] = f.lc()
		var gg Moder
		if dd == 0 {
			gg = g.mul_mod(q.c[dd].(Moder), p)
		} else {
			cxn := NewPoly(f.lv, dd+1)
			for i := range cxn.c {
				cxn.c[i] = Uint(0)
			}
			cxn.c[dd] = f.lc()
			gg = g.mul_poly_mod(cxn, p)
		}
		switch f2 := f.sub_mod(gg, p).(type) {
		case *Poly:
			if f.lv == f2.lv && len(f.c) <= len(f2.c) {
				fmt.Printf("    f=%v\n", f)
				fmt.Printf("    g=%v\n", g)
				fmt.Printf("    F=%v\n", f2)
				fmt.Printf("    G=%v\n", gg)
				panic(fmt.Sprintf("why? mod=%d", p))
			}
			f = f2
		case Uint:
			rr := f2.mul_mod(gorg.lc().(Moder), p)
			if err := rr.valid_mod(cell, p); err != nil {
				panic(fmt.Sprintf("invalid rr1 %v: %v,  <%v,%v>\n", rr, err, f, gorg))
			}
			return q.normalize().(Moder), rr.(Moder), gc
		}
	}
	if err := f.valid(); err != nil {
		panic(fmt.Sprintf("invalid f %v: %v,  <%v,%v>\n", f, err, f, gorg))
	}
	rr := f.mul_mod(gorg.lc().(Moder), p)
	rr = rr.simpl_mod(cell, p)
	if err := rr.valid_mod(cell, p); err != nil {
		panic(fmt.Sprintf("invalid rr1 %v: %v,  <%v,%v>\n", rr, err, f, gorg))
	}

	return q.normalize().(Moder), rr, gc
}

func (u *Poly) simpl_mod(cell *Cellmod, p Uint) Moder {
	if err := u.valid(); err != nil {
		panic(fmt.Sprintf("invalid u %v: %v\n", u, err))
	}
	if cell == nil {
		return u
	}
	for cell.lv > u.lv {
		cell = cell.parent
	}
	if cell.lv == u.lv {
		_, vv, _ := u.divmod_poly_mod(cell.defpoly, cell.parent, p)
		if vv == nil {
			return nil
		}
		switch v := vv.(type) {
		case Uint:
			return v
		case *Poly:
			return vv.simpl_mod(cell.parent, p)
		}
		panic("?")
	}
	v := NewPoly(u.lv, len(u.c))
	for i := range v.c {
		v.c[i] = u.mcoef(i).simpl_mod(cell, p)
	}
	return v.normalize().(Moder)
}

func (f *Poly) inv_mod(cell *Cellmod, p Uint) Moder {
	// cell.mod() 済みと仮定
	// assume: f.lv <= cell.lv
	// return 1/f mod (cell,p)
	// return nil if 定義多項式が因数分解された

	var s1 Moder = Uint(1)
	var s0 Moder = Uint(0)

	for ; cell.lv > f.lv; cell = cell.parent {
	}
	if cell == nil || cell.lv != f.lv {
		panic(fmt.Sprintf("no-cell... %d: f=%v", cell.lv, f))
	}

	f0 := cell.defpoly
	f1 := f
	for i := len(f.c) * 3; i >= 0; i-- { // 無限ループ対策...... 適当すぎるが @TODO
		q, r, _ := f0.divmod_poly_mod(f1, cell.parent, p)
		if q == nil {
			return nil
		}
		if r.IsZero() {
			// 共通根を持った...
			// 次数が同じ場合は?
			if f1.deg() < cell.defpoly.deg() {
				s1 = s0.sub_mod(s1.mul_mod(q, p), p)

				cell.factor1 = f1
				cell.factor2 = s1.(*Poly)
				return nil
			} else {
				return Uint(0)
			}
		}
		s1, s0 = s0.sub_mod(s1.mul_mod(q, p), p), s1
		switch rr := r.(type) {
		case *Poly:
			if rr.lv != f1.lv {
				rinv := rr.inv_mod(cell.parent, p)
				if rinv == nil {
					return rinv
				}
				return s1.mul_mod(rinv, p)
			}

			f1, f0 = rr, f1
		case Uint:
			if rr == 1 {
				return s1
			} else {
				rinv := rr.inv_mod(cell, p)
				if rinv == nil {
					return rinv
				}
				return s1.mul_mod(rinv, p)
			}
		}
	}

	fmt.Printf("inv_mod(%d)    f1=%v\n", p, f1)
	fmt.Printf("inv_mod(%d)    rr=%v\n", p, f)
	cell.cell.Print("cellp")
	panic("why?")
}

func (f *Poly) mod2int() *Poly {
	// mod 多項式を int に変換
	g := NewPoly(f.lv, len(f.c))
	for i, cc := range f.c {
		switch c := cc.(type) {
		case *Poly:
			g.c[i] = c.mod2int()
		case Uint:
			g.c[i] = NewInt(int64(c))
		}
	}
	return g
}

type pqinf_interpol_t struct {
	p    *Int
	q    *Int
	qui  Uint
	pq   *Int
	pinv Uint
	wk   *big.Int
}

func (f *Int) interpol_poly(g *Poly, pqinf *pqinf_interpol_t) *Poly {
	z := NewPoly(g.lv, len(g.c))
	for i := range g.c {
		switch gc := g.c[i].(type) {
		case *Poly:
			z.c[i] = f.interpol_poly(gc, pqinf)
		case Uint:
			z.c[i] = f.interpol_ui(gc, pqinf)
		}
		f = zero
	}
	return z
}

func (f *Int) interpol_ui(g Uint, pqinf *pqinf_interpol_t) *Int {
	pqinf.wk.Mod(f.n, pqinf.q.n)
	ff := Uint(pqinf.wk.Int64())
	fg := g.sub_mod(ff, pqinf.qui)
	v := pqinf.pinv.mul_mod(fg, pqinf.qui).(Uint)
	return NewInt(int64(v)).Mul(pqinf.p).Add(f).(*Int)
}

func (f *Poly) interpol_ui(g Uint, pqinf *pqinf_interpol_t) *Poly {
	z := NewPoly(f.lv, len(f.c))
	for i := range f.c {
		switch fc := f.c[i].(type) {
		case *Poly:
			z.c[i] = fc.interpol_ui(g, pqinf)
		case *Int:
			z.c[i] = fc.interpol_ui(g, pqinf)
		case Uint:
			ff := NewInt(int64(fc))
			z.c[i] = ff.interpol_ui(g, pqinf)
		}
		g = 0
	}
	return z
}

func (f *Poly) interpol_poly(g *Poly, pqinf *pqinf_interpol_t) *Poly {
	// g は Uint を係数にもつ多変数多項式
	// f は *Int か Uint を係数にもつ多変数多項式
	if f.lv < g.lv {
		fp := NewPoly(g.lv, len(g.c))
		var c RObj = f
		for i := range g.c {
			fp.c[i] = c
			c = zero
		}
		f = fp
	} else if f.lv > g.lv {
		gp := NewPoly(f.lv, len(f.c))
		var c RObj = g
		for i := range f.c {
			gp.c[i] = c
			c = Uint(0)
		}
		g = gp
	}

	var fcs []RObj = f.c
	var gcs []RObj = g.c

	if len(g.c) < len(f.c) {
		gcs = make([]RObj, len(f.c))
		copy(gcs, g.c)
		for i := len(g.c); i < len(f.c); i++ {
			gcs[i] = Uint(0)
		}
	} else if len(g.c) > len(f.c) {
		fcs = make([]RObj, len(g.c))
		copy(fcs, f.c)
		for i := len(f.c); i < len(g.c); i++ {
			fcs[i] = zero
		}
	}

	z := NewPoly(f.lv, len(fcs))
	for i := range fcs {
		switch fc := fcs[i].(type) {
		case *Poly:
			switch gc := gcs[i].(type) {
			case *Poly:
				z.c[i] = fc.interpol_poly(gc, pqinf)
			case Uint:
				z.c[i] = fc.interpol_ui(gc, pqinf)
			}
		case *Int:
			switch gc := gcs[i].(type) {
			case *Poly:
				z.c[i] = fc.interpol_poly(gc, pqinf)
			case Uint:
				z.c[i] = fc.interpol_ui(gc, pqinf)
			}
		case Uint:
			ff := NewInt(int64(fc))
			switch gc := gcs[i].(type) {
			case *Poly:
				z.c[i] = ff.interpol_poly(gc, pqinf)
			case Uint:
				z.c[i] = ff.interpol_ui(gc, pqinf)
			}
		}
	}
	return z
}

func (p *Int) _crt_init(q Uint) *pqinf_interpol_t {
	qi := NewInt(int64(q))
	_, s, _ := p.GcdEx(qi)
	pq := p.Mul(qi).(*Int)
	ss := s.Int64()
	if ss < 0 {
		ss += int64(q)
	}
	pqinf := new(pqinf_interpol_t)
	pqinf.wk = new(big.Int)
	pqinf.p = p
	pqinf.q = qi
	pqinf.qui = q
	pqinf.pq = pq
	pqinf.pinv = Uint(ss)

	return pqinf
}

func (f *Poly) crt_interpol(frr *Poly, g *Poly, p *Int, q Uint) (*Poly, *Poly, *Int, bool) {
	// returns (新しいの, 新しいのRR, p*q, 更新されたか)
	if !g.lc().IsOne() {
		panic("invalid")
	}

	pqinf := p._crt_init(q)

	f2 := f.interpol_poly(g, pqinf)

	bound := newInt()
	bound.n.Quo(pqinf.pq.n, two.n)
	bound.n.Sqrt(bound.n)
	fr2 := f2.i2q(pqinf.pq, bound)
	if fr2 != nil {
		fr2, _ = fr2.pp()
	}
	return f2, fr2, pqinf.pq, frr != nil && fr2 != nil && frr.Equals(fr2)
}

func (f *Poly) i2q(p, b *Int) *Poly {
	q := NewPoly(f.lv, len(f.c))
	for i := range f.c {
		switch c := f.c[i].(type) {
		case *Poly:
			qq := c.i2q(p, b)
			if qq == nil {
				return nil
			}
			q.c[i] = qq
		case *Int:
			q.c[i] = c.i2q(p, b)
		default:
			panic("internal error")
		}
		if q.c[i] == nil {
			return nil
		}
	}

	return q
}

func (x *Int) i2q(p, b *Int) RObj {
	// rational reconstruction
	// p > x, gcd(p, x) = 1
	// b = sqrt(p/2)
	// returns bx = a mod p && |a|,|b| < sqrt(p/2)
	// returns nil
	if x.IsZero() {
		return x
	}
	r0 := p.n
	r1 := x.n
	t1 := one.n
	t0 := zero.n

	for r1.Cmp(b.n) > 0 {
		q := new(big.Int)
		r := new(big.Int)
		q.QuoRem(r0, r1, r)

		r0, r1 = r1, r

		r = new(big.Int)
		r.Mul(q, t1)
		r.Sub(t0, r)

		t0, t1 = t1, r
	}
	if t1.Cmp(b.n) <= 0 {
		r := new(big.Int)
		sgn := 1
		// Go1.14 前は r1, t1 正を要求
		if r1.Sign() < 0 {
			r1.Neg(r1)
			sgn *= -1
		}
		if t1.Sign() < 0 {
			t1.Neg(t1)
			sgn *= -1
		}
		r.GCD(nil, nil, r1, t1)
		if r.CmpAbs(one.n) == 0 {
			if sgn < 0 {
				r1.Neg(r1)
			}
			if t1.Cmp(one.n) == 0 {
				q := newInt()
				q.n.Set(r1)
				return q
			} else {
				q := newRat()
				q.n.SetFrac(r1, t1)
				return q
			}
		}
	}
	return nil
}
