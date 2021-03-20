package ganrac

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

func NewList(args []interface{}) *List {
	lst := new(List)
	lst.v = make([]GObj, len(args))
	for i := 0; i < len(args); i++ {
		lst.v[i] = args[i].(GObj)
	}
	return lst
}
