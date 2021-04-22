package ganrac

import (
	"fmt"
	"testing"
)

func TestBinIntBase(t *testing.T) {
	a := newBinInt()
	a.n.SetInt64(-4)
	a.m = -2

	if s := a.Sign(); s >= 0 {
		t.Errorf("a=%v,s=%v", a, s)
	}
	if s := a.IsZero(); s {
		t.Errorf("a=%v,s=%v", a, s)
	}
	if s := a.IsOne(); s {
		t.Errorf("a=%v,s=%v", a, s)
	}
	if s := a.IsMinusOne(); !s {
		t.Errorf("a=%v,s=%v", a, s)
	}
	fmt.Printf("len=%d\n", a.n.BitLen())

	b := newBinInt()
	b.n.SetInt64(+8)
	b.m = -3

	if s := b.Sign(); s <= 0 {
		t.Errorf("input=%v,s=%v", b, s)
	}
	if s := b.IsZero(); s {
		t.Errorf("input=%v,s=%v", b, s)
	}
	if s := b.IsOne(); !s {
		t.Errorf("input=%v,s=%v", b, s)
	}
	if s := b.IsMinusOne(); s {
		t.Errorf("input=%v,s=%v", b, s)
	}

	if s := a.Add(b); s.Sign() != 0 {
		t.Errorf("input=`%v` + `%v`,s=%v", a, b, s)
	}
	if s := b.Add(a); s.Sign() != 0 {
		t.Errorf("input=`%v` + `%v`,s=%v", a, b, s)
	}

	if s := b.Neg(); !a.Equals(s) || !s.Equals(a) || s.Sign() >= 0 {
		t.Errorf("s=%v", s)
	}
	if s := a.Neg(); !s.Equals(b) || !b.Equals(s) || s.Sign() <= 0 {
		t.Errorf("s=%v", s)
	}
	// if s := a.CmpAbs(b); s != 0 {
	// 	t.Errorf("s=%v", s)
	// }
	if s := a.Cmp(b); s >= 0 {
		t.Errorf("s=%v", s)
	}
	if s := b.Cmp(a); s <= 0 {
		t.Errorf("s=%v", s)
	}

	// 壊れていないよね
	if s := a.IsMinusOne(); !s {
		t.Errorf("a=%v,s=%v", a, s)
	}
	if s := b.IsOne(); !s {
		t.Errorf("input=%v,s=%v", b, s)
	}
}

func TestBinIntSubst(t *testing.T) {
	NewGANRAC()
	for _, s := range []struct {
		input  *Poly
		num    int64
		den    int
		lv     Level
		expect RObj
	}{
		{
			NewPolyCoef(0, 2, 3, 4, 5),
			7, 0, 0,
			NewInt(1934),
		}, {
			NewPolyCoef(1,
				NewInt(5),
				NewPolyCoef(0, 2, 3),
				NewPolyCoef(0, 5, 7, -3)),
			7, -2, 0,
			NewPolyCoef(1, 80, 116, 129),
		}, {
			NewPolyCoef(0, 2, 3),
			7, -2, 0,
			NewInt(29),
		}, {
			NewPolyCoef(0, 2, 3, 4, 5),
			7, -2, 0,
			NewInt(2963),
		}, {
			NewPolyCoef(1,
				NewInt(5),
				NewPolyCoef(0, 2, 3),
				NewPolyCoef(0, 5, 7, -3)),
			7, -2, 0,
			NewPolyCoef(1, 80, 116, 129),
		}, {
			NewPolyCoef(2,
				NewInt(5),
				NewPolyCoef(1, 2, 3, 4, 5, 6),
				NewPolyCoef(0, -2, 7, 11),
				NewPolyCoef(1, NewPolyCoef(0, -5, 3, 2),
					NewInt(-13),
					NewPolyCoef(0, 3, 1))),
			7, -2, 1,
			NewPolyCoef(2,
				NewInt(1280),
				NewInt(26258),
				NewPolyCoef(0, -512, 1792, 2816),
				NewPolyCoef(0, -4752, 1552, 512)),
		},
	} {
		b := newBinIntInt64(s.num, s.den)
		o := b.subst_poly(s.input, s.lv)
		if !o.Equals(s.expect) {
			t.Errorf("lv=%d\ninput =%v\nx=%v*2^(%d) => `%v`\nexpect=%v\noutput=%v", s.lv, s.input, s.num, s.den, b, s.expect, o)
			return
		}

		q := b.ToIntRat()
		o = q.subst_poly(s.input, s.lv)
		if !o.Equals(s.expect) {
			t.Errorf("lv=%d\ninput =%v\nx=%v*2^(%d)\nexpect=%v\noutput=%v", s.lv, s.input, s.num, s.den, s.expect, o)
			continue
		}

	}
}
