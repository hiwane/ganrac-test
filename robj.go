package ganrac

import (
	"fmt"
)

// ring ring
// in R[x], in R
// *Poly, *Int, *Rat, *BinInt, *Interval
type RObj interface {
	GObj
	equaler
	Add(x RObj) RObj // z+x
	Sub(x RObj) RObj // z-x
	Mul(x RObj) RObj
	Div(x NObj) RObj
	Pow(x *Int) RObj
	Subst(x RObj, lv Level) RObj
	Neg() RObj
	//	Set(x RObj) RObj
	Sign() int
	IsZero() bool
	IsOne() bool
	IsMinusOne() bool
	IsNumeric() bool
	valid() error
	mul_2exp(m uint) RObj
	toIntv(prec uint) RObj
}

type RObjSample struct {
}

func (z *RObjSample) Tag() uint {
	return TAG_NONE
}

func (z *RObjSample) String() string {
	return "sample"
}

func (z *RObjSample) Sign() int {
	return 0
}

func (z *RObjSample) IsZero() bool {
	return false
}

func (z *RObjSample) IsOne() bool {
	return false
}

func (z *RObjSample) IsMinusOne() bool {
	return false
}

func (z *RObjSample) IsNumeric() bool {
	return false
}

func (z *RObjSample) Set(x RObj) RObj {
	return z
}

func (z *RObjSample) Neg() RObj {
	return z
}

func (z *RObjSample) Add(x RObj) RObj {
	return z
}

func (z *RObjSample) Sub(x RObj) RObj {
	// @TODO
	xn := x.Neg()
	return z.Add(xn)
}

func (z *RObjSample) Mul(x RObj) RObj {
	return z
}

func (z *RObjSample) Div(x NObj) RObj {
	return z
}

func (z *RObjSample) Pow(x *Int) RObj {
	return z
}

func (z *RObjSample) Equals(x interface{}) bool {
	return false
}

func (z *RObjSample) Subst(x RObj, lv Level) RObj {
	return z
}

func (z *RObjSample) valid() error {
	return nil
}

func (z *RObjSample) mul_2exp(m uint) RObj {
	return z
}

func (z *RObjSample) toIntv(prec uint) RObj {
	return z
}

func (z *RObjSample) Format(s fmt.State, format rune) {
}
