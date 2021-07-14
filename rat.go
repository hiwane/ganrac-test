package ganrac

import (
	"fmt"
	"math/big"
)

var brat_one = big.NewRat(1, 1)
var brat_mone = big.NewRat(-1, 1)

type Rat struct {
	Number
	n *big.Rat
}

func newRat() *Rat {
	v := new(Rat)
	v.n = new(big.Rat)
	return v
}

func NewRatInt64(num, den int64) *Rat {
	v := new(Rat)
	v.n = big.NewRat(num, den)
	return v
}

func NewRatFrac(num, den *Int) *Rat {
	v := newRat()
	v.n.SetFrac(num.n, den.n)
	return v
}

func (x *Rat) numtag() uint {
	return NTAG_RAT
}

func (x *Rat) Equals(y interface{}) bool {
	c, ok := y.(*Rat)
	return ok && x.n.Cmp(c.n) == 0
}

func (z *Rat) normal() RObj {
	if !z.n.IsInt() {
		return z
	}
	zi := newInt()
	zi.n.Set(z.n.Num())
	return zi
}

func (x *Rat) Add(yy RObj) RObj {
	switch y := yy.(type) {
	case *Int:
		yr := new(big.Rat)
		yr.SetInt(y.n)
		z := newRat()
		z.n.Add(x.n, yr)
		return z
	case *Rat:
		z := newRat()
		z.n.Add(x.n, y.n)
		return z.normal()
	case *BinInt:
		z := y.ToIntRat()
		return x.Add(z)
	}
	panic("stop")
}

func (x *Rat) Sub(yy RObj) RObj {
	switch y := yy.(type) {
	case *Int:
		yr := new(big.Rat)
		yr.SetInt(y.n)
		z := newRat()
		z.n.Sub(x.n, yr)
		return z
	case *Rat:
		z := newRat()
		z.n.Sub(x.n, y.n)
		return z.normal()
	}
	panic("stop")
}

func (x *Rat) Mul(yy RObj) RObj {
	switch y := yy.(type) {
	case *Int:
		yr := new(big.Rat)
		yr.SetInt(y.n)
		z := newRat()
		z.n.Mul(x.n, yr)
		return z.normal()
	case *Rat:
		z := newRat()
		z.n.Mul(x.n, y.n)
		return z.normal()
	}
	panic("stop")
}

func (x *Rat) Div(yy NObj) RObj {
	switch y := yy.(type) {
	case *Int:
		yr := new(big.Rat)
		yr.SetFrac(one.n, y.n)

		z := newRat()
		z.n.Mul(x.n, yr)
		return z
	case *Rat:
		yr := new(big.Rat)
		yr.Inv(y.n)

		z := newRat()
		z.n.Mul(x.n, yr)
		return z.normal()
	}

	panic("stop")
}

func (x *Rat) Pow(y *Int) RObj {
	den := newInt()
	den.n.Set(x.n.Denom())

	num := newInt()
	num.n.Set(x.n.Num())

	if y.Sign() > 0 {
		deni := den.Pow(y).(*Int)
		numi := num.Pow(y).(*Int)
		return numi.Div(deni)
	} else if y.Sign() < 0 {
		yi := y.Neg().(*Int)
		deni := den.Pow(yi).(*Int)
		numi := num.Pow(yi).(*Int)
		return deni.Div(numi)
	} else {
		return one
	}
}

func (z *Rat) Subst(x RObj, lv Level) RObj {
	return z
}

func (x *Rat) Neg() RObj {
	z := newRat()
	z.n.Neg(x.n)
	return z
}

func (x *Rat) String() string {
	return x.n.String()
}

func (x *Rat) Format(s fmt.State, format rune) {
	switch format {
	case 'e', 'E', 'f', 'F', 'g', 'G':
		f := new(big.Float)
		if w, ok := s.Precision(); ok {
			f.SetPrec(uint(w) + 10)
		}
		f.SetRat(x.n)
		f.Format(s, format)
	case FORMAT_TEX:
		fmt.Fprintf(s, "\\frac{%v}{%v}", x.n.Num(), x.n.Denom())
	case FORMAT_DUMP:
		x.n.Num().Format(s, 'd')
		fmt.Fprintf(s, "/")
		x.n.Denom().Format(s, 'd')
	case FORMAT_SRC:
		if x.n.Num().IsInt64() && x.n.Denom().IsInt64() {
			fmt.Fprintf(s, "NewRatInt64(%v, %v)", x.n.Num(), x.n.Denom())
		} else {
			fmt.Fprintf(s, "NewRatInt64(ParseInt(\"%v\", 10), ParseInt(\"%v\", 10))", x.n.Num(), x.n.Denom())
		}
	default:
		x.n.Num().Format(s, format)
		fmt.Fprintf(s, "/")
		x.n.Denom().Format(s, format)
	}
}

func (x *Rat) Sign() int {
	return x.n.Sign()
}

func (x *Rat) IsZero() bool {
	return x.n.Sign() == 0
}

func (x *Rat) IsOne() bool {
	return x.n.Cmp(brat_one) == 0
}

func (x *Rat) IsMinusOne() bool {
	return x.n.Cmp(brat_mone) == 0
}

func (z *Rat) Cmp(xx NObj) int {
	switch x := xx.(type) {
	case *Int:
		xr := new(big.Int)
		xr.Mul(x.n, z.n.Denom())
		return z.n.Num().Cmp(xr)
	case *Rat:
		return z.n.Cmp(x.n)
	case *BinInt:
		xr := x.ToIntRat()
		return z.Cmp(xr)
	}
	panic(fmt.Sprintf("unknown: z=%v, x=%v", z, xx))
}

func (z *Rat) CmpAbs(xx NObj) int {
	switch x := xx.(type) {
	case *Int:
		xr := new(big.Int)
		xr.Mul(x.n, z.n.Denom())
		return z.n.Num().CmpAbs(xr)
	case *Rat:
		s1 := z.Sign()
		s2 := x.Sign()
		if s1 < s2 {
			return -1
		} else if s1 > s2 {
			return +1
		} else if s1 == 0 {
			return 0
		}
		// same sign
		return s1 * z.n.Cmp(x.n)
	}
	panic("unknown")
}

func (z *Rat) Abs() NObj {
	if z.Sign() >= 0 {
		return z
	} else {
		return z.Neg().(NObj)
	}
}

func (z *Rat) valid() error {
	if z.n.IsInt() {
		return fmt.Errorf("den = 1. rat=%v", z)
	}
	if brat_one.Cmp(big.NewRat(1, 1)) != 0 {
		return fmt.Errorf("brat_one is broken: %v", zero)
	}
	if brat_mone.Cmp(big.NewRat(-1, 1)) != 0 {
		return fmt.Errorf("brat_mone is broken: %v", zero)
	}
	return nil
}

// func (z *Rat) ToInt(n int) *Int {
// 	v := newInt()
// 	v.n.Div(z.n.Num(), z.n.Denom())
// 	return v
// }

func (z *Rat) Float() float64 {
	f, _ := z.n.Float64()
	return f
}

func (x *Rat) subst_poly(p *Poly, lv Level) RObj {
	deg := p.Deg(lv)
	if deg == 0 {
		return p
	}

	dd := new(Int)
	dd.n = x.n.Denom()

	dens := make([]RObj, deg+1)
	dens[0] = one
	dens[1] = dd

	for i := 1; i < deg; i++ {
		dens[i+1] = dens[i].Mul(dd)
	}

	num := new(Int)
	num.n = x.n.Num()

	return p.subst_frac(num, dens, lv)
}

func (x *Rat) mul_2exp(m uint) RObj {
	v := newRat()
	num := new(big.Int)
	den := x.n.Denom()

	num.Lsh(x.n.Num(), m)
	v.n.SetFrac(num, den)
	return v
}

func (x *Rat) toIntv(prec uint) RObj {
	z := newInterval(prec)
	z.inf.SetRat(x.n)
	z.sup.SetRat(x.n)
	return z
}
