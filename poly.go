package ganrac

import (
	"fmt"
	"io"
	"strings"
)

type Level uint

// poly in K[x_lv,...,x_n]
type Poly struct { // recursive expression
	lv Level
	c  []RObj
}

func NewPoly(lv Level, deg_1 int) *Poly {
	p := new(Poly)
	p.c = make([]RObj, deg_1)
	p.lv = lv
	return p
}

func NewPolyInts(lv Level, coeffs ...int64) *Poly {
	p := NewPoly(lv, len(coeffs))
	for i, c := range coeffs {
		p.c[i] = NewInt(c)
	}
	return p
}

func NewPolyCoef(lv Level, coeffs ...RObj) *Poly {
	p := NewPoly(lv, len(coeffs))
	for i, c := range coeffs {
		p.c[i] = c
	}
	return p
}

func (z *Poly) valid() error {
	if z.c == nil {
		return fmt.Errorf("coefs is null")
	}
	if len(z.c) < 2 {
		return fmt.Errorf("numeric... %v", z)
	}
	if z.c[len(z.c)-1].IsZero() {
		return fmt.Errorf("lc should not be zero... %v", z)
	}
	for _, c := range z.c {
		if c == nil {
			return fmt.Errorf("coef is null")
		}
		err := c.valid()
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *Poly) Equals(x RObj) bool {
	p, ok := x.(*Poly)
	if !ok {
		return false
	}
	if p.lv != z.lv || len(p.c) != len(z.c) {
		return false
	}
	for i := 0; i < len(p.c); i++ {
		if !z.c[i].Equals(p.c[i]) {
			return false
		}
	}
	return true
}

func (z *Poly) Deg(lv Level) int {
	if lv == z.lv {
		return len(z.c) - 1
	} else if lv < z.lv {
		return 0
	}
	m := 0
	for _, c := range z.c {
		p, ok := c.(*Poly)
		if !ok {
			continue
		}
		d := p.Deg(lv)
		if d > m {
			m = d
		}
	}
	return m
}

func (z *Poly) Coef(lv Level, d uint) RObj {
	if lv == z.lv {
		if d >= uint(len(z.c)) {
			return zero
		} else {
			return z.c[d]
		}
	} else if lv < z.lv {
		return zero
	}
	r := NewPoly(z.lv, len(z.c))
	for i, c := range z.c {
		p, ok := c.(*Poly)
		if ok {
			r.c[i] = p.Coef(lv, d)
		} else {
			if d == 0 {
				r.c[i] = c
			} else {
				r.c[i] = zero
			}
		}
	}
	for i := len(z.c) - 1; i > 0; i-- {
		if !r.c[i].IsZero() {
			r.c = r.c[:i+1]
			return r
		}
	}
	return r.c[0]
}

func (z *Poly) Tag() uint {
	return TAG_POLY
}

func (z *Poly) hasVar(lv Level) bool {
	if z.lv > lv {
		return false
	} else if z.lv == lv {
		return true
	}
	for _, c := range z.c {
		cc, ok := c.(*Poly)
		if ok && cc.hasVar(lv) {
			return true
		}
	}
	return false
}

func (z *Poly) Sign() int {
	// sign of leading coeff.
	return z.c[len(z.c)-1].Sign()
}

func (z *Poly) String() string {
	var b strings.Builder
	z.stringFV(&b, varlist)
	return b.String()
}

func (z *Poly) stringFV(b io.Writer, vs []string) {
	for i := len(z.c) - 1; i >= 0; i-- {
		if s := z.c[i].Sign(); s == 0 {
			continue
		} else {
			if z.c[i].IsNumeric() {
				if s > 0 {
					if i != len(z.c)-1 {
						fmt.Fprintf(b, "+")
					}
					if i == 0 || !z.c[i].IsOne() {
						fmt.Fprintf(b, "%v", z.c[i])
						if i != 0 {
							fmt.Fprintf(b, "*")
						}
					}
				} else {
					if i != 0 && z.c[i].IsMinusOne() {
						fmt.Fprintf(b, "-")
					} else {
						fmt.Fprintf(b, "%v", z.c[i])
						if i != 0 {
							fmt.Fprintf(b, "*")
						}
					}
				}
			} else if p, ok := z.c[i].(*Poly); ok {
				if i != len(z.c)-1 {
					fmt.Fprintf(b, "+")
				}
				fmt.Fprintf(b, "(")
				p.stringFV(b, varlist)
				fmt.Fprintf(b, ")*")
			}
			if i > 0 {
				fmt.Fprintf(b, "%s", varlist[z.lv])
				if i > 1 {
					fmt.Fprintf(b, "^%d", i)
				}
			}
		}
	}
}

func (z *Poly) isVar() bool {
	return len(z.c) == 2 && z.c[0].IsZero() && z.c[1].IsOne()
}

func (z *Poly) IsZero() bool {
	return false
}

func (z *Poly) IsOne() bool {
	return false
}

func (z *Poly) IsMinusOne() bool {
	return false
}

func (z *Poly) IsNumeric() bool {
	return false
}

func (z *Poly) Set(x RObj) RObj {
	return z
}

func (z *Poly) Copy() RObj {
	return z.copy()
}

func (z *Poly) copy() *Poly {
	u := NewPoly(z.lv, len(z.c))
	for i, c := range z.c {
		u.c[i] = c
	}
	return u
}

func (z *Poly) Neg() RObj {
	x := z.copy()
	for i := 0; i < len(x.c); i++ {
		x.c[i] = x.c[i].Neg()
	}
	return x
}

func (x *Poly) Add(y RObj) RObj {
	if y.IsNumeric() {
		z := x.copy()
		z.c[0] = z.c[0].Add(y)
		return z
	}
	p, _ := y.(*Poly)
	if p.lv > x.lv {
		z := x.copy()
		z.c[0] = p.Add(z.c[0])
		return z
	} else if p.lv < x.lv {
		z := p.copy()
		z.c[0] = x.Add(z.c[0])
		return z
	} else {
		var dmin int
		var q *Poly
		if len(p.c) < len(x.c) {
			dmin = len(p.c)
			q = x
		} else {
			dmin = len(x.c)
			q = p
		}
		z := NewPoly(p.lv, len(q.c))
		for i := 0; i < dmin; i++ {
			z.c[i] = Add(x.c[i], p.c[i])
		}
		for i := dmin; i < len(q.c); i++ {
			z.c[i] = q.c[i]
		}
		for i := len(q.c) - 1; i > 0; i-- {
			if !z.c[i].IsZero() {
				z.c = z.c[:i+1]
				return z
			}
		}
		return z.c[0]
	}
}

func (z *Poly) Sub(y RObj) RObj {
	// @TODO とりまサボり.
	yn := y.Neg()
	return z.Add(yn)
}

func (x *Poly) Mul(yy RObj) RObj {
	// @TODO とりあえず素朴版 -> Karatsuba へ
	if yy.IsNumeric() {
		if yy.IsZero() {
			return yy
		}
		z := NewPoly(x.lv, len(x.c))
		for i := 0; i < len(x.c); i++ {
			z.c[i] = x.c[i].Mul(yy)
		}
		return z
	}
	y, _ := yy.(*Poly)
	if y.lv > x.lv {
		z := NewPoly(x.lv, len(x.c))
		for i := 0; i < len(x.c); i++ {
			z.c[i] = y.Mul(x.c[i])
		}
		return z
	} else if y.lv < x.lv {
		z := NewPoly(y.lv, len(y.c))
		for i := 0; i < len(y.c); i++ {
			z.c[i] = x.Mul(y.c[i])
		}
		return z
	}
	z := NewPoly(x.lv, len(y.c)+len(x.c)-1)
	for i := 0; i < len(z.c); i++ {
		z.c[i] = zero
	}
	for i := 0; i < len(x.c); i++ {
		if x.c[i].IsZero() {
			continue
		}
		xiyy := y.Mul(x.c[i])
		xiy, _ := xiyy.(*Poly)
		for j := len(xiy.c) - 1; j >= 0; j-- {
			z.c[i+j] = Add(z.c[i+j], xiy.c[j])
		}
	}

	return z
}

func (x *Poly) Div(y NObj) RObj {
	z := NewPoly(x.lv, len(x.c))
	for i, c := range x.c {
		z.c[i] = c.Div(y)
	}
	return z
}

func (x *Poly) Pow(y *Int) RObj {
	// return x^y
	// int版と同じ手法. 通常 x^m 以外では使わないから放置
	// @TODO 2項定理使ったほうが効率的?
	if y.Sign() < 0 {
		return nil // unsupported.
	}
	m := y.n.BitLen() - 1
	if m < 0 {
		return NewInt(1)
	}

	t := x
	var z *Poly
	for i := 0; i < m; i++ {
		if y.n.Bit(i) != 0 {
			if z == nil {
				z = t
			} else {
				zz := z.Mul(t)
				z, _ = zz.(*Poly)
			}
		}

		tt := t.Mul(t)
		t, _ = tt.(*Poly)
	}
	if z == nil {
		return t
	}

	return z.Mul(t)
}

func (z *Poly) Subst(xs []RObj, lvs []Level, idx int) RObj {
	// lvs: sorted

	for ; idx < len(xs) && z.lv > lvs[idx]; idx++ {
	}
	if idx == len(xs) {
		return z
	}
	if z.lv < lvs[idx] {
		p := NewPoly(z.lv, len(z.c))
		for i := 0; i < len(z.c); i++ {
			p.c[i] = z.c[i].Subst(xs, lvs, idx)
		}
		for i := len(z.c) - 1; i > 0; i-- {
			if !p.c[i].IsZero() {
				p.c = p.c[:i+1]
				return p
			}
		}
		return p.c[0]
	}
	x := xs[idx]
	p := z.c[len(z.c)-1].Subst(xs, lvs, idx+1)
	for i := len(z.c) - 2; i >= 0; i-- {
		p = Add(Mul(p, x), z.c[i].Subst(xs, lvs, idx+1))
	}
	return p
}

func (z *Poly) RootBound() (NObj, error) {
	for _, c := range z.c {
		if _, ok := c.(NObj); !ok {
			return nil, fmt.Errorf("supportedx only for univariate polynomial")
		}
	}

	var m NObj
	m = z.c[0].(NObj)
	for i := 1; i < len(z.c)-1; i++ {
		if m.CmpAbs(z.c[i].(NObj)) < 0 {
			m = z.c[i].(NObj)
		}
	}
	m = m.Div(z.c[len(z.c)-1].(NObj)).(NObj)
	return m.Abs().AddInt(1), nil
}
