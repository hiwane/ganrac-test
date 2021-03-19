package ganrac

const (
	TAG_NONE = iota
	TAG_INT
	TAG_POLY
)

type Coef interface {
	Add(x Coef) Coef
	Sub(x Coef) Coef
	Mul(x Coef) Coef
	Neg() Coef
	Set(x Coef) Coef
	String() string
	Sign() int
	IsZero() bool
	IsOne() bool
	IsMinusOne() bool
	IsNumeric() bool
	Equals(x Coef) bool
	New() Coef
	Tag() uint
}

type CoefSample struct {
}

func (z *CoefSample) Tag() uint {
	return TAG_NONE
}

func (z *CoefSample) String() string {
	return "sample"
}

func (z *CoefSample) Sign() int {
	return 0
}

func (z *CoefSample) IsZero() bool {
	return false
}

func (z *CoefSample) IsOne() bool {
	return false
}

func (z *CoefSample) IsMinusOne() bool {
	return false
}

func (z *CoefSample) IsNumeric() bool {
	return false
}

func (z *CoefSample) New() Coef {
	v := new(CoefSample)
	return v
}

func (z *CoefSample) Set(x Coef) Coef {
	return z
}

func (z *CoefSample) Neg() Coef {
	return z
}

func (z *CoefSample) Add(x Coef) Coef {
	return z
}

func (z *CoefSample) Sub(x Coef) Coef {
	// サボり ver
	xn := x.Neg()
	return z.Add(xn)
}

func (z *CoefSample) Mul(x Coef) Coef {
	return z
}

func (z *CoefSample) Equals(x Coef) bool {
	return false
}
