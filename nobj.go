package ganrac

// numeric
type NObj interface {
	RObj
	Float() float64
	Cmp(x NObj) int
	CmpAbs(x NObj) int
	Abs() NObj
	AddInt(n int64) NObj
	ToInt(n int) *Int // 整数に丸める.
}

type Number struct {
}

func (x *Number) Tag() uint {
	return TAG_INT
}

func (x *Number) IsNumeric() bool {
	return true
}
