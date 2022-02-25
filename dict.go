package ganrac

import (
	"fmt"
)

type Dict struct {
	GObj
	v map[string]GObj
}

func (z *Dict) Tag() uint {
	return TAG_LIST
}

func (z *Dict) String() string {
	sep := "{"
	s := ""
	for k, v := range z.v {
		s += fmt.Sprintf("%s%s:%s", sep, k, v.String())
		sep = ", "
	}
	return s + "}"
}

func (z *Dict) Format(s fmt.State, format rune) {
	left := "{"
	right := "}"
	sep := ":"
	switch format {
	case FORMAT_DUMP:
		left = fmt.Sprintf("(dict %d", len(z.v))
		right = ")"
		sep = " "
	case FORMAT_TEX:
		left = "\\left["
		right = "\\right]"
	case FORMAT_SRC:
		left = "NewDict("
		right = ")"
	}

	seq := left
	for k, v := range z.v {
		fmt.Fprintf(s, "%s\"%s\"%s", seq, k, sep)
		v.Format(s, format)
		seq = ","

	}
	if seq != "," {
		fmt.Fprintf(s, "%s", seq)
	}
	fmt.Fprintf(s, "%s", right)
}

func (z *Dict) Len() int {
	return len(z.v)
}

func (z *Dict) Equals(vv interface{}) bool {
	v, ok := vv.(*Dict)
	if !ok || len(z.v) != len(v.v) {
		return false
	}
	for k, vval := range v.v {
		if zval, ok := z.v[k]; !ok {
			return false
		} else if vvalz, ok := vval.(equaler); !ok {
			return false
		} else if !vvalz.Equals(zval) {
			return false
		}
	}

	return true
}

func (z *Dict) Indets(b []bool) {
	for _, p := range z.v {
		q, ok := p.(indeter)
		if ok {
			q.Indets(b)
		}
	}
}

func NewDict() *Dict {
	d := new(Dict)
	d.v = make(map[string]GObj)
	return d
}

func (z *Dict) Set(k string, v GObj) error {
	z.v[k] = v
	return nil
}

func (z *Dict) Get(k string) (GObj, error) {
	v, ok := z.v[k]
	if ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("element not found")
	}
}
