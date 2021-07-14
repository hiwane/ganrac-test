package ganrac

import (
	"fmt"
	"testing"
)

func TestDescartes(t *testing.T) {
	for _, s := range []struct {
		p      []int64
		np, nn int
	}{
		{[]int64{-2, 0, 1}, 1, 1},
		{[]int64{4, 12, 21, 20, 9, -3, -7, -3, 0, 1}, 2, 7}, // (x+1)(x-2)^2(x^2+x+1)^3 QEbook 115
		{[]int64{3, 2, -1, -3, -2, 1}, 2, 3},                // QEbook 125
	} {
		lv := Level(0)
		p := NewPolyInts(lv, s.p...)
		np := p.descartesSignRules()
		if np != s.np {
			t.Errorf("+p(+x)=%v, expect=%d, actual=%d", s.p, s.np, np)
		}

		q := p.Neg().(*Poly)
		np = q.descartesSignRules()
		if np != s.np {
			t.Errorf("-p(+x)=%v, expect=%d, actual=%d", s.p, s.np, np)
		}

		q = p.Subst(NewPolyCoef(lv, 0, -1), lv).(*Poly)
		nn := q.descartesSignRules()
		if nn != s.nn {
			t.Errorf("+p(-x)=%v, expect=%d, actual=%d", s.p, s.nn, nn)
		}

		q = q.Neg().(*Poly)
		nn = q.descartesSignRules()
		if nn != s.nn {
			t.Errorf("-p(-x)=%v, expect=%d, actual=%d", s.p, s.nn, nn)
		}
	}
}

func TestConvertRange(t *testing.T) {
	for _, s := range []struct {
		n         int64
		m         int
		p, expect []int64
	}{
		//		{1, 4, []int64{-6, 25, -27, 0, 4}, []int64{686, 245, -168, -55, -4}},
		{3, 1, []int64{-6, 25, -27, 0, 4}, []int64{14850, 43830, 48426, 23738, 4356}}, // 6..8
		{0, 2, []int64{-6, 25, -27, 0, 4}, []int64{686, -588, -168, 76, -6}},          // 0..4
	} {
		lv := Level(0)
		p := NewPolyInts(lv, s.p...)
		expect := NewPolyInts(lv, s.expect...)

		q := p.convertRange(newBinIntInt64(s.n, s.m))
		if !expect.Equals(q) {
			t.Errorf("\ninput=%v\nexpect=%v\nactual=%v\n", p, expect, q)
		}
	}
}

func TestRealRoot(t *testing.T) {
	for _, s := range []struct {
		p []int64
		n int // # of root
	}{
		// 無平方と仮定
		{[]int64{720, -1764, 1624, -735, 175, -21, 1}, 6}, // x=1,2,3,4,5,6
		{[]int64{1, 3, 2}, 2},                             // CA p39
		{[]int64{-2, 0, 1}, 2},
		{[]int64{0, -2, 0, 1}, 3},
		{[]int64{+3, -4, 1}, 2},
		{[]int64{-10, 31, 37, -124, 12}, 4},
		{[]int64{-2, 1, -2, 1}, 1},      // CA p182
		{[]int64{0, +6, -7, 0, 1}, 4},   // 0,1,2 -3
		{[]int64{0, -110, -1, 1}, 3},    // 0,11,-10
		{[]int64{0, 6, 0, -5, 0, 1}, 5}, // 0,+-sqrt(2),+-sqrt(3)
	} {
		lv := Level(0)
		pp := NewPolyInts(lv, s.p...)
		qq := pp.subsXinv()
		if s.p[0] == 0 {
			qq = qq.Mul(NewPolyVar(lv)).(*Poly)
		}
		for _, p := range []*Poly{
			pp,
			pp.Neg().(*Poly),
			pp.Subst(NewPolyCoef(lv, 0, -1), lv).(*Poly),
			qq,
			qq.Neg().(*Poly),
			qq.Subst(NewPolyCoef(lv, 0, -1), lv).(*Poly),
		} {
			r, err := p.RealRootIsolation(10)
			if err != nil {
				t.Errorf("err %v\ninput=%p", err, p)
				continue
			}
			if r.Len() != s.n {
				t.Errorf("# of root: expect=%d, actual=%d\ninput=%v\nret=%v\n", s.n, r.Len(), p, r)
				continue
			}

			for i, intv := range r.v {
				intvl, ok := intv.(*List)
				if !ok {
					t.Errorf("[%d] not a list: actual=%d", i, intv)
				}

				left, ok := intvl.v[0].(NObj)
				if !ok {
					t.Errorf("[%d] left is not a numeric: actual=%d", i, intvl.v[0])
				}

				right, ok := intvl.v[1].(NObj)
				if !ok {
					t.Errorf("[%d] right is not a numeric: right actual=%d", i, intvl.v[1])
				}

				sgn_l := p.Subst(left, lv).Sign()
				sgn_u := p.Subst(right, lv).Sign()
				if sgn_l == 0 && sgn_u == 0 {
					if !left.Equals(right) {
						t.Errorf("[%d] p(l)=p(r)=0 but l != r: %v, %v", i, left, right)
						continue
					}
				} else if sgn_l == 0 || sgn_u == 0 {
					t.Errorf("[%d] p(l%d)p(r%d)=0: %v, %v", i, sgn_l, sgn_u, left, right)
					continue
				} else if sgn_l == sgn_u {
					fmt.Printf("intv=%v, left=%v, ri=%v\n", intv, left, right)
					t.Errorf("[%d] not an iso. intv.: f(%e)=%d, f(%e)=%d\np=%v\nintv=%v", i, left.Float(), sgn_l, right.Float(), sgn_u, p, intv)
					continue
				}

				if left.Cmp(right) > 0 {
					t.Errorf("[%d] invalid intv. [%e, %e] %v", i, left.Float(), right.Float(), p)
				}
			}
			// @TODO 区間がソートされていること, 重複がないこと
			for i := 1; i < s.n; i++ {
				l1 := r.v[i-1].(*List).v[0].(NObj)
				r1 := r.v[i-1].(*List).v[1].(NObj)

				l2 := r.v[i-0].(*List).v[0].(NObj)
				r2 := r.v[i-0].(*List).v[1].(NObj)

				if r1.Cmp(l2) >= 0 {
					t.Errorf("[%d] overlap. [%e,%e],[%e,%e]", i, l1.Float(), r1.Float(), l2.Float(), r2.Float())
				}
			}
		}
	}
}
