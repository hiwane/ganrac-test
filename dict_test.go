package ganrac

import (
	"strings"
	"testing"
)

func TestDictParser(t *testing.T) {
	g := NewGANRAC()
	for i, s := range []struct {
		line string
		d    map[string]GObj
		ret  equaler
	}{
		{"A={};", map[string]GObj{}, nil},
		{"A[\"a\"] = 3;", map[string]GObj{"a": NewInt(3)}, NewInt(3)},
		{"A[\"b\"] = 5;", map[string]GObj{"a": NewInt(3), "b": NewInt(5)}, NewInt(5)},
		{"A[\"a\"] = 7;", map[string]GObj{"a": NewInt(7), "b": NewInt(5)}, NewInt(7)},
		{"A={a: 3, b: 11};", map[string]GObj{"a": NewInt(3), "b": NewInt(11)}, nil},
		{"A[\"c\"] = [1, 2];", map[string]GObj{"a": NewInt(3), "b": NewInt(11),
			"c": NewList(NewInt(1), NewInt(2))},
			NewList(NewInt(1), NewInt(2))},
	} {
		ret, err := g.Eval(strings.NewReader(s.line))
		if err != nil {
			t.Errorf("[%d] invalid %s\nerr=%v", i, s.line, err)
			break
		}

		D := NewDict()
		D.v = s.d
		if s.ret == nil && !D.Equals(ret) ||
			s.ret != nil && !s.ret.Equals(ret) {
			t.Errorf("[%d] invalid %s, return\nret=%v", i, s.line, ret)
			break
		}
		A := g.varmap["A"]
		a, ok := A.(*Dict)
		if !ok {
			t.Errorf("[%d] invalid %s\nnot dict %v", i, s.line, A)
			break
		}
		if a.Len() != len(s.d) {
			t.Errorf("[%d] invalid %s\nlen. expect=%d, actual=%d, %v", i, s.line, a.Len(), len(s.d), a)
			break
		}
		for k, v := range s.d {
			if av, ok := a.v[k]; !ok {
				t.Errorf("[%d] invalid %s not found k=%s\nexpect=%v\nactual=%v", i, s.line, k, s.d, a)
			} else if avv, ok := av.(equaler); !ok {
				t.Errorf("[%d] invalid %s not equaler k=%s\nexpect=%v\nactual=%v", i, s.line, k, s.d, a)
			} else if !avv.Equals(v) {
				t.Errorf("[%d] invalid %s not equal k=%s\nexpect=%v\nactual=%v", i, s.line, k, s.d, a)
			} else {
				continue
			}
			break
		}
	}
}
