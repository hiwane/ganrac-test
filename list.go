package ganrac

import (
	"fmt"
)

type List struct {
	GObj
	v []GObj
}

func (z *List) Tag() uint {
	return TAG_LIST
}

func (z *List) String() string {
	s := "["
	for i := 0; i < len(z.v); i++ {
		if i != 0 {
			s += ","
		}
		s += z.v[i].String()
	}
	return s + "]"
}

func (z *List) Get(ii *Int) (GObj, error) {
	ilen := NewInt(int64(len(z.v)))
	if ii.Sign() < 0 || ii.Cmp(ilen) >= 0 {
		return nil, fmt.Errorf("list index out of range")
	}
	m := int(ii.n.Int64())
	return z.v[m], nil
}

func (z *List) Geti(i int) (GObj, error) {
	if i < 0 || i >= len(z.v) {
		return nil, fmt.Errorf("list index out of range")
	}
	return z.v[i], nil
}

func (z *List) Len() int {
	return len(z.v)
}

func (z *List) Indets(b []bool) {
	for _, p := range z.v {
		q, ok := p.(Indeter)
		if ok {
			q.Indets(b)
		}
	}
}

func (z *List) Append(a GObj) {
	z.v = append(z.v, a)
}

func NewList(args []interface{}) *List {
	lst := new(List)
	lst.v = make([]GObj, len(args))
	for i := 0; i < len(args); i++ {
		lst.v[i] = args[i].(GObj)
	}
	return lst
}
