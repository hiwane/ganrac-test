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
