package ganrac

type Str struct {
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
