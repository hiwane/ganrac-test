package ganrac

type String struct {
	GObj
	s string
}

func NewString(s string) *String {
	p := new(String)
	p.s = s
	return p
}

func (s *String) String() string {
	return "\"" + s.s + "\""
}

func (s *String) Tag() uint {
	return TAG_STR
}
