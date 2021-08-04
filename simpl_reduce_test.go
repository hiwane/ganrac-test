package ganrac

import (
	"fmt"
	"testing"
)

func TestSimplReduce(t *testing.T) {
	g := NewGANRAC()
	g.verbose = 0
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		fmt.Printf("skip TestSimplReduce... (no ox)\n")
		return
	}
	defer connc.Close()
	defer connd.Close()

	var opt QEopt
	opt.Algo = 0

	x := NewPolyVar(0)
	y := NewPolyVar(1)
	z := NewPolyVar(2)

	for ii, ss := range []struct {
		input  Fof
		expect Fof // simplified input
	}{
		{
			newFmlAnds(NewAtom(x, EQ), NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, x), 1), GT)), // x==0 && z+x*y>0
			newFmlAnds(NewAtom(x, EQ), NewAtom(z, GT)),                                       // x==0 && z>0
		}, {
			newFmlAnds(NewAtom(y, EQ), NewAtom(NewPolyCoef(1, 1, 1), GT)), // y==0 && y+1 > 0
			newFmlAnds(NewAtom(y, EQ)),                                    // y==0
		}, {
			newFmlAnds(
				NewAtom(NewPolyCoef(1, -1, 0, 2), EQ),
				NewAtom(NewPolyCoef(1, 1, 1), EQ)), // 2*y^2+1=0 && y+1 = 0
			falseObj,
		}, {
			newFmlAnds(NewAtom(NewPolyCoef(0, 0, 1), EQ),
				NewQuantifier(false, []Level{0}, NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, NewPolyCoef(0, 0, 1)), 1), EQ))), // x==0 && ex([x], z+x*y==0)
			nil,
		}, {
			newFmlAnds(NewAtom(NewPolyCoef(0, 0, 1), EQ), NewQuantifier(false, []Level{0}, NewAtom(NewPolyCoef(0, 1, 1), GT))),
			NewAtom(NewPolyCoef(0, 0, 1), EQ),
		}, {
			newFmlAnds( // x==0 && z==0 && ex([x], z+y+x==0)
				NewAtom(NewPolyCoef(0, 0, 1), EQ),
				NewAtom(NewPolyCoef(2, 2, 1), EQ),
				NewQuantifier(false, []Level{0},
					NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 1), 1), 1), EQ))),
			newFmlAnds(
				NewAtom(NewPolyCoef(0, 0, 1), EQ),
				NewAtom(NewPolyCoef(2, 2, 1), EQ),
				NewQuantifier(false, []Level{0},
					NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, 1), 1), EQ))),
		}, {
			// x==0 && ex([w], x*w^2+y*w+1==0 && y*w^3+z*w+x==0 && w-1<=0)
			newFmlAnds(
				NewAtom(NewPolyCoef(0, 0, 1), EQ),
				NewQuantifier(false, []Level{3}, newFmlAnds(
					NewAtom(NewPolyCoef(3, 1, NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 1)), EQ),
					NewAtom(NewPolyCoef(3, NewPolyCoef(0, 0, 1), NewPolyCoef(2, 0, 1), 0, NewPolyCoef(1, 0, 1)), EQ),
					NewAtom(NewPolyCoef(3, -1, 1), LE)))),
			// x==0 && ex([w], y*w+1==0 && y*w^3+z*w==0 && w-1<=0)
			newFmlAnds(
				NewAtom(NewPolyCoef(0, 0, 1), EQ),
				NewQuantifier(false, []Level{3}, newFmlAnds(
					NewAtom(NewPolyCoef(3, 1, NewPolyCoef(1, 0, 1)), EQ),
					NewAtom(NewPolyCoef(3, 0, NewPolyCoef(2, 0, 1), 0, NewPolyCoef(1, 0, 1)), EQ),
					NewAtom(NewPolyCoef(3, -1, 1), LE)))),
		},
	} {
		if ss.expect == nil {
			ss.expect = ss.input
		}
		for jj, s := range []struct {
			input  Fof
			expect Fof
		}{
			{ss.input, ss.expect},
			{ss.input.Not(), ss.expect.Not()},
		} {
			// fmt.Printf("[%d,%d] s=%v\n", ii, jj, s.input)
			inf := newReduceInfo()
			o := s.input.simplReduce(g, inf)
			if testSameFormAndOr(o, s.expect) {
				continue
			}

			u := newFmlEquiv(o, s.expect)
			switch uqe := g.QE(u, opt).(type) {
			case *AtomT:
				continue
			default:
				t.Errorf("<%d,%d>\n input=%v\nexpect=%v\noutput=%v\ncmp=%v", ii, jj, s.input, s.expect, o, uqe)
			}
		}
	}
}
