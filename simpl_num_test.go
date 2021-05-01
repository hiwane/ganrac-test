package ganrac

import (
	"fmt"
	"testing"
)

func TestSimplNum(t *testing.T) {
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
		},
	} {

		fmt.Printf("===== in=%v\n", s.a)
		o, tf, ff := s.a.simplNum(nil, nil)
		if !o.Equals(s.expect) {
			t.Errorf("\ninput =%v\nexpect=%v\noutput=%v\nt=%v\nf=%v\n", s.a, s.expect, o, tf, ff)
			continue
		}
	}
}
