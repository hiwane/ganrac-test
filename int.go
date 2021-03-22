package ganrac

import (
	"fmt"
	"math/big"
)

type Int struct {
	Number
	n *big.Int
}

var zero *Int = newInt()
var one *Int = NewInt(1)

func newInt() *Int {
	v := new(Int)
	v.n = new(big.Int)
	return v
}

func NewInt(n int64) *Int {
	if n == 0 {
		return zero
	}
	v := new(Int)
	v.n = big.NewInt(n)
	return v
}

func ParseInt(s string, base int) *Int {
	v := newInt()
	_, ok := v.n.SetString(s, base)
	if ok {
		return v
	} else {
		return nil
	}
}

func (x *Int) Equals(y RObj) bool {
	c, ok := y.(*Int)
	return ok && x.n.Cmp(c.n) == 0
}

func (x *Int) AddInt(n int64) NObj {
	z := newInt()
	z.n.SetInt64(n)
	z.n.Add(z.n, x.n)
	return z
}

func (x *Int) Add(y RObj) RObj {
	switch yi := y.(type) {
	case *Int:
		z := newInt()
		z.n.Add(x.n, yi.n)
		return z
	case *Rat:
		xr := new(big.Rat)
		xr.SetInt(x.n)
		z := newRat()
		z.n.Add(xr, yi.n)
		return z
	}
	return nil
}

func (x *Int) Sub(y RObj) RObj {
	switch yi := y.(type) {
	case *Int:
		z := newInt()
		z.n.Sub(x.n, yi.n)
		return z
	case *Rat:
		xr := new(big.Rat)
		xr.SetInt(x.n)
		z := newRat()
		z.n.Sub(xr, yi.n)
		return z
	}
	return nil
}

func (x *Int) Mul(y RObj) RObj {
	switch yi := y.(type) {
	case *Int:
		z := newInt()
		z.n.Mul(x.n, yi.n)
		return z
	case *Rat:
		xr := new(big.Rat)
		xr.SetInt(x.n)
		z := newRat()
		z.n.Mul(xr, yi.n)
		return z.normal()
	}
	return nil
}

func (x *Int) Neg() RObj {
	z := newInt()
	z.n.Neg(x.n)
	return z
}

func (x *Int) Set(y RObj) RObj {
	yi, _ := y.(*Int)
	x.n.Set(yi.n)
	return x
}

func (x *Int) Div(yy NObj) RObj {
	switch y := yy.(type) {
	case *Int:
		if y.n.Cmp(one.n) == 0 {
			return x
		}
		z := newRat()
		z.n.SetFrac(x.n, y.n)
		return z.normal()
	case *Rat:
		yr := new(big.Rat)
		yr.Inv(y.n)
		xr := new(big.Rat)
		xr.SetInt(x.n)

		z := newRat()
		z.n.Mul(xr, yr)
		return z.normal()
	}
	return nil // @TODO
}

func (x *Int) Pow(y *Int) RObj {
	// return x^y
	if y.Sign() < 0 {
		z := newRat()
		z.n.SetFrac(one.n, x.n)
		return z.Pow(y.Neg().(*Int))
	}
	m := y.n.BitLen() - 1
	if m < 0 {
		return NewInt(1)
	}

	t := new(big.Int)
	t.Set(x.n)

	z := new(big.Int)
	z.SetInt64(1)
	for i := 0; i < m; i++ {
		if y.n.Bit(i) != 0 {
			z2 := new(big.Int)
			z2.Mul(z, t)
			z = z2
		}

		t2 := new(big.Int)
		t2.Mul(t, t)
		t = t2
	}

	zz := newInt()
	zz.n.Mul(z, t)
	return zz
}

func (x *Int) IsInt64() bool {
	return x.n.IsInt64()
}
func (x *Int) Int64() int64 {
	return x.n.Int64()
}

func (x *Int) String() string {
	return x.n.String()
}

func (x *Int) Sign() int {
	return x.n.Sign()
}

func (x *Int) IsZero() bool {
	return x.n.Sign() == 0
}

func (x *Int) IsOne() bool {
	return x.n.Cmp(one.n) == 0
}

func (x *Int) IsMinusOne() bool {
	if !x.n.IsInt64() {
		return false
	}
	m := x.n.Int64()
	return m == -1
}

func (z *Int) Subst(x []RObj, lv []Level, idx int) RObj {
	return z
}

func (z *Int) Cmp(xx NObj) int {
	switch x := xx.(type) {
	case *Int:
		return z.n.Cmp(x.n)
	case *Rat:
		zr := new(big.Int)
		zr.Mul(z.n, x.n.Denom())
		return zr.Cmp(x.n.Num())
	}
	panic("unknown")
}

func (z *Int) CmpAbs(xx NObj) int {
	switch x := xx.(type) {
	case *Int:
		return z.n.CmpAbs(x.n)
	case *Rat:
		zr := new(big.Int)
		zr.Mul(z.n, x.n.Denom())
		return zr.CmpAbs(x.n.Num())
	}
	panic("unknown")
}

func (z *Int) Abs() NObj {
	if z.Sign() >= 0 {
		return z
	} else {
		return z.Neg().(NObj)
	}
}

func (z *Int) valid() error {
	if zero.n.Sign() != 0 {
		return fmt.Errorf("zero is broken: %v", zero)
	}
	if one.String() != "1" {
		return fmt.Errorf("one is broken: %v", one)
	}
	return nil
}

func (z *Int) ToInt(n int) *Int {
	return z
}
