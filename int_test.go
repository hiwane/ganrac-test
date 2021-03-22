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
		a, b, pow int64
		div       NObj
	}{
		{2, 0, 1, nil},
		{2, 1, 2, NewInt(2)},
		{3, 4, 81, NewRatInt64(3, 4)},
		{-3, 2, +9, NewRatInt64(-3, 2)},
		{-3, 3, -27, NewInt(-1)},
		{5, 4, 625, NewRatInt64(5, 4)},
		{4, 2, 16, NewInt(2)},
		{-4, 2, 16, NewInt(-2)},
	} {
		a := NewInt(s.a)
		b := NewInt(s.b)

		expect := NewInt(s.a + s.b)
		for i, c := range []RObj{
			a.Add(b), b.Add(a), a.AddInt(s.b), b.AddInt(s.a),
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
		if !c.Equals(expect) {
			t.Errorf("invalid %d^%d expect=%v actual=%v", s.a, s.b, expect, c)
		}

		if s.b != 0 {
			c = a.Div(b)
			if !c.Equals(s.div) {
				t.Errorf("invalid %d/%d expect=%v actual=%v", s.a, s.b, s.div, c)
			}
		}

	}
}
