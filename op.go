package ganrac

import (
	"fmt"
)

var varlist = []string{
	"x", "y", "z", "w", "a", "b", "c", "e", "f", "g", "h",
}

func var2lv(v string) (Level, error) {
	for i, x := range varlist {
		if x == v {
			return Level(i), nil
		}
	}
	return 0, fmt.Errorf("undefined variable `%s`.", v)
}

func InitVar(vlist []string) error {
	for i, v := range vlist {
		if v == "init" {
			return fmt.Errorf("%s is reserved", v)
		}
		for _, bft := range builtin_func_table {
			if v == bft.name {
				return fmt.Errorf("%s is reserved", v)
			}
		}
		for j := 0; j < i; j++ {
			if vlist[j] == v {
				return fmt.Errorf("%s is duplicated", v)
			}
		}
	}

	varlist = vlist
	return nil
}

func Add(x, y RObj) RObj {
	if x.Tag() >= y.Tag() {
		return x.Add(y)
	} else {
		return y.Add(x)
	}
}

func Sub(x, y RObj) RObj {
	return Add(x, y.Neg())
}

func Mul(x, y RObj) RObj {
	if x.Tag() >= y.Tag() {
		return x.Mul(y)
	} else {
		return y.Mul(x)
	}
}
