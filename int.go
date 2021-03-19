package ganrac

import (
	"math/big"
)

type Int struct {
	n *big.Int
}

func newInt() *Int {
	v := new(Int)
	v.n = new(big.Int)
	return v
}

func NewInt(n int64) *Int {
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
	return v
}

func (x *Int) Equals(y Coef) bool {
	if y.Tag() != TAG_INT {
		return false
	}
	c, _ := y.(*Int)
	return x.n.Cmp(c.n) == 0
}

func (x *Int) Tag() uint {
	return TAG_INT
}

func (x *Int) Add(y Coef) Coef {
	yi, _ := y.(*Int)
	z := newInt()
	z.n.Add(x.n, yi.n)
	return z
}

func (x *Int) Sub(y Coef) Coef {
	yi, _ := y.(*Int)
	z := newInt()
	z.n.Sub(x.n, yi.n)
	return z
}

func (x *Int) Mul(y Coef) Coef {
	yi, _ := y.(*Int)
	z := newInt()
	z.n.Mul(x.n, yi.n)
	return z
}

func (x *Int) Neg() Coef {
	z := newInt()
	z.n.Neg(x.n)
	return z
}

func (x *Int) Set(y Coef) Coef {
	yi, _ := y.(*Int)
	x.n.Set(yi.n)
	return x
}

func (x *Int) New() Coef {
	v := new(Int)
	v.n = new(big.Int)
	return v
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
	if !x.n.IsInt64() {
		return false
	}
	m := x.n.Int64()
	return m == 1
}

func (x *Int) IsMinusOne() bool {
	if !x.n.IsInt64() {
		return false
	}
	m := x.n.Int64()
	return m == -1
}

func (x *Int) IsNumeric() bool {
	return true
}
