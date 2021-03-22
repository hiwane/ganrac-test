package ganrac

import (
	"testing"
)

func TestRatOp2(t *testing.T) {
	for _, s := range []struct {
		a1, a2, b1, b2     int64
		add, sub, mul, div RObj
	}{
		{1, 2, 4, 3,
			NewRatInt64(11, 6), NewRatInt64(3-8, 6), NewRatInt64(2, 3), NewRatInt64(3, 8)},
		{4, 3, 2, 3,
			NewInt(2), NewRatInt64(2, 3), NewRatInt64(8, 9), NewInt(2)},
		{4, 3, 1, 3,
			NewRatInt64(5, 3), NewInt(1), NewRatInt64(4, 9), NewInt(4)},
	} {
		a := NewRatInt64(s.a1, s.a2)
		b := NewRatInt64(s.b1, s.b2)

		c := a.Add(b)
		if !c.Equals(s.add) {
			t.Errorf("invalid add a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.add, c)
		}

		c = b.Add(a)
		if !c.Equals(s.add) {
			t.Errorf("invalid add b=%v, a=%v, expect=%v, actual=%v\n", b, a, s.add, c)
		}

		c = a.Sub(b)
		if !c.Equals(s.sub) {
			t.Errorf("invalid sub a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.sub, c)
		}

		c = b.Sub(a)
		if !c.Equals(s.sub.Neg()) {
			t.Errorf("invalid sub b=%v, a=%v, expect=%v, actual=%v\n", b, a, s.sub, c)
		}

		c = a.Mul(b)
		if !c.Equals(s.mul) {
			t.Errorf("invalid mul a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.mul, c)
		}

		c = b.Mul(a)
		if !c.Equals(s.mul) {
			t.Errorf("invalid mul b=%v, a=%v, expect=%v, actual=%v\n", b, a, s.mul, c)
		}

		c = a.Div(b)
		if !c.Equals(s.div) {
			t.Errorf("invalid mul a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.div, c)
		}
	}
}

func TestRatIntOp2(t *testing.T) {
	for _, s := range []struct {
		a1, a2, b          int64
		add, sub, mul, div RObj
	}{
		{1, 3, 2,
			NewRatInt64(7, 3), NewRatInt64(-5, 3), NewRatInt64(2, 3), NewRatInt64(1, 6)},
		{4, 3, 2,
			NewRatInt64(10, 3), NewRatInt64(-2, 3), NewRatInt64(8, 3), NewRatInt64(2, 3)},
		{4, 3, 6,
			NewRatInt64(18+4, 3), NewRatInt64(4-18, 3), NewInt(8), NewRatInt64(2, 9)},
	} {
		a := NewRatInt64(s.a1, s.a2)
		b := NewInt(s.b)

		c := a.Add(b)
		if !c.Equals(s.add) {
			t.Errorf("invalid add a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.add, c)
		}

		c = b.Add(a)
		if !c.Equals(s.add) {
			t.Errorf("invalid add b=%v, a=%v, expect=%v, actual=%v\n", b, a, s.add, c)
		}

		c = a.Sub(b)
		if !c.Equals(s.sub) {
			t.Errorf("invalid sub a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.sub, c)
		}

		c = b.Sub(a)
		if !c.Equals(s.sub.Neg()) {
			t.Errorf("invalid sub b=%v, a=%v, expect=%v, actual=%v\n", b, a, s.sub, c)
		}

		c = a.Mul(b)
		if !c.Equals(s.mul) {
			t.Errorf("invalid mul a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.mul, c)
		}

		c = b.Mul(a)
		if !c.Equals(s.mul) {
			t.Errorf("invalid mul b=%v, a=%v, expect=%v, actual=%v\n", b, a, s.mul, c)
		}

		c = a.Div(b)
		if !c.Equals(s.div) {
			t.Errorf("invalid mul a=%v, b=%v, expect=%v, actual=%v\n", a, b, s.div, c)
		}
	}
}
