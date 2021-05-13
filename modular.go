package ganrac

import (
	"fmt"
	"math/big"
)

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
	neg_mod(p Uint) Moder
	inv_mod(cell *Cellmod, p Uint) Moder
	simpl_mod(cell *Cellmod, p Uint) Moder
	valid_mod(cell *Cellmod, p Uint) error
}

func NewCellmod(cell *Cell) *Cellmod {
	c := new(Cellmod)
	c.cell = cell
	c.lv = cell.lv
	c.de = true
	return c
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
	for cell.lv > q.lv {
		cell = cell.parent
	}

	if q.lv <= cell.lv {
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
	fmt.Fprintf(s, "%"+string(format), uint32(p))
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

func (p Uint) Add(x RObj) RObj                        { panic("no!") }
func (p Uint) Sub(x RObj) RObj                        { panic("no!") }
func (p Uint) Mul(x RObj) RObj                        { panic("no!") }
func (p Uint) Div(x NObj) RObj                        { panic("no!") }
func (p Uint) Pow(x *Int) RObj                        { panic("no!") }
func (p Uint) Subst(x []RObj, lv []Level, n int) RObj { panic("no!") }
func (p Uint) Neg() RObj                              { panic("no!") }
func (p Uint) numTag() uint                           { panic("no!") }
func (p Uint) Float() float64                         { panic("no!") }
func (p Uint) Cmp(x NObj) int                         { panic("no!") }
func (p Uint) CmpAbs(x NObj) int                      { panic("no!") }
func (p Uint) Abs() NObj                              { return p }
func (p Uint) subst_poly(f *Poly, lv Level) RObj      { panic("no!") }

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
		if cc == nil {
			panic("why?")
		}
		switch c := cc.(type) {
		case *Poly:
			g.c[i] = c.mod_pp(pp, wk)
		case *Int:
			wk.Mod(c.n, pp)
			g.c[i] = Uint(wk.Int64())
		default:
			panic("unsupported")
		}
	}
	return g.normalize().(Moder)
}

func (f *Poly) mcoef(d int) Moder {
	return f.c[d].(Moder)
}

func (f Uint) add_mod(gg Moder, p Uint) Moder {
	if f == 0 {
		return gg
	}
	switch g := gg.(type) {
	case Uint:
		r := f + g
		if r >= p {
			r -= p
		}
		return r
	case *Poly:
		z := g.copy()
		z.c[0] = f.add_mod(g.mcoef(0), p)
		return z
	default:
		panic("unsupported")
	}
}

func (f *Poly) add_mod(gg Moder, p Uint) Moder {
	switch g := gg.(type) {
	case Uint:
		z := f.copy()
		z.c[0] = g.add_mod(z.mcoef(0), p)
		return z
	case *Poly:
		if f.lv < g.lv {
			z := g.copy()
			z.c[0] = f.add_mod(z.mcoef(0), p)
			return z
		} else if f.lv > g.lv {
			z := f.copy()
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
			z := q.copy()
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
	g := gg.(Moder).neg_mod(p)
	return f.add_mod(g, p)
}

func (f *Poly) sub_mod(gg Moder, p Uint) Moder {
	g := gg.(Moder).neg_mod(p)
	return f.add_mod(g, p)
}

func (f Uint) neg_mod(p Uint) Moder {
	if f == 0 {
		return f
	}
	return p - f
}

func (f *Poly) neg_mod(p Uint) Moder {
	z := f.copy()
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
		panic("")
	}
}

func (f *Poly) mul_uint_mod(g Uint, p Uint) Moder {
	z := NewPoly(f.lv, len(f.c))
	for i, cc := range f.c {
		switch c := cc.(type) {
		case *Poly:
			z.c[i] = c.mul_uint_mod(g, p)
		case Uint:
			z.c[i] = c.mul_uint_mod(g, p)
		default:
			panic("")
		}
	}
	return z
}

func (f *Poly) mul_poly_mod(g *Poly, p Uint) Moder {
	if f.lv < g.lv {
		z := NewPoly(g.lv, len(g.c))
		for i := range z.c {
			z.c[i] = f.mul_mod(g.mcoef(i), p)
		}
		return z
	} else if f.lv > g.lv {
		z := NewPoly(f.lv, len(f.c))
		for i := range z.c {
			z.c[i] = g.mul_mod(f.mcoef(i), p)
		}
		return z
	}
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

func (f *Poly) divmod_poly_mod(gorg *Poly, cell *Cellmod, p Uint) (Moder, Moder) {
	// assume: f.lv == gorg.lv

	if len(f.c) < len(gorg.c) {
		return Uint(0), f
	}

	///////////////////////////
	// g を monic にする
	///////////////////////////
	g := gorg
	switch lc := g.lc().(type) {
	case Uint:
		if !lc.IsOne() {
			inv := lc.inv_mod(cell, p)
			g = g.mul_mod(inv, p).(*Poly)
			f = f.mul_mod(inv, p).(*Poly)
		}
	case *Poly:
		inv := lc.inv_mod(cell, p)
		if inv == nil {
			// 定義多項式が因数分解されてしまった?
			return nil, nil
		}
		if inv.IsZero() {
			// 主係数は 0 だった.
			gorg.c[gorg.deg()] = Uint(0)
			switch gg := gorg.normalize().(type) {
			case *Poly:
				if gg.lv == gorg.lv {
					return f.divmod_poly_mod(gg, cell, p)
				}
				// 数になった.
				return Uint(0), f
			case Uint:
				return Uint(0), f
			}
		}
		g = g.mul_mod(inv, p).simpl_mod(cell, p).(*Poly)
		switch ff := f.mul_mod(inv, p).simpl_mod(cell, p).(type) {
		case *Poly:
			f = ff
			if f.lv != g.lv {
				return Uint(0), f
			}
		case Uint:
			return Uint(0), ff
		}
	}

	q := NewPoly(f.lv, len(f.c)-len(g.c)+1)
	for i := range q.c {
		q.c[i] = Uint(0)
	}
	for j := len(f.c); f.lv == g.lv && len(f.c) >= len(g.c); j-- {
		dd := len(f.c) - len(g.c)
		q.c[dd] = f.lc()
		cxn := NewPoly(f.lv, dd+1)
		for i := range cxn.c {
			cxn.c[i] = Uint(0)
		}
		cxn.c[dd] = f.lc()
		gg := g.mul_poly_mod(cxn, p)
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
			return q.normalize().(Moder), rr.(Moder)
		}
	}
	rr := f.mul_mod(gorg.lc().(Moder), p).(Moder)
	rr = rr.simpl_mod(cell, p)

	return q.normalize().(Moder), rr
}

func (u *Poly) simpl_mod(cell *Cellmod, p Uint) Moder {
	if cell == nil {
		return u
	}
	for cell.lv > u.lv {
		cell = cell.parent
	}
	if cell.lv == u.lv {
		_, vv := u.divmod_poly_mod(cell.defpoly, cell.parent, p)
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

	f0 := cell.defpoly
	f1 := f
	for i := len(f.c); i >= 0; i-- {
		q, r := f0.divmod_poly_mod(f1, cell.parent, p)
		if q == nil {
			return nil
		}
		if r.IsZero() {
			// 共通根を持った...
			// 次数が同じ場合は?
			if f1.deg() < cell.defpoly.deg() {
				cell.factor1 = f1
				cell.factor2 = q.(*Poly)
				return nil
			} else {
				return Uint(0)
			}
		}
		s1, s0 = s0.sub_mod(s1.mul_mod(q, p), p), s1
		switch rr := r.(type) {
		case *Poly:
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

	fmt.Printf("    f=%v\n", f)
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

func (f *Poly) crt_interpol(g *Poly, p *Int, q Uint) (*Poly, *Int, bool) {
	// returns (新しいの, 更新されたか)
	pqinf := p._crt_init(q)

	f2 := f.interpol_poly(g, pqinf)
	return f2, pqinf.pq, f.Equals(f2)
}
