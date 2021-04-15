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
