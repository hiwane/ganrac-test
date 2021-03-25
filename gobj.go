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

type Lener interface {
	Len() int
}

type Indeter interface {
	Indets(b []bool)
}
