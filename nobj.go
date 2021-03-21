package ganrac

// numeric
type NObj interface {
	GObj
	Cmp(x NObj) int
	CmpAbs(x NObj) int
}

type Number struct {
}

func (x *Number) Tag() uint {
	return TAG_INT
}

func (x *Number) IsNumeric() bool {
	return true
}
