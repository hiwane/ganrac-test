package ganrac

import (
	"fmt"
	"math/big"
)

var mone *Int = NewInt(-1) // -1
var zero *Int = newInt()   // 0
var one *Int = NewInt(1)
var two *Int = NewInt(2)
var four *Int = NewInt(4)

type Int struct {
	Number
	n *big.Int
}

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

func NewIntZ(z *big.Int) *Int {
	v := new(Int)
	v.n = z
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
func (x *Int) numTag() uint {
	return NTAG_INT
}

func (x *Int) Equals(y interface{}) bool {
	c, ok := y.(*Int)
	return ok && x.n.Cmp(c.n) == 0
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
	case *BinInt:
		return yi.AddInt(x)
	case *Interval:
		if x.IsZero() {
			return zero.toIntv(1)
		}
	}
	fmt.Printf("add: x=%v, y=%v\n", x, y)
	panic("not implememted")
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
	fmt.Printf("sub: x=%v, y=%v\n", x, y)
	panic("not implememted")
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
	case *BinInt:
		return yi.MulInt(x)
	case *Poly:
		fmt.Printf("mul: x=%v, y=%v\n", x, y)
		panic("poly!")
	}
	fmt.Printf("mul: x=%v, y=%v\n", x, y)
	panic("stop")
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
	panic("not implemented") // @TODO
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

func (x *Int) Format(s fmt.State, format rune) {
	switch format {
	case 'e', 'E', 'f', 'F', 'g', 'G':
		f := new(big.Float)
		if w, ok := s.Precision(); ok {
			f.SetPrec(uint(w) + 10)
		}
		f.SetInt(x.n)
		f.Format(s, format)
	case FORMAT_DUMP, FORMAT_TEX, FORMAT_QEPCAD:
		x.n.Format(s, 'd')
	case FORMAT_SRC:
		if x.n.IsInt64() {
			if x.IsZero() {
				fmt.Fprintf(s, "zero")
			} else if x.IsOne() {
				fmt.Fprintf(s, "one")
			} else if x.IsMinusOne() {
				fmt.Fprintf(s, "mone")
			} else if x.Cmp(two) == 0 {
				fmt.Fprintf(s, "two")
			} else {
				fmt.Fprintf(s, "NewInt(%v)", x.n)
			}
		} else {
			fmt.Fprintf(s, "ParseInt(\"%v\", 10)", x.n)
		}
	default:
		x.n.Format(s, format)
	}
}

func (x *Int) Sign() int {
	return x.n.Sign()
}

func (x *Int) IsZero() bool {
	return x.n.Sign() == 0
}

func (x *Int) IsOne() bool {
	return x.n.Sign() > 0 && x.n.BitLen() == 1
}

func (x *Int) IsMinusOne() bool {
	return x.n.Sign() < 0 && x.n.BitLen() == 1
}

func (z *Int) Subst(x RObj, lv Level) RObj {
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
	case *BinInt:
		if x.m > 0 {
			xr := new(big.Int)
			xr.Lsh(x.n, uint(x.m))
			return z.n.Cmp(xr)
		} else if x.m == 0 {
			return z.n.Cmp(x.n)
		} else {
			zr := new(big.Int)
			zr.Lsh(z.n, uint(-x.m))
			return zr.Cmp(x.n)
		}

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

// func (z *Int) ToInt(n int) *Int {
// 	return z
// }

func (z *Int) Float() float64 {
	f := new(big.Float)
	f.SetInt(z.n)
	ff, _ := f.Float64()
	return ff
}

func (x *Int) Gcd(y *Int) *Int {
	if x.Sign() == 0 || y.Sign() == 0 {
		return zero
	}
	s := x.n.CmpAbs(y.n)
	if s == 0 {
		if x.Sign() > 0 {
			return x
		}
		if y.Sign() > 0 {
			return y
		}
		return x.Abs().(*Int)
	}
	var r1, r2 *big.Int
	r1 = new(big.Int)
	r2 = new(big.Int)
	if s > 0 {
		r1.Abs(x.n)
		r2.Abs(y.n)
	} else {
		r1.Abs(y.n)
		r2.Abs(x.n)
	}
	for {
		r := new(big.Int)
		r.Mod(r1, r2)
		if r.Sign() == 0 {
			break
		}
		r1 = r2
		r2 = r
	}
	v := new(Int)
	v.n = r2
	return v
}

func (x *Int) GcdEx(y *Int) (*Int, *Int, *Int) {
	// extended Euclid algorithm
	// returns (gcd, a, b) where gcd = a*x + b*y
	r1 := x.n
	r2 := y.n
	a1 := big.NewInt(1)
	a2 := big.NewInt(0)
	b1 := a2
	b2 := a1

	for {
		q := new(big.Int)
		r := new(big.Int)
		q.QuoRem(r1, r2, r)
		if r.Sign() == 0 {
			g := newInt()
			g.n.Set(r2)
			a := newInt()
			a.n.Set(a2)
			b := newInt()
			b.n.Set(b2)
			return g, a, b
		}

		r1, r2 = r2, r
		a := new(big.Int)
		a.Sub(a1, a.Mul(a2, q))
		b := new(big.Int)
		b.Sub(b1, b.Mul(b2, q))

		a1, a2 = a2, a
		b1, b2 = b2, b
	}
}

func (x *Int) subst_poly(p *Poly, lv Level) RObj {
	return p.Subst(x, lv)
}

func (x *Int) mul_2exp(m uint) RObj {
	v := newInt()
	v.n.Lsh(x.n, m)
	return v
}

func (x *Int) toIntv(prec uint) RObj {
	z := newInterval(prec)
	z.inf.SetInt(x.n)
	z.sup.SetInt(x.n)
	return z
}

func (x *Int) Bit(i int) uint {
	return x.n.Bit(i)
}
