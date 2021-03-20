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
