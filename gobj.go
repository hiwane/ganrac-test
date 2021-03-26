package ganrac

const (
	TAG_NONE = iota
	TAG_STR
	TAG_INT
	TAG_POLY
	TAG_FOF
	TAG_LIST
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
