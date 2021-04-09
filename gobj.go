package ganrac

import "io"

const (
	TAG_NONE = iota
	TAG_STR
	TAG_NUM
	TAG_POLY
	TAG_FOF
	TAG_LIST
	TAG_CAD
)

// ganrac object
type GObj interface {
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

type dumper interface {
	dump(b io.Writer) // for debug print
}

type printer interface {
	Print(b io.Writer, args ...interface{}) error
}
