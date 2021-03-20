package ganrac

import (
	"testing"
)

func TestPolyString(t *testing.T) {
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
		p.c = make([]RObj, 0)
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
		p.c = make([]RObj, 0)
		p.c = append(p.c, NewInt(v.c0))
		p.c = append(p.c, NewInt(v.c1))
		p.c = append(p.c, NewInt(v.c2))
		if p.String() != v.exp {
			t.Errorf("invalid poly.mul p=%v, exp=%s, [%d,%d,%d]", p, v.exp, v.c2, v.c1, v.c0)
		}

		q := NewPolyInts(p.lv, v.c0, v.c1, v.c2)
		if q.String() != v.exp || !p.Equals(q) || !q.Equals(p) {
			t.Errorf("invalid poly.mul q=%v, exp=%s, [%d,%d,%d]", q, v.exp, v.c2, v.c1, v.c0)
		}
	}
}

func TestPolyAdd(t *testing.T) {
	lv := Level(0)
	for _, s := range []struct {
		a, b   []int64
		expect RObj
	}{
		{[]int64{1, 1}, []int64{1, 2}, NewPolyInts(lv, 2, 3)},
		{[]int64{1, 2, 3}, []int64{4, 5}, NewPolyInts(lv, 5, 7, 3)},
		{[]int64{1, 4, 5}, []int64{1, 3, -5}, NewPolyInts(lv, 2, 7)},
		{[]int64{1, 1}, []int64{1, -1}, NewInt(2)},
		{[]int64{2, 1}, []int64{-2, -1}, NewInt(0)},
	} {
		a := NewPolyInts(lv, s.a...)
		b := NewPolyInts(lv, s.b...)
		c := a.Add(b)
		if !c.Equals(s.expect) {
			t.Errorf("invalid poly.add a=%v, b=%v, exp=%v, actual=%v", a, b, s.expect, c)
		}

		d := b.Add(a)
		if !d.Equals(s.expect) {
			t.Errorf("invalid poly.add b=%v, a=%v, exp=%v, actual=%v", b, a, s.expect, d)
		}
	}
}

func TestPolyAddLv(t *testing.T) {
	var a, b, c RObj
	var ep *Poly
	a = NewPolyInts(0, 2, 3, 4)
	b = NewPolyInts(1, 5, 6, 7)
	ep = NewPolyInts(0, 2, 3, 4)
	ep.c[0] = NewPolyInts(1, 7, 6, 7)

	c = a.Add(b)
	if !ep.Equals(c) {
		t.Errorf("invalid poly.add a=%v, b=%v, expect=%v, actual=%v", a, b, ep, c)
	}
	c = b.Add(a)
	if !ep.Equals(c) {
		t.Errorf("invalid poly.add b=%v, a=%v, expect=%v, actual=%v", a, b, ep, c)
	}

	b = NewInt(9)
	ep = NewPolyInts(0, 11, 3, 4)
	c = a.Add(b)
	if !ep.Equals(c) {
		t.Errorf("invalid poly.add a=%v, b=%v, expect=%v, actual=%v", a, b, ep, c)
	}
}

func TestPolyMul(t *testing.T) {
	lv := Level(0)
	for _, s := range []struct {
		a, b, expect []int64
	}{
		{[]int64{1, 2}, []int64{1, 1}, []int64{1, 3, 2}},
		{[]int64{1, 1}, []int64{1, 2, 1}, []int64{1, 3, 3, 1}},
		{[]int64{1, 1}, []int64{1, 3, 3, 1}, []int64{1, 4, 6, 4, 1}},
		{[]int64{2, 1}, []int64{4, 3, 1}, []int64{8, 10, 5, 1}},
	} {
		a := NewPolyInts(lv, s.a...)
		b := NewPolyInts(lv, s.b...)
		ep := NewPolyInts(lv, s.expect...)
		c := a.Mul(b)
		if !c.Equals(ep) {
			t.Errorf("invalid poly a=%v, b=%v, exp=%v, actual=%v", a, b, ep, c)
		}

		d := b.Mul(a)
		if !d.Equals(ep) {
			t.Errorf("invalid poly b=%v, a=%v, exp=%v, actual=%v", b, a, ep, d)
		}
	}
}

func TestPolyMulLv(t *testing.T) {
	var a, b, c RObj
	var ep *Poly
	a = NewPolyInts(0, 3, 5, 6) // 6*x^2+5x+3
	b = NewPolyInts(1, 7, 11)   // 11y+7
	ep = NewPoly(0, 3)          // 5*(11y+7)x + 3*(11y+7)
	ep.c[0] = NewPolyInts(1, 21, 33)
	ep.c[1] = NewPolyInts(1, 35, 55)
	ep.c[2] = NewPolyInts(1, 42, 66)

	c = a.Mul(b)
	if !ep.Equals(c) {
		t.Errorf("invalid poly.mul a=%v, b=%v, expect=%v, actual=%v", a, b, ep, c)
	}
	c = b.Mul(a)
	if !ep.Equals(c) {
		t.Errorf("invalid poly.mul b=%v, a=%v, expect=%v, actual=%v", a, b, ep, c)
	}

	m := int64(9)
	b = NewInt(m)
	ep = NewPolyInts(0, 3*m, 5*m, 6*m)
	c = a.Mul(b)
	if !ep.Equals(c) {
		t.Errorf("invalid poly.mul a=%v, b=%v, expect=%v, actual=%v", a, b, ep, c)
	}

	b = NewInt(0)
	c = a.Mul(b)
	if !b.Equals(c) {
		t.Errorf("invalid poly.mul a=%v, b=%v, expect=%v, actual=%v", a, b, b, c)
	}
}

func TestPolyPow(t *testing.T) {
	lv := Level(0)
	zero := NewInt(0)
	one := NewInt(1)
	for _, s := range []struct {
		a      []int64
		b      int64
		expect []int64
	}{
		{[]int64{1, 2}, 2, []int64{1, 4, 4}},
		{[]int64{1, 1}, 3, []int64{1, 3, 3, 1}},
		{[]int64{1, 1}, 4, []int64{1, 4, 6, 4, 1}},
		{[]int64{0, 1}, 3, []int64{0, 0, 0, 1}},
		{[]int64{0, 2}, 3, []int64{0, 0, 0, 8}},
	} {
		a := NewPolyInts(lv, s.a...)
		b := NewInt(s.b)
		ep := NewPolyInts(lv, s.expect...)
		c := a.Pow(b)
		if !c.Equals(ep) {
			t.Errorf("invalid poly.pow a=%v, b=%v, exp=%v, actual=%v", a, b, ep, c)
		}

		c = a.Pow(zero)
		if !c.Equals(one) {
			t.Errorf("invalid poly.pow a=%v, b=0, exp=1, actual=%v", a, c)
		}
	}
}
