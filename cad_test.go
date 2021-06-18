package ganrac

import (
	"fmt"
	"testing"
)

func TestCADeasy(t *testing.T) {
	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		fmt.Printf("skip TestCADeasy... (no ox)\n")
		return
	}
	defer connc.Close()
	defer connd.Close()

	x := NewPolyVar(0)
	y := NewPolyVar(1)

	for _, s := range []struct {
		input      Fof
		root_truth int8
	}{
		{
			NewQuantifier(true, []Level{0},
				NewQuantifier(false, []Level{1},
					NewAtom(x.Sub(y), EQ))),
			t_true,
		}, {
			NewQuantifier(false, []Level{0},
				NewQuantifier(true, []Level{1},
					NewAtom(x.Sub(y), EQ))),
			t_false,
		}, {
			NewQuantifier(false, []Level{0, 1},
				NewFmlAnd(NewAtom(x.Mul(x).Add(y.Mul(y)).Add(NewInt(-9)), LE),
					NewAtom(x.Mul(x).Add(NewInt(-5)), GT))),
			t_true,
		}, {
			NewQuantifier(false, []Level{0, 1},
				NewFmlAnd(NewAtom(NewPolyCoef(0, -2, 0, 1), EQ),
					NewAtom(NewPolyCoef(1, NewInt(-1), NewPolyCoef(0, -2, 0, 1)), EQ))),
			t_false,
			//		}, { exAdam1().Input, 1,
		},
	} {
		cad, err := NewCAD(s.input, g)
		if err != nil {
			t.Errorf("\ninput =%v\nerr=%v\n", s.input, err)
			continue
		}
		cad.Projection(0)
		cad.Lift()

		if cad.root.truth != s.root_truth {
			t.Errorf("\ninput =%v\nexpect=%v\noutput=%v\n", s.input, s.root_truth, cad.root.truth)
		}
	}
}
