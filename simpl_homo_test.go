package ganrac

import (
	"fmt"
	"testing"
)

func TestHomoSolve(t *testing.T) {
	var opt QEopt
	for _, s := range []struct {
		a [][]int
		o bool
		x []int
	}{
		{
			[][]int{
				{1, 0, -1, 2, 0},
				{1, -1, 0, 1, 0},
				{0, 0, 0, 0, 0},
			}, true,
			[]int{1, 1, 1, 0, -1},
		},
	} {
		b := make([]int, len(s.a))
		x := make([]int, len(s.x))
		for j := range x {
			x[j] = -1
		}
		o := opt.homo_solve(s.a, b, x)
		if o != s.o {
			for _, a := range s.a {
				fmt.Printf("%2v\n", a)
			}
			t.Errorf("o=%v : expect=%v\n x    =%2v\nexpect=%2v\n", o, s.o, x, s.x)
			continue
		}

		for j := range x {
			if s.x[j] >= 0 && s.x[j] != x[j] {
				for _, a := range s.a {
					fmt.Printf("%2v\n", a)
				}
				t.Errorf("o=%v : expect=%v\n x    =%2v\nexpect=%2v\n", o, s.o, x, s.x)
				continue
			}
		}
	}
}

func TestPolyHomoReconstruct(t *testing.T) {
	// 5*z^2+4*x*y*z+3*x^2*y+2*x
	p1 := NewPolyCoef(2,
		NewPolyCoef(1, NewPolyCoef(0, 0, 2), NewPolyCoef(0, 0, 0, 3)),
		NewPolyCoef(1, 0, NewPolyCoef(0, 0, 4)), 5)

	for i, ss := range []struct {
		lvs    Levels
		lv     Level
		tdeg   int
		input  *Poly
		output *Poly
	}{
		{
			[]Level{0, 1, 2}, 3,
			3,
			p1,
			// 5*w*z^2+4*x*y*z+3*x^2*y+2*x*w^2
			NewPolyCoef(3,
				NewPolyCoef(2, NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, 3)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 4))),
				NewPolyCoef(2, 0, 0, 5),
				NewPolyCoef(0, 0, 2)),
		}, {
			[]Level{0, 1, 3}, 3,
			3,
			p1,
			// 5*z^2*w^3+4*x*y*z*w+3*x^2*y+2*x*w^2
			NewPolyCoef(3,
				NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, 3)),
				NewPolyCoef(2, 0, NewPolyCoef(1, 0, NewPolyCoef(0, 0, 4))),
				NewPolyCoef(0, 0, 2),
				NewPolyCoef(2, 0, 0, 5)),
		}, {
			[]Level{0, 2, 3}, 3,
			2,
			p1,
			// 2*x*w+5*z^2+4*x*y*z+3*x^2*y
			NewPolyCoef(3,
				NewPolyCoef(2, NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, 3)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 4)), 5),
				NewPolyCoef(0, 0, 2)),
		}, {
			[]Level{1, 2, 3}, 3,
			2,
			p1,
			// 2*x*w^2+3*x^2*y*w+5*z^2+4*x*y*z
			NewPolyCoef(3,
				NewPolyCoef(2, 0, NewPolyCoef(1, 0, NewPolyCoef(0, 0, 4)), 5),
				NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, 3)),
				NewPolyCoef(0, 0, 2)),
		},
	} {
		for j, tt := range []RObj{
			one, mone, two,
			NewPolyVar(5),
			Add(NewPolyVar(5), mone),
			Mul(Add(NewPolyVar(5), mone), Add(NewPolyVar(6), mone)),
		} {
			input := ss.input.Mul(tt).(*Poly)
			tdeg := input.tdeg(ss.lvs)
			if tdeg != ss.tdeg {
				t.Errorf("<%d,%d>invalid tdeg ret=%d\nLvs=%d/%v, tdeg=%d\ninput=%v\noutput=%v",
					i, j, tdeg, ss.lv, ss.lvs, ss.tdeg, input, ss.output)
				continue
			}

			o := input.homo_reconstruct(ss.lv, ss.lvs, tdeg)
			if err := o.valid(); err != nil {
				t.Errorf("<%d,%d>invalid homo valid %v\nLvs=%d/%v, tdeg=%d\ninput=%v\noutput=%v\nactual=%v",
					i, j, err, ss.lv, ss.lvs, ss.tdeg, input, ss.output, o)
				continue

			}

			e := ss.output.Mul(tt)
			if !o.Equals(e) {
				t.Errorf("<%d,%d>invalid homo\nLvs=%d/%v, tdeg=%d\ninput=%v\noutput=%v\nactual=%v",
					i, j, ss.lv, ss.lvs, ss.tdeg, input, e, o)
				continue

			}
		}
	}
}
