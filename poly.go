package ganrac

import (
	"fmt"
	"io"
	"strings"
)

type Poly struct { // recursive expression
	lv uint
	c  []Coef
}

func NewPoly(lv uint) *Poly {
	p := new(Poly)
	p.c = make([]Coef, 0)
	p.lv = lv
	return p
}

func NewPolyInts(lv uint, coeffs ...int64) *Poly {
	p := NewPoly(lv)
	for _, c := range coeffs {
		p.c = append(p.c, NewInt(c))
	}
	return p
}

func (z *Poly) Equals(x Coef) bool {
	if x.IsNumeric() {
		return false
	}
	p, _ := x.(*Poly)
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

func (z *Poly) Tag() uint {
	return TAG_POLY
}

func (z *Poly) Sign() int {
	return 1
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

func (z *Poly) New() Coef {
	v := new(Poly)
	v.lv = z.lv
	return v
}

func (z *Poly) Set(x Coef) Coef {
	return z
}

func (z *Poly) Neg() Coef {
	return z
}

func (x *Poly) Add(y Coef) Coef {
	if y.IsNumeric() {
		z := NewPoly(x.lv)
		copy(z.c, x.c)
		z.c[0] = z.c[0].Add(y)
		return z
	}
	p, _ := y.(*Poly)
	if p.lv > x.lv {
		z := NewPoly(x.lv)
		copy(z.c, x.c)
		z.c[0] = p.Add(z.c[0])
		return z
	} else if p.lv < x.lv {
		z := NewPoly(p.lv)
		copy(z.c, p.c)
		z.c[0] = x.Add(z.c[0])
		return z
	}

	// if x.IsNumeric() {
	// 	if y.IsNumeric() {
	// 	} else {
	// 	}
	// }
	//
	// if x.lv == y.lv {
	// 	z.lv = 0
	// 	var n, m int
	// 	var u []Coef
	// 	if len(x.c) < len(y.c) {
	// 		m = len(y.c)
	// 		n = len(x.c)
	// 		u = y.c
	// 	} else {
	// 		m = len(x.c)
	// 		n = len(y.c)
	// 		u = x.c
	// 	}
	// 	z.c = make([]Coef, m)
	// 	for i := 0; i <= n; i++ {
	// 		z.c[i].Add(x.c[i], y.c[i])
	// 	}
	// 	for i := n + 1; i < m; i++ {
	// 		z.c[i].Set(u[i])
	// 	}
	// }
	return x
}

func (z *Poly) Sub(y Coef) Coef {
	// とりまサボり.
	yn := y.Neg()
	return z.Add(yn)
}

func (z *Poly) Mul(y Coef) Coef {
	return z
}
