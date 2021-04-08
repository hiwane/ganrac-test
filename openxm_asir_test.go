package ganrac

import (
	"testing"
)

func TestAsirDiscrim(t *testing.T) {

	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
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
			NewPolyCoef(2, NewPolyInts(1, 0, 0, 1), NewPolyInts(0, 0, -4)), // b^2-4ac
		},
	} {
		output := g.ox.Discrim(s.input, s.lv)
		if !output.Equals(s.expect) {
			t.Errorf("lv=%d\ninput =%v\nexpect=%v\noutput=%v\n", s.lv, s.input, s.expect, output)
		}
	}
}
