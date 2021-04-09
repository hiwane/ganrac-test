package ganrac

const (
	NTAG_NONE   uint = 0
	NTAG_INT    uint = 1
	NTAG_RAT    uint = 2
	NTAG_BININT uint = 3
)

// numeric
type NObj interface {
	RObj
	numTag() uint
	Float() float64
	Mul2Exp(m uint) NObj
	Cmp(x NObj) int
	CmpAbs(x NObj) int
	Abs() NObj
	subst_poly(p *Poly, lv Level) RObj

	// ToInt(n int) *Int // 整数に丸める.
}

type Number struct {
}

/////////////////////////////////////
// GObj
/////////////////////////////////////
func (x *Number) numTag() uint {
	return NTAG_NONE
}

func (x *Number) Tag() uint {
	return TAG_NUM
}

func (x *Number) String() string {
	return "not implemented"
}

/////////////////////////////////////
// RObj
/////////////////////////////////////
// func (x *Number) Equals(yy interface{}) bool {
// 	y, ok := yy.(NObj)
// 	if !ok {
// 		return false
// 	}
// 	return x.Cmp(y) == 0
// }

// func (x *Number) Add(yy RObj) RObj {
// 	panic("not implemented")
// }
//
// func (x *Number) Sub(yy RObj) RObj {
// 	return x.Add(yy.Neg())
// }
//
// func (x *Number) Mul(yy RObj) RObj {
// 	panic("not implemented")
// }
// func (x *Number) Div(yy NObj) RObj {
// 	panic("not implemented")
// }
// func (x *Number) Pow(yy *Int) RObj {
// 	panic("not implemented")
// }
// func (x *Number) Subst(yy []RObj, lv []Level, n int) RObj {
// 	return x
// }
// func (x *Number) Neg() RObj {
// 	panic("not implemented")
// }
// func (x *Number) Sign() int {
// 	panic("not implemented")
// }
//
// func (x *Number) IsZero() bool {
// 	return x.Sign() == 0
// }
//
// func (x *Number) IsOne() bool {
// 	panic("not implemented")
// }
//
// func (x *Number) IsMinusOne() bool {
// 	panic("not implemented")
// }

func (x *Number) IsNumeric() bool {
	return true
}

func (x *Number) valid() error {
	return nil
}

/////////////////////////////////////
// NObj
/////////////////////////////////////

// func (x *Number) Float() float64 {
// 	panic("unimplemented")
// }
//
// func (x *Number) Cmp(y NObj) int {
// 	panic("unimplemented")
// }
//
// func (x *Number) CmpAbs(y NObj) int {
// 	panic("unimplemented")
// }
//
// func (x *Number) Abs() NObj {
// 	if x.Sign() >= 0 {
// 		return x
// 	} else {
// 		return x.Neg().(NObj)
// 	}
// }
//
// func (x *Number) AddInt64(n int64) NObj {
// 	return x.Add(NewInt(n)).(NObj)
// }
