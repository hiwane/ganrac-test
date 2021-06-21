package ganrac

import (
	"fmt"
	"testing"
)

func TestAsirDiscrim(t *testing.T) {

	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		fmt.Printf("skip TestAsirDiscrim... (no ox)\n")
		return
	}
	defer connc.Close()
	defer connd.Close()

	for _, s := range []struct {
		lv     Level
		input  *Poly
		expect RObj
	}{
		{3,
			NewPolyCoef(3, NewPolyVar(0), NewPolyVar(1), NewPolyVar(2)),    // ax^2+bx+c
			NewPolyCoef(2, NewPolyCoef(1, 0, 0, 1), NewPolyCoef(0, 0, -4)), // b^2-4ac
		},
	} {
		output := g.ox.Discrim(s.input, s.lv)
		if !output.Equals(s.expect) {
			t.Errorf("lv=%d\ninput =%v\nexpect=%v\noutput=%v\n", s.lv, s.input, s.expect, output)
		}
	}
}

func TestAsirSres(t *testing.T) {
	// Quantifier Elimination for Formulas Constrained by Quadratic Equations via Slope Resultants
	// Hoon Hong, The computer J., 1993

	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		fmt.Printf("skip TestAsirSres... (no ox)\n")
		return
	}
	defer connc.Close()
	defer connd.Close()

	// > vars(x,u,v,w);
	// > A = u*x^2+v*x+1;
	// > B = v*x^3+w*x+u;
	// > C = w*x^2+v*x+u;
	A := NewPolyCoef(2, NewPolyCoef(1, 1, NewPolyCoef(0, 0, 0, 1)), NewPolyCoef(0, 0, 1))
	B := NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 0, 0, 1)), NewPolyCoef(0, 0, 1))
	C := NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 1)), NewPolyCoef(0, 0, 0, 1))

	TB := NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 0, 0, 0, 2), 0, NewPolyCoef(1, 0, 3), 0, -1), NewPolyCoef(2, 0, NewPolyCoef(1, 0, 0, -1)))
	SB := NewPolyCoef(3, NewPolyCoef(2, 0, NewPolyCoef(1, 0, -1), 0, 1), NewPolyCoef(1, 0, 0, 1))
	TC := NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 0, 0, 2), 0, NewPolyCoef(1, 0, -1)), NewPolyCoef(2, NewPolyCoef(1, 0, -2), 0, 1))
	SC := NewPolyCoef(3, NewPolyCoef(2, 0, NewPolyCoef(1, 0, 1)), NewPolyCoef(2, 0, -1))

	for ii, ss := range []struct {
		p      *Poly
		q      *Poly
		k      int32
		expect *Poly
	}{
		{A, B, 0, TB},
		{A, B, 1, SB},
		{A, C, 0, TC},
		{A, C, 1, SC},
	} {
		o := g.ox.Sres(ss.p, ss.q, 0, ss.k)
		if !o.Equals(ss.expect) {
			t.Errorf("invalid <%d,%d>\ninput=%v\nexpect=%v\noutput=%v\n",
				ii, ss.k, ss.q, ss.expect, o)
		}
	}

}
