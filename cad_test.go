package ganrac

import (
	"testing"
)

func TestCADeasy(t *testing.T) {
	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
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
			NewQuantifier(true, []Level{1},
				NewQuantifier(false, []Level{0},
					NewAtom(x.Sub(y), EQ))),
			0,
		}, {
			NewQuantifier(false, []Level{1},
				NewQuantifier(true, []Level{0},
					NewAtom(x.Sub(y), EQ))),
			1,
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
