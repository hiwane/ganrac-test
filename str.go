package ganrac

type Str struct {
	GObj
	s string
}

func NewString(s string) *Str {
	p := new(Str)
	p.s = s
	return p
}

func (s *Str) String() string {
	return "\"" + s.s + "\""
}

func (s *Str) Tag() uint {
	return TAG_STR
}
