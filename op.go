package ganrac

import (
	"fmt"
)

// グローバル変数にしたくないのだけど，
// どうやって保持すべきなのか....
var varlist []varInfo

var varstr2lv map[string]Level

type varInfo struct {
	v string
	p *Poly
}

func varstr(lv Level) string {
	if 0 <= lv && int(lv) < len(varlist) {
		return varlist[lv].v
	} else {
		return fmt.Sprintf("_<%d>", lv)
	}
}

func var2lv(v string) (Level, error) {
	for i, x := range varlist {
		if x.v == v {
			return Level(i), nil
		}
	}
	return 0, fmt.Errorf("undefined variable `%s`.", v)
}

func (g *Ganrac) InitVarList(vlist []string) error {
	for i, v := range vlist {
		if v == "init" {
			return fmt.Errorf("%s is reserved", v)
		}
		for _, bft := range g.builtin_func_table {
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

	varlist = make([]varInfo, len(vlist))
	varstr2lv = make(map[string]Level, len(vlist))
	for i := 0; i < len(vlist); i++ {
		varlist[i] = varInfo{vlist[i], NewPolyInts(Level(i), 0, 1)}
		varstr2lv[vlist[i]] = Level(i)
	}

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
	if y.IsNumeric() || !x.IsNumeric() {
		return x.Sub(y)
	} else {
		// num - poly
		return y.Neg().Add(x)
	}
}

func Mul(x, y RObj) RObj {
	if x.Tag() >= y.Tag() {
		return x.Mul(y)
	} else {
		return y.Mul(x)
	}
}
