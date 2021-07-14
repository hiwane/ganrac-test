package ganrac

import (
	"fmt"
	"math/big"
)

type BinInt struct {
	// express n * 2^m
	n *big.Int
	m int
	Number
}

func newBinInt() *BinInt {
	v := new(BinInt)
	v.n = new(big.Int)
	return v
}

func newBinIntInt64(n int64, m int) *BinInt {
	v := new(BinInt)
	v.n = big.NewInt(n)
	v.m = m
	return v
}

func NewBinInt(n *big.Int, m int) *BinInt {
	v := new(BinInt)
	v.n = n
	v.m = m
	return v
}

func (x *BinInt) Equals(yy interface{}) bool {
	y, ok := yy.(NObj)
	if !ok {
		return false
	}
	return x.Cmp(y) == 0
}

func (x *BinInt) String() string {
	return fmt.Sprintf("%v", x)
}

func (x *BinInt) Format(s fmt.State, format rune) {
	switch format {
	case 'e', 'E', 'f', 'F', 'g', 'G':
		f := new(big.Float)
		if w, ok := s.Precision(); ok {
			f.SetPrec(uint(w) + 10)
		}
		f.SetInt(x.n)
		f.SetMantExp(f, x.m)
		f.Format(s, format)
		return
	case FORMAT_DUMP: // dump
		fmt.Fprintf(s, "(bin %v %d)", x.n, x.m)
		return
	case FORMAT_TEX:
		if x.m > 0 {
			if x.m == 1 {
				fmt.Fprintf(s, "%v \\cdot 2", x.n)
			} else {
				fmt.Fprintf(s, "%v \\cdot 2^{%d}", x.n, x.m)
			}
		} else if x.m < 0 {
			if x.m == -1 {
				fmt.Fprintf(s, "\\frac{%v}{2}", x.n)
			} else {
				fmt.Fprintf(s, "\\frac{%v}{2^{%d}}", x.n, x.m)
			}
		} else {
			fmt.Fprintf(s, "%v", x.n)
		}
		return
	case FORMAT_SRC:
		if x.n.IsInt64() {
			fmt.Fprintf(s, "newBinIntInt64(%v, %d)", x.n, x.m)
		} else {
			fmt.Fprintf(s, "NewBinInt(ParseInt(\"%v\", 10).n, %d)", x.n, x.m)
		}
	default:
		x.n.Format(s, format)
	}

	if x.m > 0 {
		if x.m == 1 {
			fmt.Fprintf(s, "*2")
		} else {
			fmt.Fprintf(s, "*2^%d", x.m)
		}
	} else if x.m < 0 {
		if x.m == -1 {
			fmt.Fprintf(s, "/2")
		} else {
			fmt.Fprintf(s, "/2^%d", -x.m)
		}
	}
}

func (x *BinInt) AddInt(y *Int) RObj {
	if x.m < 0 {
		z := newBinInt()
		z.m = x.m
		z.n.Lsh(y.n, uint(-x.m))
		z.n.Add(z.n, x.n)
		return z
	} else {
		z := newInt()
		z.n.Lsh(x.n, uint(x.m))
		z.n.Add(z.n, y.n)
		return z
	}
}

func (x *BinInt) Add(yy RObj) RObj {
	switch y := yy.(type) {
	case *BinInt:
		z := newBinInt()
		if y.m == x.m {
			z.n.Add(y.n, x.n)
			z.m = x.m
		} else if y.m < x.m {
			z.n.Lsh(x.n, uint(x.m-y.m))
			z.n.Add(z.n, y.n)
			z.m = y.m
		} else {
			z.n.Lsh(y.n, uint(y.m-x.m))
			z.n.Add(z.n, x.n)
			z.m = x.m
		}
		return z
	case *Int:
		z := x.AddInt(y)
		return z
	case *Rat:
		z := x.ToIntRat()
		return z.Add(y)
	}
	panic("unknown")
}

func (x *BinInt) Sub(yy RObj) RObj {
	y := yy.Neg()
	return x.Add(y)
}

func (x *BinInt) MulInt(y *Int) RObj {
	z := newBinInt()
	z.m = x.m
	z.n.Mul(x.n, y.n)
	return z

}

func (x *BinInt) Mul(yy RObj) RObj {
	switch y := yy.(type) {
	case *BinInt:
		z := newBinInt()
		z.m = x.m + y.m
		z.n.Mul(x.n, y.n)
		return z
	case *Int:
		z := x.MulInt(y)
		return z
	case *Rat:
		z := x.ToIntRat()
		return z.Mul(y)
	}
	panic("not implemented")
}

func (x *BinInt) Div(yy NObj) RObj {
	z := x.ToIntRat()
	return z.Div(yy)
}

func (x *BinInt) Pow(r *Int) RObj {
	if r.Sign() >= 0 {
		zz := NewIntZ(x.n)
		xr := zz.Pow(r).(*Int)
		z := newBinInt()
		z.n = xr.n
		z.m = x.m * int(r.n.Int64())
		return z
	}
	panic("not implemented")
}

func (z *BinInt) Subst(x RObj, lv Level) RObj {
	return z
}

func (x *BinInt) Neg() RObj {
	z := newBinInt()
	z.n.Neg(x.n)
	z.m = x.m
	return z
}

func (x *BinInt) Sign() int {
	return x.n.Sign()
}

func (x *BinInt) isAbsOne() bool {
	// returns |x| == 1
	if x.n.BitLen() != -x.m+1 {
		return false
	}
	for i := -x.m - 1; i >= 0; i-- {
		if x.n.Bit(i) != 0 {
			return false
		}
	}
	return true
}

func (x *BinInt) IsZero() bool {
	return x.n.Sign() == 0
}

func (x *BinInt) IsOne() bool {
	return x.n.Sign() > 0 && x.isAbsOne()
}

func (x *BinInt) IsMinusOne() bool {
	return x.n.Sign() < 0 && x.isAbsOne()
}

func (z *BinInt) ToIntRat() NObj {
	if z.m < 0 {
		den := big.NewInt(1)
		den.Lsh(den, uint(-z.m))

		x := newRat()
		x.n.SetFrac(z.n, den)
		return x
	} else {
		x := newInt()
		x.n.Lsh(z.n, uint(z.m))
		return x
	}
}

func (z *BinInt) Float() float64 {
	f := new(big.Float)
	z.setToBigFloat(f)
	ff, _ := f.Float64()
	return ff
}

func (z *BinInt) Cmp(xx NObj) int {
	sl := z.Sign()
	sh := xx.Sign()
	if sl != sh {
		return sl - sh
	}

	switch x := xx.(type) {
	case *BinInt:
		if x.m == z.m {
			return z.n.Cmp(x.n)
		}
		y := new(big.Int)
		if x.m < z.m {
			y.Lsh(z.n, uint(z.m-x.m))
			return y.Cmp(x.n)
		} else {
			y.Lsh(x.n, uint(x.m-z.m))
			return z.n.Cmp(y)

		}
	case *Int:
		if z.m >= 0 {
			zz := z.ToIntRat()
			return zz.Cmp(x)
		} else {
			x2 := newInt()
			x2.n.Lsh(x.n, uint(-z.m))

			zz := newInt()
			zz.n.Set(z.n)
			return zz.Cmp(x2)
		}
	case *Rat:
		zz := z.ToIntRat()
		return zz.Cmp(x)
	}
	fmt.Printf("z=%v\n", z)
	fmt.Printf("x=%v\n", xx)
	panic("unknown")
}

func (z *BinInt) CmpAbs(xx NObj) int {
	sl := z.Sign()
	sh := xx.Sign()
	if sl == sh {
		return z.Cmp(xx) * sl
	}

	x := xx.Neg().(NObj)
	return z.Cmp(x) * sl
}

func (z *BinInt) Abs() NObj {
	if z.Sign() >= 0 {
		return z
	} else {
		return z.Neg().(NObj)
	}
}

// binary interval の右端
func (x *BinInt) upperBound() *BinInt {
	z := newBinInt()
	z.n.Add(x.n, one.n)
	z.m = x.m
	return z
}

// binary interval の中点
func (x *BinInt) midBinIntv() *BinInt {
	z := newBinInt()
	z.n.Lsh(x.n, 1)
	z.n.Add(z.n, one.n)
	z.m = x.m - 1
	return z
}

// binary interval の区間幅を半分にする.
// n を 2 倍して，m を減らす.
func (x *BinInt) halveIntv() *BinInt {
	z := newBinInt()
	z.n.Lsh(x.n, 1)
	z.m = x.m - 1
	return z
}

func (x *BinInt) subst_poly(p *Poly, lv Level) RObj {
	if x.m > 0 {
		xx := newInt()
		xx.n.Lsh(x.n, uint(x.m))
		return p.Subst(xx, lv)
	} else if x.m < 0 {
		deg := p.Deg(lv)
		return p.subst_num_2exp(NewIntZ(x.n), uint(-x.m), lv, deg)
	} else {
		xx := new(Int)
		xx.n = x.n
		return p.Subst(xx, lv)
	}
}

func (x *BinInt) mul_2exp(m uint) RObj {
	v := newBinInt()
	v.n = x.n
	v.m = x.m + int(m)
	return v
}

func (x *BinInt) setToBigFloat(y *big.Float) {
	y.SetInt(x.n)
	y.SetMantExp(y, x.m)
}

func (x *BinInt) toIntv(prec uint) RObj {
	z := newInterval(prec)
	x.setToBigFloat(z.inf)
	x.setToBigFloat(z.sup)
	return z
}
