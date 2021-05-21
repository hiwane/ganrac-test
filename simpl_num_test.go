package ganrac

import (
	"fmt"
	"testing"
)

func TestSimplNum(t *testing.T) {

	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		return
	}
	defer connc.Close()
	defer connd.Close()

	for _, s := range []struct {
		a      Fof
		expect Fof
	}{
		{
			NewAtom(NewPolyCoef(0, 1, 0, 1), GT),
			trueObj,
		}, {
			NewAtom(NewPolyCoef(0, 1, 0, 1), LE),
			falseObj,
		}, {
			// (x-1)*(x-3)>=0 && x*(x-4) >= 0
			newFmlAnds(NewAtom(NewPolyCoef(0, 3, -4, 1), GE), NewAtom(NewPolyCoef(0, 0, -4, 1), GE)),
			// x*(x-4) >= 0
			NewAtom(NewPolyCoef(0, 0, -4, 1), GE),
		}, {
			// (x-2)^2 >= 10 && (x-1)^2 >= 2
			newFmlAnds(NewAtom(NewPolyCoef(0, -6, -4, 1), GE), NewAtom(NewPolyCoef(0, -1, -2, 1), GE)),
			// (x-2)^2 >= 10
			NewAtom(NewPolyCoef(0, -6, -4, 1), GE),
		}, {
			// x>2 && x^2+y^2 > 1
			newFmlAnds(NewAtom(NewPolyCoef(0, -2, 1), GT), NewAtom(NewPolyCoef(1, NewPolyCoef(0, -1, 0, 1), 0, 1), GT)),
			NewAtom(NewPolyCoef(0, -2, 1), GT), // x>2
		}, {
			// adam2-1 && projection
			// x>0 && y>0 && 4*y^2+4*x^2-1<=0 && y^2-y+x^2-x==2
			newFmlAnds(
				NewAtom(NewPolyCoef(0, 0, 1), GT),
				NewAtom(NewPolyCoef(1, 0, 1), GT),
				NewAtom(NewPolyCoef(1, NewPolyCoef(0, -1, 0, 4), 0, 4), LE), NewAtom(NewPolyCoef(1, NewPolyCoef(0, -2, -1, 1), -1, 1), EQ)),
			falseObj,
		}, {
			newFmlAnds(NewAtom(NewPolyCoef(1, -1, 1), GT), NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 0, 1), LT)),
			newFmlAnds(NewAtom(NewPolyCoef(1, -1, 1), GT), NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 0, 1), LT)),
		},
	} {
		fmt.Printf("===== in=%v\n", s.a)
		o, tf, ff := s.a.simplNum(g, nil, nil)
		if !o.Equals(s.expect) {
			t.Errorf("\ninput =%v\nexpect=%v\noutput=%v\nt=%v\nf=%v\n", s.a, s.expect, o, tf, ff)
			continue
		}
	}
}

func TestSimplNumUniPoly(test *testing.T) {
	lv := Level(0)
	for _, s := range []struct {
		a  *Poly
		t  []NObj
		f  []NObj
		op OP
	}{
		{
			NewPolyCoef(lv, 2, 0, -1),
			[]NObj{NewInt(0), nil},
			[]NObj{},
			OP_TRUE,
		},
	} {
		t := newNumRegion()
		for i := 0; i+1 < len(s.t); i += 2 {
			t.r[lv] = append(t.r[lv], &ninterval{s.t[i], s.t[i+1]})
		}
		f := newNumRegion()
		for i := 0; i+1 < len(s.f); i += 2 {
			f.r[lv] = append(f.r[lv], &ninterval{s.f[i], s.f[i+1]})
		}

		op, pos, neg := s.a.simplNumUniPoly(t, f)
		if op != s.op {
			test.Errorf("a=%v, t=%v, f=%v\nexpect=%v\nactual=%v\npos=%v\nneg=%v\n",
				s.a, t, f, s.op, op, pos, neg)
		}
	}
}
