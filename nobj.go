package ganrac

const (
	NTAG_NONE = iota
	NTAG_INT
	NTAG_RAT
	NTAG_BININT
	NTAG_INTERVAL
	NTAG_MOD
)

// numeric
// *Int, *Rat, *BinInt, *Interval
type NObj interface {
	RObj
	numTag() uint
	Float() float64
	Cmp(x NObj) int
	CmpAbs(x NObj) int
	Abs() NObj
	subst_poly(p *Poly, lv Level) RObj

	// ToInt(n int) *Int // 整数に丸める.
}

type Number struct {
}

func (x *Number) numTag() uint {
	return NTAG_NONE
}

func (x *Number) Tag() uint {
	return TAG_NUM
}

func (x *Number) String() string {
	return "not implemented"
}

func (x *Number) IsNumeric() bool {
	return true
}

func (x *Number) valid() error {
	return nil
}
