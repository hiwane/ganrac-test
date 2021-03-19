package ganrac

import (
	"testing"
)

func TestPoly(t *testing.T) {
	for _, v := range []struct {
		c1, c0 int64
		exp    string
	}{
		{1, 1, "x+1"},
		{1, 0, "x"},
		{1, -1, "x-1"},
		{-1, 1, "-x+1"},
		{-1, 0, "-x"},
		{-1, -1, "-x-1"},
		{+2, 3, "2*x+3"},
		{+2, 0, "2*x"},
		{+2, -3, "2*x-3"},
		{-2, 3, "-2*x+3"},
		{-2, 0, "-2*x"},
		{-2, -3, "-2*x-3"},
	} {
		p := new(Poly)
		p.c = make([]Coef, 0)
		p.c = append(p.c, NewInt(v.c0))
		p.c = append(p.c, NewInt(v.c1))
		if p.String() != v.exp {
			t.Errorf("invalid poly p=%v, exp=%s, [%d,%d]", p, v.exp, v.c1, v.c0)
		}

		q := NewPolyInts(p.lv, v.c0, v.c1)
		if q.String() != v.exp || !p.Equals(q) || !q.Equals(p) {
			t.Errorf("invalid poly q=%v, exp=%s, [%d,%d]", q, v.exp, v.c1, v.c0)
		}
	}

	for _, v := range []struct {
		exp        string
		c2, c1, c0 int64
	}{
		{"x^2+x+1", 1, 1, 1},
		{"x^2+1", 1, 0, 1},
		{"x^2", 1, 0, 0},
		{"-x^2", -1, 0, 0},
		{"x^2-1", 1, 0, -1},
		{"2*x^2+3*x+4", 2, 3, 4},
		{"-2*x^2-3*x-4", -2, -3, -4},
		{"2*x^2-x+4", 2, -1, 4},
	} {
		p := new(Poly)
		p.c = make([]Coef, 0)
		p.c = append(p.c, NewInt(v.c0))
		p.c = append(p.c, NewInt(v.c1))
		p.c = append(p.c, NewInt(v.c2))
		if p.String() != v.exp {
			t.Errorf("invalid poly p=%v, exp=%s, [%d,%d,%d]", p, v.exp, v.c2, v.c1, v.c0)
		}

		q := NewPolyInts(p.lv, v.c0, v.c1, v.c2)
		if q.String() != v.exp || !p.Equals(q) || !q.Equals(p) {
			t.Errorf("invalid poly q=%v, exp=%s, [%d,%d,%d]", q, v.exp, v.c2, v.c1, v.c0)
		}
	}
}
