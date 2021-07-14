package ganrac

// I/F は big.Int に揃える
import (
	"fmt"
	"math/big"
)

type Interval struct {
	inf *big.Float // lower value
	sup *big.Float // upper value
}

func (z *Interval) String() string {
	return fmt.Sprintf("[%f,%f]", z.inf, z.sup)
}

func (x *Interval) Format(s fmt.State, format rune) {
	if x == nil {
		fmt.Fprintf(s, "<NiLIntv>")
		return
	}
	fmt.Fprintf(s, "[")
	if x.inf == nil {
		fmt.Fprintf(s, "<NiL>")
	} else {
		x.inf.Format(s, format)
	}
	fmt.Fprintf(s, ",")
	if x.sup == nil {
		fmt.Fprintf(s, "<NiL>")
	} else {
		x.sup.Format(s, format)
	}
	fmt.Fprintf(s, "]")
}

func newInterval(prec uint) *Interval {
	f := new(Interval)
	f.inf = new(big.Float)
	f.inf.SetMode(big.ToNegativeInf)
	f.inf.SetPrec(prec)
	f.sup = new(big.Float)
	f.sup.SetMode(big.ToPositiveInf)
	f.sup.SetPrec(prec)
	return f
}

func NewIntervalInt64(n int64, prec uint) *Interval {
	z := newInterval(prec)
	z.inf.SetInt64(n)
	z.sup.SetInt64(n)
	return z
}

func NewIntervalFloat(n *big.Float, prec uint) *Interval {
	z := newInterval(prec)
	z.SetFloat(n)
	return z
}

func (z *Interval) SetFloat(n *big.Float) {
	z.inf.Set(n)
	z.sup.Set(n)
}

func (z *Interval) SetInt64(n int64) {
	z.inf.SetInt64(n)
	z.sup.SetInt64(n)
}

func (z *Interval) SetFloat64(n float64) {
	z.inf.SetFloat64(n)
	z.sup.SetFloat64(n)
}

func (z *Interval) clonePrec(prec uint) *Interval {
	if prec == z.Prec() {
		return z
	}
	x := newInterval(prec)
	x.inf.Set(z.inf)
	x.sup.Set(z.sup)
	return x
}

func (z *Interval) SetPrec(prec uint) {
	z.inf.SetPrec(prec)
	z.sup.SetPrec(prec)
}

func (z *Interval) Prec() uint {
	return z.inf.Prec()
}

func (c *Interval) ContainsZero() bool {
	lsgn := c.inf.Sign()
	usgn := c.sup.Sign()

	return lsgn <= 0 && usgn >= 0
}

func MaxPrec(x, y *Interval) uint {
	if x.inf.Prec() <= y.inf.Prec() {
		return x.inf.Prec()
	} else {
		return y.inf.Prec()
	}
}

func (x *Interval) Tag() uint {
	return TAG_NUM
}

func (x *Interval) Neg() RObj {
	z := newInterval(x.Prec())
	z.sup.Neg(x.inf)
	z.inf.Neg(x.sup)
	return z
}

func (x *Interval) Add(yy RObj) RObj {
	y := yy.(*Interval)
	z := newInterval(MaxPrec(x, y))
	z.inf.Add(x.inf, y.inf)
	z.sup.Add(x.sup, y.sup)
	if err := z.valid(); err != nil {
		panic(err.Error())
	}
	return z
}

func (x *Interval) Sub(yy RObj) RObj {
	y := yy.(*Interval)
	z := newInterval(MaxPrec(x, y))
	z.sup.Sub(x.sup, y.inf)
	z.inf.Sub(x.inf, y.sup)
	if err := z.valid(); err != nil {
		panic(err.Error())
	}
	return z
}

func (x *Interval) AddFloat(y *big.Float) *Interval {
	z := newInterval(x.Prec())
	z.inf = z.inf.Add(x.inf, y)
	z.sup = z.sup.Add(x.sup, y)
	if err := z.valid(); err != nil {
		panic(err.Error())
	}
	return z
}

func (x *Interval) SubFloat(y *big.Float) *Interval {
	z := newInterval(x.Prec())
	z.inf = z.inf.Sub(x.inf, y)
	z.sup = z.sup.Sub(x.sup, y)
	if err := z.valid(); err != nil {
		panic(err.Error())
	}
	return z
}

func (x *Interval) MulFloat(y *big.Float) *Interval {
	z := newInterval(x.Prec())
	if y.Sign() >= 0 {
		z.inf = z.inf.Mul(x.inf, y)
		z.sup = z.sup.Mul(x.sup, y)
	} else {
		z.sup = z.sup.Mul(x.inf, y)
		z.inf = z.inf.Mul(x.sup, y)
	}

	if err := z.valid(); err != nil {
		panic(err.Error())
	}
	return z
}

func (x *Interval) MulInt64(y int64) *Interval {
	z := newInterval(x.Prec())
	yy := new(big.Float)
	yy.SetInt64(y)
	if y >= 0 {
		z.inf = z.inf.Mul(x.inf, yy)
		z.sup = z.sup.Mul(x.sup, yy)
	} else {
		z.sup = z.sup.Mul(x.inf, yy)
		z.inf = z.inf.Mul(x.sup, yy)
	}

	if err := z.valid(); err != nil {
		panic(err.Error())
	}
	return z
}

func (x *Interval) Mul(yy RObj) RObj {
	y := yy.(*Interval)
	z := newInterval(MaxPrec(x, y))
	if x.IsZero() {
		return x
	} else if y.IsZero() {
		return y
	}
	if x.inf.Sign() >= 0 {
		if y.inf.Sign() >= 0 {
			z.inf.Mul(x.inf, y.inf)
			z.sup.Mul(x.sup, y.sup)
		} else if y.sup.Sign() <= 0 {
			// x >= 0, y <= 0
			z.inf.Mul(x.sup, y.inf)
			z.sup.Mul(x.inf, y.sup)
		} else {
			z.inf.Mul(x.sup, y.inf)
			z.sup.Mul(x.sup, y.sup)
		}
	} else if x.sup.Sign() <= 0 {
		if y.inf.Sign() >= 0 {
			z.inf.Mul(x.inf, y.sup)
			z.sup.Mul(x.sup, y.inf)
		} else if y.sup.Sign() <= 0 {
			// [-xl, -xu] * [-yl, -yu] => [xu*yu, xl*yl]
			z.inf.Mul(x.sup, y.sup)
			z.sup.Mul(x.inf, y.inf)
		} else {
			// [-xl, -xu] * [-yl, +yu] => [xu*yu, xl*yl]
			z.inf.Mul(x.inf, y.sup)
			z.sup.Mul(x.inf, y.inf)
		}
	} else {
		if y.inf.Sign() >= 0 {
			// [-xl, xu] * [yl, yu]
			z.inf.Mul(x.inf, y.sup)
			z.sup.Mul(x.sup, y.sup)
		} else if y.sup.Sign() <= 0 {
			// [-xl, xu] * [-yl, -yu]
			z.inf.Mul(x.sup, y.inf)
			z.sup.Mul(x.inf, y.inf)
		} else {
			// [-xl, +xu] * [-yl, +yu] => [min(-xl*yu,-xu,yl), max(xl*yl,xu*yu)]
			u := new(big.Float)
			u.SetPrec(z.inf.Prec())
			u.SetMode(big.ToNegativeInf)
			u.Mul(x.inf, y.sup)
			z.inf.Mul(x.sup, y.inf)
			cmp := u.Cmp(z.inf)
			if cmp < 0 {
				z.inf.Set(u)
			}
			u.SetMode(big.ToPositiveInf)
			u.Mul(x.inf, y.inf)
			z.sup.Mul(x.sup, y.sup)
			cmp = u.Cmp(z.sup)
			if cmp > 0 {
				z.sup.Set(u)
			}
		}
	}

	if false {
		k := new(big.Float)
		k.SetPrec(z.Prec())
		for _, xx := range []*big.Float{x.inf, x.sup} {
			for _, yy := range []*big.Float{y.inf, y.sup} {
				k.SetMode(big.ToPositiveInf)
				k.Mul(xx, yy)
				if k.Cmp(z.sup) > 0 {
					panic("1")
				}

				k.SetMode(big.ToNegativeInf)
				if z.inf.Cmp(k) > 0 {
					panic("2")
				}
			}
		}
	}
	if err := z.valid(); err != nil {
		panic(err.Error())
	}

	return z
}

func (x *Interval) Div(yy NObj) RObj {
	// zero を含まないと仮定.
	y := yy.(*Interval)
	one := big.NewFloat(1.)
	z := newInterval(y.Prec())
	z.sup.Quo(one, y.inf)
	z.inf.Quo(one, y.sup)
	return x.Mul(z)
}

func (x *Interval) Pow(yy *Int) RObj {
	return nil
}

func (x *Interval) Sign() int {
	// sign()=0 はゼロではなく，未知..
	// ゼロかどうかは IsZero() を利用せよ
	if x.inf.Sign() > 0 {
		return 1
	} else if x.sup.Sign() < 0 {
		return -1
	} else {
		return 0
	}
}

func (x *Interval) IsZero() bool {
	return x.inf.Sign() >= 0 && x.sup.Sign() <= 0
}

func (x *Interval) IsOne() bool {
	return false
}

func (x *Interval) IsMinusOne() bool {
	return false
}

func (x *Interval) IsNumeric() bool {
	return true
}

func (x *Interval) valid() error {
	if x.inf.Cmp(x.sup) > 0 {
		return fmt.Errorf("lv > uv")
	}
	return nil
}

func (x *Interval) mul_2exp(m uint) RObj {
	return x.Mul2Exp(m)
}

func (x *Interval) numTag() uint {
	return NTAG_INTERVAL
}

func (x *Interval) Float() float64 {
	return 0.
}

func (x *Interval) Mul2Exp(m uint) NObj {
	return nil
}
func (x *Interval) Cmp(y NObj) int {
	return 0
}
func (x *Interval) CmpAbs(y NObj) int {
	return 0
}
func (x *Interval) Abs() NObj {
	return x
}
func (x *Interval) subst_poly(p *Poly, lv Level) RObj {
	return x
}
func (x *Interval) Equals(v interface{}) bool {
	return false
}

func (x *Interval) Subst(v RObj, lv Level) RObj {
	return x
}

func (x *Interval) toIntv(prec uint) RObj {
	return x
}

func (x *Interval) mid(p float64, prec uint) *big.Float {
	l := big.NewFloat(p)
	r := big.NewFloat(1 - p)
	l.Mul(l, x.inf)
	r.Mul(r, x.sup)
	r.Add(l, r)
	return r
}
