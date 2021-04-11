package ganrac

// I/F は big.Int に揃える
import (
	"fmt"
	"math/big"
)

type Interval struct {
	lv *big.Float // lower value
	uv *big.Float // upper value
}

func (z *Interval) String() string {
	return fmt.Sprintf("[%s,%s]", z.lv.String(), z.uv.String())
}

func newInterval(prec uint) *Interval {
	f := new(Interval)
	f.lv.SetMode(big.ToNegativeInf)
	f.lv.SetPrec(prec)
	f.uv.SetMode(big.ToNegativeInf)
	f.uv.SetPrec(prec)
	return f
}

func NewIntervalInt64(n int64, prec uint) *Interval {
	z := newInterval(prec)
	z.lv.SetInt64(n)
	z.uv.SetInt64(n)
	return z
}

func (z *Interval) clonePrec(prec uint) *Interval {
	if prec == z.Prec() {
		return z
	}
	x := newInterval(prec)
	x.lv.Set(z.lv)
	x.uv.Set(z.uv)
	return x
}

func (z *Interval) SetPrec(prec uint) {
	z.lv.SetPrec(prec)
	z.uv.SetPrec(prec)
}

func (z *Interval) Prec() uint {
	return z.lv.Prec()
}

func (c *Interval) ContainsZero() bool {
	lsgn := c.lv.Sign()
	usgn := c.uv.Sign()

	return lsgn <= 0 && usgn >= 0
}

func MaxPrec(x, y *Interval) uint {
	if x.lv.Prec() <= y.lv.Prec() {
		return x.lv.Prec()
	} else {
		return y.lv.Prec()
	}
}

func (x *Interval) Tag() uint {
	return TAG_NUM
}

func (x *Interval) Neg() RObj {
	z := newInterval(x.Prec())
	z.uv.Neg(x.lv)
	z.lv.Neg(x.uv)
	return z
}

func (x *Interval) Add(yy RObj) RObj {
	y := yy.(*Interval)
	z := newInterval(MaxPrec(x, y))
	z.lv.Add(x.lv, y.lv)
	z.uv.Add(x.uv, y.uv)
	return z
}

func (x *Interval) Sub(yy RObj) RObj {
	y := yy.(*Interval)
	z := newInterval(MaxPrec(x, y))
	z.lv.Add(x.lv, y.lv)
	z.uv.Add(x.uv, y.uv)
	return z
}

func (x *Interval) Mul(yy RObj) RObj {
	y := yy.(*Interval)
	z := newInterval(MaxPrec(x, y))
	if x.lv.Sign() >= 0 {
		if y.lv.Sign() >= 0 {
			z.lv.Mul(x.lv, y.lv)
			z.uv.Mul(x.uv, y.uv)
		} else if y.uv.Sign() <= 0 {
			// x >= 0, y <= 0
			z.lv.Mul(x.uv, y.lv)
			z.uv.Mul(x.lv, y.uv)
		} else {
			z.lv.Mul(x.uv, y.lv)
			z.uv.Mul(x.uv, y.uv)
		}
	} else if x.uv.Sign() <= 0 {
		if y.lv.Sign() >= 0 {
			z.lv.Mul(x.lv, y.uv)
			z.uv.Mul(x.uv, y.lv)
		} else if y.uv.Sign() <= 0 {
			// [-xl, -xu] * [-yl, -yu] => [xu*yu, xl*yl]
			z.lv.Mul(x.uv, y.uv)
			z.uv.Mul(x.lv, y.lv)
		} else {
			// [-xl, -xu] * [-yl, +yu] => [xu*yu, xl*yl]
			z.lv.Mul(x.lv, y.uv)
			z.uv.Mul(x.lv, y.lv)
		}
	} else {
		if y.lv.Sign() >= 0 {
			// [-xl, xu] * [yl, yu]
			z.lv.Mul(x.lv, y.uv)
			z.uv.Mul(x.uv, y.uv)
		} else if y.uv.Sign() <= 0 {
			// [-xl, xu] * [-yl, -yu]
			// [-xl, xu] * (-yl, -yu]
			z.lv.Mul(x.uv, y.lv)
			z.uv.Mul(x.lv, y.lv)
		} else {
			// [-xl, +xu] * [-yl, +yu] => [min(-xl*yu,-xu,yl), max(xl*yl,xu*yu)]
			u := new(big.Float)
			u.SetPrec(z.lv.Prec())
			u.SetMode(big.ToNegativeInf)
			u.Mul(x.lv, y.uv)
			z.lv.Mul(x.uv, y.lv)
			cmp := u.Cmp(z.lv)
			if cmp < 0 {
				z.lv.Set(u)
			}
			u.SetMode(big.ToPositiveInf)
			u.Mul(x.lv, y.lv)
			z.uv.Mul(x.uv, y.uv)
			cmp = u.Cmp(z.uv)
			if cmp > 0 {
				z.uv.Set(u)
			}
		}
	}

	return z
}

func (x *Interval) Div(yy NObj) RObj {
	return nil
}

func (x *Interval) Pow(yy *Int) RObj {
	return nil
}

func (x *Interval) Sign() int {
	if x.lv.Sign() > 0 {
		return 1
	} else if x.uv.Sign() < 0 {
		return -1
	} else {
		return 0
	}
}

func (x *Interval) IsZero() bool {
	return x.lv.Sign() >= 0 && x.uv.Sign() <= 0
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
	if x.lv.Cmp(x.uv) > 0 {
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

func (x *Interval) Subst(v []RObj, lv []Level, n int) RObj {
	return x
}

func (x *Interval) toIntv(prec uint) RObj {
	return x
}
