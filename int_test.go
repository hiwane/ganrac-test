package ganrac

import (
	"testing"
)

func TestInts(t *testing.T) {
	v := NewInt(0)
	if v.Sign() != 0 || v.String() != "0" || v.IsOne() || v.IsMinusOne() || !v.IsZero() {
		t.Errorf("invalid int v=%v, sign=%d, str=%s\n", v, v.Sign(), v.String())
	}

	v = NewInt(1)
	if v.Sign() <= 0 || v.String() != "1" || !v.IsOne() || v.IsMinusOne() || v.IsZero() {
		t.Errorf("invalid int v=%v, sign=%d, str=%s\n", v, v.Sign(), v.String())
	}

	v = NewInt(-1)
	if v.Sign() >= 0 || v.String() != "-1" || v.IsOne() || !v.IsMinusOne() || v.IsZero() {
		t.Errorf("invalid int v=%v, sign=%d, str=%s\n", v, v.Sign(), v.String())
	}

	var g GObj
	g = v
	var r RObj
	r, _ = g.(RObj)
	_, ok := r.(*Int)
	if !ok {
		t.Errorf("invalid\n")
	}

	for _, s := range []int64{2, 5, 12345} {
		a := NewInt(s)
		an := NewInt(-s)

		b := a.Neg().(*Int)
		if !b.Equals(an) {
			t.Errorf("invalid neg a=%v, -a=%v\n", a, b)
		}

		c := b.Neg()
		if !c.Equals(a) {
			t.Errorf("invalid negneg a=%v, --a=%v\n", a, c)
		}

		c = b.Abs().(*Int)
		if !c.Equals(a) {
			t.Errorf("invalid  a=%v, |-a|=%v\n", a, c)
		}

		c = a.Abs().(*Int)
		if !c.Equals(a) {
			t.Errorf("invalid  a=%v, |+a|=%v\n", a, c)
		}
	}
}

func TestIntPow(t *testing.T) {
	for _, s := range []struct{ a, b, expect int64 }{
		{2, 0, 1},
		{2, 1, 2},
		{2, 2, 4},
		{2, 3, 8},
		{2, 4, 16},
		{2, 5, 32},
		{3, 0, 1},
		{3, 1, 3},
		{3, 2, 9},
		{3, 3, 27},
	} {
		a := NewInt(s.a)
		b := NewInt(s.b)
		c := a.Pow(b)
		expect := NewInt(s.expect)
		if !c.Equals(expect) {
			t.Errorf("invalid %d^%d expect=%d actual=%v", s.a, s.b, s.expect, c)
		}
	}
}

func TestIntOp2(t *testing.T) {
	for _, s := range []struct {
		a, b, pow, gcd int64
		div            NObj
	}{
		{2, 0, 1, 0, nil},
		{2, 1, 2, 1, NewInt(2)},
		{3, 4, 81, 1, NewRatInt64(3, 4)},
		{-3, 2, +9, 1, NewRatInt64(-3, 2)},
		{-3, 3, -27, 3, NewInt(-1)},
		{5, 4, 625, 1, NewRatInt64(5, 4)},
		{4, 2, 16, 2, NewInt(2)},
		{-4, 2, 16, 2, NewInt(-2)},
		{12, 8, -1, 4, NewRatInt64(3, 2)},
		{86400, 131040, -1, 1440, NewRatInt64(12*5, 13*7)},
	} {
		a := NewInt(s.a)
		b := NewInt(s.b)

		expect := NewInt(s.a + s.b)
		for i, c := range []RObj{
			a.Add(b), b.Add(a), a.Add(NewInt(s.b)), b.Add(NewInt(s.a)),
		} {
			if !c.Equals(expect) {
				t.Errorf("[%d] invalid %d+%d expect=%v actual=%v", i, s.a, s.b, expect, c)
			}
		}

		expect = NewInt(s.a * s.b)
		for i, c := range []struct {
			actual, expect RObj
			sgn            string
		}{
			{a.Sub(b), NewInt(s.a - s.b), ""},
			{b.Sub(a), NewInt(s.b - s.a), "-"},
		} {
			if !c.actual.Equals(c.expect) {
				t.Errorf("[%d] invalid %s(%d-%d) expect=%v actual=%v", i, c.sgn, s.a, s.b, c.expect, c.actual)
			}
		}

		expect = NewInt(s.a * s.b)
		for i, c := range []RObj{
			a.Mul(b), b.Mul(a),
		} {
			if !c.Equals(expect) {
				t.Errorf("[%d] invalid %d*%d expect=%v actual=%v", i, s.a, s.b, expect, c)
			}
		}

		expect = NewInt(s.pow)
		c := a.Pow(b)
		if s.pow >= 0 && !c.Equals(expect) {
			t.Errorf("invalid %d^%d expect=%v actual=%v", s.a, s.b, expect, c)
		}

		if s.b != 0 {
			c = a.Div(b)
			if !c.Equals(s.div) {
				t.Errorf("invalid %d/%d expect=%v actual=%v", s.a, s.b, s.div, c)
			}
		}

		expect = NewInt(s.gcd)
		c = a.Gcd(b)
		if !c.Equals(expect) {
			t.Errorf("invalid gcd(%d,%d) expect=%v actual=%v", s.a, s.b, expect, c)
		}
	}
}

func TestIntGcd(t *testing.T) {
	for _, s := range []struct {
		a, b, expect string
	}{
		{"2790221028", "65587796069", "2297"},
		{"1155036", "2448444", "12"},
		{"1221773556", "576362964", "5988"},
	} {
		a := ParseInt(s.a, 10)
		b := ParseInt(s.b, 10)
		expect := ParseInt(s.expect, 10)

		g := a.Gcd(b)
		if !g.Equals(expect) {
			t.Errorf("invalid gcd(%v,%v) expect=%v actual=%v", s.a, s.b, expect, g)
		}
	}
}

func TestIntGcdEx(t *testing.T) {
	for _, s := range []struct {
		a, b, expect int64
	}{
		{12 * 151, 12 * 157, 12},
		{983, 991, 1},
		{983, 673, 1},
		{983, 991 * 673, 1},
		{2 * 3 * 5 * 11 * 983, 191 * 991 * 673, 1},
	} {
		a := NewInt(s.a)
		b := NewInt(s.b)
		expect := NewInt(s.expect)

		g := a.Gcd(b)
		if !g.Equals(expect) {
			t.Errorf("invalid gcd(%v,%v) expect=%v actual=%v", s.a, s.b, expect, g)
			continue
		}

		g2, s2, t2 := a.GcdEx(b)
		if !g2.Equals(expect) {
			t.Errorf("invalid gcdEx(%v,%v) expect=%v actual=%v", s.a, s.b, expect, g2)
			continue
		}

		sa := s2.Mul(a)
		tb := t2.Mul(b)
		satb := sa.Add(tb)
		if !g2.Equals(satb) {
			t.Errorf("invalid gcdEx(%v,%v)=(%v,%v,%v)", s.a, s.b, g2, s2, t2)
			continue
		}
	}
}
