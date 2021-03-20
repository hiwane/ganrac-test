package ganrac

// ring ring
// in R[x], in R
type RObj interface {
	Gobj
	Add(x RObj) RObj // z+x
	Sub(x RObj) RObj // z-x
	Mul(x RObj) RObj
	Pow(x *Int) RObj
	Neg() RObj
	//	Set(x RObj) RObj
	Sign() int
	IsZero() bool
	IsOne() bool
	IsMinusOne() bool
	IsNumeric() bool
	Equals(x RObj) bool
	New() RObj
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

func (z *RObjSample) New() RObj {
	v := new(RObjSample)
	return v
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
	// サボり ver
	xn := x.Neg()
	return z.Add(xn)
}

func (z *RObjSample) Mul(x RObj) RObj {
	return z
}

func (z *RObjSample) Pow(x *Int) RObj {
	return z
}

func (z *RObjSample) Equals(x RObj) bool {
	return false
}
