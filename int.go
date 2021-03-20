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

func (x *Int) Equals(y RObj) bool {
	if y.Tag() != TAG_INT {
		return false
	}
	c, _ := y.(*Int)
	return x.n.Cmp(c.n) == 0
}

func (x *Int) Tag() uint {
	return TAG_INT
}

func (x *Int) Add(y RObj) RObj {
	yi, _ := y.(*Int)
	z := newInt()
	z.n.Add(x.n, yi.n)
	return z
}

func (x *Int) Sub(y RObj) RObj {
	yi, _ := y.(*Int)
	z := newInt()
	z.n.Sub(x.n, yi.n)
	return z
}

func (x *Int) Mul(y RObj) RObj {
	yi, _ := y.(*Int)
	z := newInt()
	z.n.Mul(x.n, yi.n)
	return z
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

func (x *Int) Pow(y *Int) RObj {
	// return x^y
	if y.Sign() < 0 {
		return nil // unsupported.
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

func (x *Int) New() RObj {
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
