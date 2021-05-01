package ganrac

import (
	"fmt"
)

const (
	TAG_NONE = iota
	TAG_STR
	TAG_NUM
	TAG_POLY
	TAG_FOF
	TAG_LIST
	TAG_CAD

	FORMAT_TEX  = 'P'
	FORMAT_DUMP = 'V'
	FORMAT_SRC  = 'S'
)

// ganrac object
// RObj, NObj, Fof, List, *String
type GObj interface {
	fmt.Formatter
	String() string
	Tag() uint
}

type lener interface {
	Len() int
}

type indeter interface {
	Indets(b []bool)
}

type equaler interface {
	Equals(v interface{}) bool
}

type subster interface {
	Subst(xs []RObj, lvs []Level) GObj
}

func gobjToIntv(g GObj, prec uint) GObj {
	switch a := g.(type) {
	case RObj:
		return a.toIntv(prec)
	case *List:
		return a.toIntv(prec)
	default:
		return a
	}
}

func gobjSubst(g GObj, rr []RObj, lv []Level) GObj {
	switch f := g.(type) {
	case subster:
		return f.Subst(rr, lv)
	case *List:
		return f.Subst(rr, lv)
	case RObj:
		return f.Subst(rr, lv, 0)
	case Fof:
		return f.Subst(rr, lv)
	default:
		return f
	}
}
