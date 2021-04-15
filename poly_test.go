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
		{[]int64{2, 1}, []int64{-2, -1}, zero},
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
	ep = NewPolyInts(1, -1, 6, 7)
	ep.c[0] = NewPolyInts(0, 7, 3, 4)

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

	ep = NewPoly(1, 2) //   5*(11y+7)x + 3*(11y+7)
	ep.c[0] = NewPolyInts(0, 21, 35, 42)
	ep.c[1] = NewPolyInts(0, 33, 55, 66)

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

	b = zero
	c = a.Mul(b)
	if !b.Equals(c) {
		t.Errorf("invalid poly.mul a=%v, b=%v, expect=%v, actual=%v", a, b, b, c)
	}
}

func TestPolyPow(t *testing.T) {
	lv := Level(0)
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

func TestPolySubst(t *testing.T) {
	lv := Level(0)
	for _, s := range []struct {
		a      []int64
		b      int64
		expect int64
	}{
		{[]int64{1, 2}, 3, 7},
		{[]int64{1, 2}, 0, 1},
		{[]int64{1, 2}, 1, 3},
		{[]int64{1, 2, 5}, 1, 8},
		{[]int64{1, 2, 5}, 3, 52},
	} {
		a := NewPolyInts(lv, s.a...)
		b := NewInt(s.b)
		ep := NewInt(s.expect)
		c := a.Subst([]RObj{b}, []Level{lv}, 0)
		if !c.Equals(ep) {
			t.Errorf("invalid poly.subst a=%v, b=%v, exp=%v, actual=%v", a, b, ep, c)
		}
	}
}

func TestHasSameTerm(t *testing.T) {
	for _, s := range []struct {
		a      *Poly
		b      *Poly
		expect bool
	}{
		{
			NewPolyInts(0, 1, 2, 3, 0, 5),
			NewPolyInts(0, 1, 5, 8, 0, -3),
			true},
		{
			NewPolyInts(0, 1, 2, 3, 0, 5),
			NewPolyInts(1, 1, 2, 3, 0, 5),
			false},
		{
			NewPolyCoef(0, one, zero, two, one, NewPolyInts(1, 1, 1, 1)),
			NewPolyCoef(0, two, zero, two, one, NewPolyInts(1, 1, 1, 1)),
			true},
		{
			NewPolyCoef(0, one, zero, two, one, NewPolyInts(1, 1, 0, 1)),
			NewPolyCoef(0, two, zero, two, one, NewPolyInts(1, 1, 1, 1)),
			false},
	} {
		c := s.a.hasSameTerm(s.b, true)
		if c != s.expect {
			t.Errorf("a=%v, b=%v, expect=%v, actual=%v", s.a, s.b, c, s.expect)
		}

		c = s.b.hasSameTerm(s.a, true)
		if c != s.expect {
			t.Errorf("a=%v, b=%v, expect=%v, actual=%v", s.a, s.b, c, s.expect)
		}

		an := s.a.Neg().(*Poly)
		c = an.hasSameTerm(s.b, true)
		if c != s.expect {
			t.Errorf("a=%v, -a=%v, expect=%v, actual=%v", s.a, an, c, s.expect)
		}

		an = s.a.Mul(two).(*Poly)
		c = an.hasSameTerm(s.b, true)
		if c != s.expect {
			t.Errorf("a=%v, -2a=%v, expect=%v, actual=%v", s.a, an, c, s.expect)
		}

		bn := s.b.Neg().(*Poly)
		c = bn.hasSameTerm(s.a, true)
		if c != s.expect {
			t.Errorf("-b=%v, a=%v, expect=%v, actual=%v", bn, s.a, c, s.expect)
		}

		// 自身は true
		c = s.a.hasSameTerm(s.a, true)
		if !c {
			t.Errorf("a=%v, expect=%v, actual=true", s.a, c)
		}
		c = s.b.hasSameTerm(s.b, true)
		if !c {
			t.Errorf("b=%v, expect=%v, actual=true", s.b, c)
		}
	}
}

func TestSubstFrac(t *testing.T) {
	for _, s := range []struct {
		p        *Poly
		lv       Level
		num, den RObj
		expect   RObj
	}{
		{
			NewPolyInts(0, -11, 13),
			0,
			NewInt(5), NewInt(7),
			NewInt(-12),
		}, {
			NewPolyInts(0, 2, 3, 1),
			0,
			NewInt(5), NewInt(7),
			NewInt(228),
		}, {
			NewPolyCoef(2,
				NewPolyInts(1, 0, 3),
				NewPolyInts(1, -7, 5, -3),
				NewPolyCoef(1, NewInt(5), NewPolyInts(0, 1, 2, 3, 4, 5))),
			1,
			NewInt(5), NewInt(7),
			NewPolyCoef(2,
				NewInt(105),
				NewInt(-243),
				NewPolyInts(0, 280, 70, 105, 140, 175)),
		},
	} {
		d := s.p.Deg(s.lv)

		// prepare
		dens := make([]RObj, d+1)
		dens[0] = one
		dens[1] = s.den
		for i := 2; i <= d; i++ {
			dens[i] = dens[i-1].Mul(s.den)
		}

		actual := s.p.subst_frac(s.num, dens, s.lv)
		if !actual.Equals(s.expect) {
			t.Errorf("p=%v, x=(%v)/(%v), expect=%v, actual=%v\n", s.p, s.num, s.den, s.expect, actual)
		}

		dens = append(dens, dens[len(dens)-1].Mul(s.den))
		expect := s.expect.Mul(s.den)

		actual = s.p.subst_frac(s.num, dens, s.lv)
		if !actual.Equals(expect) {
			t.Errorf("<1> p=%v, x=(%v)/(%v), expect=%v, actual=%v\n", s.p, s.num, s.den, expect, actual)
		}

		dens = append(dens, dens[len(dens)-1].Mul(s.den))
		expect = expect.Mul(s.den)

		actual = s.p.subst_frac(s.num, dens, s.lv)
		if !actual.Equals(expect) {
			t.Errorf("<2> p=%v, x=(%v)/(%v), expect=%v, actual=%v\n", s.p, s.num, s.den, expect, actual)
		}

	}
}

func TestPolyDiff(t *testing.T) {
	for _, s := range []struct {
		p      *Poly
		lv     Level
		expect RObj
	}{
		{
			NewPolyInts(0, -11, 13),
			0,
			NewInt(13),
		}, {
			NewPolyInts(0, 2, 3, 1),
			0,
			NewPolyInts(0, 3, 2),
		}, {
			NewPolyCoef(1,
				NewPolyInts(0, 2, 3, 4),
				NewPolyInts(0, -3, -5, -6),
				NewPolyInts(0, -2, 11)),
			1,
			NewPolyCoef(1,
				NewPolyInts(0, -3, -5, -6),
				NewPolyInts(0, -4, 22)),
		}, {
			NewPolyCoef(1,
				NewPolyInts(0, 0, 1),
				NewInt(1)),
			0,
			NewInt(1),
		},
	} {
		c := s.p.diff(s.lv)
		if err := c.valid(); err != nil {
			t.Errorf("f[%d]=%v, actual=%v: %v", s.lv, s.p, c, err)
		}

		if !c.Equals(s.expect) {
			t.Errorf("f[%d]=%v, expect=%v, actual=%v", s.lv, s.p, s.expect, c)
		}
	}
}

func TestSubstBinint1Var(t *testing.T) {
	lv := Level(0)
	for _, s := range []struct {
		numer  int64
		denom  uint
		p      *Poly
		expect *Poly
	}{
		{5, 0, NewPolyInts(lv, -5, -3, 2), NewPolyInts(lv, 30, 17, 2)},
		{5, 2, NewPolyInts(lv, -5, -3, 2), NewPolyInts(lv, -90, 32, 32)},
	} {
		c := s.p.subst_binint_1var(NewInt(s.numer), s.denom)
		if !c.Equals(s.expect) {
			t.Errorf("\ninput =%v\nexpect=%v\nactual=%v\n", s.p, s.expect, c)
		}
	}

	for _, s := range []struct {
		numer  int64
		denom  int
		p      *Poly
		expect RObj
	}{
		{-1, -1, NewPolyInts(lv, 1, 3, 2), zero},
		{-1, -1, NewPolyInts(lv, -1, -3, -2), zero},
		{-4, -3, NewPolyInts(lv, -1, -3, -2), zero},
	} {
		c := s.p.subst1(newBinIntInt64(s.numer, s.denom), lv)
		if !c.Equals(s.expect) {
			t.Errorf("subst2: %d*2^(%d)\ninput =%v\nexpect=%v\nactual=%v\n", s.numer, s.denom, s.p, s.expect, c)
		}
	}
}

func TestSdiv(t *testing.T) {
	for _, s := range []struct {
		x, y   *Poly
		expect RObj
	}{
		{
			NewPolyInts(0, 6, 11, 6, 1), // x=z*y
			NewPolyInts(0, 3, 1),        // y
			NewPolyInts(0, 2, 3, 1),     // z
		}, {
			NewPolyInts(0, 6, 9, 3), // x=z*y
			NewPolyInts(0, 2, 3, 1), // y
			NewInt(3),
		}, {
			NewPolyCoef(2, NewPolyVar(0), NewPolyCoef(1, zero, NewPolyVar(0))), // x*y*z+x
			NewPolyCoef(2, one, NewPolyVar(1)),                                 // (y*z+1)
			NewPolyVar(0),
		}, {
			NewPolyCoef(2, NewPolyVar(1), NewPolyCoef(1, zero, NewPolyVar(0))), // x*y*z+y
			NewPolyCoef(2, one, NewPolyVar(0)),                                 // (x*z+1)
			NewPolyVar(1),
		},
	} {
		q := s.x.sdiv(s.y)
		if q == nil || !q.Equals(s.expect) {
			t.Errorf("\ninputx=%v\ninputy=%v\nexpect=%v\noutput=%v", s.x, s.y, s.expect, q)
			continue
		}

		if qqq, ok := s.expect.(*Poly); ok && qqq.lv == s.x.lv {
			q = s.x.sdiv(s.expect.(*Poly))
			if q == nil || !q.Equals(s.y) {
				t.Errorf("\ninputx=%v\ninputy=%v\nexpect=%v\noutput=%v", s.x, s.y, s.expect, q)
				continue
			}
		}

		q = s.y.Mul(s.expect)
		if q == nil || !q.Equals(s.x) {
			t.Errorf("\ninputx=%v\ninputy=%v\nexpect=%v\noutput=%v", s.x, s.y, s.expect, q)
			continue
		}

	}

}

func TestPolMul2Exp(t *testing.T) {
	s := NewPolyCoef(0,
		NewInt(7),
		NewRatInt64(3, 5),
		newBinIntInt64(3, -10),
		NewInt(13),
	)

	var p, q RObj
	p = s.mul_2exp(0)
	q = s
	if !p.Equals(q) {
		t.Errorf("m=1\ns=%v\np=%v\nq=%v\n", s, p, q)
	}

	p = s.mul_2exp(1)
	q = s.Mul(NewInt(2))
	if !p.Equals(q) {
		t.Errorf("m=2\ns=%v\np=%v\nq=%v\n", s, p, q)
	}

	p = s.mul_2exp(3)
	q = s.Mul(NewInt(8))

	if !p.Equals(q) {
		t.Errorf("m=8\ns=%v\np=%v\nq=%v\n", s, p, q)
	}
}

func TestPolReduce(t *testing.T) {
	for _, s := range []struct {
		x, y   *Poly
		expect RObj
	}{
		{
			NewPolyCoef(1, NewPolyInts(0, -2, 0, 3), NewPolyInts(0, -5, 1, 3)),
			NewPolyInts(0, -2, 0, 3),
			NewPolyCoef(1, zero, NewPolyInts(0, -3, 1)),
		}, {
			NewPolyCoef(1, NewPolyInts(0, -5, 1, 3), NewPolyInts(0, -2, 0, 3)),
			NewPolyInts(0, -2, 0, 3),
			NewPolyInts(0, -3, 1),
		}, {
			NewPolyCoef(1, NewPolyInts(0, 0, 0, 2), NewPolyInts(0, 1, 0, 1)),
			NewPolyInts(0, -2, 0, 3),
			NewPolyInts(1, 4, 5),
		},
	} {
		o := s.x.reduce(s.y)
		if !o.Equals(s.expect) {
			t.Errorf("\nx=%v\ny=%v\nexpect=%v\nactual=%v", s.x, s.y, s.expect, o)
		}
	}
}
