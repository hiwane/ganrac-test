package ganrac

import (
	"testing"
)

func TestOpRObj2(t *testing.T) {
	for i, s := range []struct {
		f       func(x, y RObj) RObj
		op      string
		a, b, o RObj
	}{
		{Sub, "-", NewInt(10), NewRatInt64(1, 3), NewRatInt64(29, 3)},
		{Sub, "-", NewRatInt64(4, 3), NewRatInt64(1, 3), NewInt(1)},
		{Sub, "-", NewRatInt64(1, 2), NewRatInt64(1, 3), NewRatInt64(1, 6)},
		{Sub, "-", NewInt(10), NewInt(3), NewInt(7)},
		{Sub, "-", NewPolyInts(0, 5, 7), NewInt(3), NewPolyInts(0, 2, 7)},
		{Sub, "-", NewInt(3), NewPolyInts(0, 5, 7), NewPolyInts(0, -2, -7)},
		{Sub, "-", NewPolyInts(0, 2, 3, 11), NewPolyInts(0, 5, 7), NewPolyInts(0, -3, -4, 11)},

		{Add, "+", NewInt(10), NewInt(3), NewInt(13)},
		{Add, "+", NewPolyInts(0, 5, 7), NewInt(3), NewPolyInts(0, 8, 7)},
		{Add, "+", NewPolyInts(0, 2, 3, 11), NewPolyInts(0, 5, 7), NewPolyInts(0, 7, 10, 11)},
		{Add, "+", NewInt(10), NewRatInt64(1, 3), NewRatInt64(31, 3)},
		{Add, "+", NewRatInt64(4, 3), NewRatInt64(2, 3), NewInt(2)},
		{Add, "+", NewRatInt64(-2, 3), NewRatInt64(2, 3), zero},

		{Mul, "*", NewInt(10), NewInt(3), NewInt(30)},
		{Mul, "*", NewPolyInts(0, 5, 7), NewInt(3), NewPolyInts(0, 15, 21)},
		{Mul, "*", NewPolyInts(0, 2, 3, 11), NewPolyInts(0, 5, 7), NewPolyInts(0, 10, 29, 76, 77)},
		{Mul, "*", NewInt(10), NewRatInt64(1, 3), NewRatInt64(10, 3)},
		{Mul, "*", NewRatInt64(4, 3), NewRatInt64(2, 3), NewRatInt64(8, 9)},
		{Mul, "*", NewRatInt64(4, 3), NewRatInt64(3, 2), NewInt(2)},
	} {
		act := s.f(s.a, s.b)
		if !act.Equals(s.o) {
			t.Errorf("[%2d,%s] a=%v, b=%v, expect=%v, actual=%v", i, s.op, s.a, s.b, s.o, act)
		}

		if s.op == "-" {
			continue
		}

		// 可換
		act = s.f(s.b, s.a)
		if !act.Equals(s.o) {
			t.Errorf("[%2d,%s] b=%v, a=%v, expect=%v, actual=%v", i, s.op, s.b, s.a, s.o, act)
		}
	}
}
