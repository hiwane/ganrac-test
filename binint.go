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

func NewBinInt(n int64, m int) *BinInt {
	v := new(BinInt)
	v.n = big.NewInt(n)
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
	if x.m == 0 {
		return fmt.Sprintf("%v", x.n)
	} else if x.m > 0 {
		if x.m == 1 {
			return fmt.Sprintf("%v*2", x.n)
		} else {
			return fmt.Sprintf("%v*2^%d", x.n, x.m)
		}
	} else {
		if x.m == -1 {
			return fmt.Sprintf("%v/2", x.n)
		} else {
			return fmt.Sprintf("%v/2^%d", x.n, -x.m)
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

func (x *BinInt) Mul2Exp(m uint) NObj {
	z := new(BinInt)
	z.n = x.n
	z.m = x.m + int(m)
	return z
}

func (x *BinInt) Div(yy NObj) RObj {
	panic("not implemented")
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

func (z *BinInt) Subst(x []RObj, lv []Level, idx int) RObj {
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
	f.SetInt(z.n)
	f = f.SetMantExp(f, z.m)
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