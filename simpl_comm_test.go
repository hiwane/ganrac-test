package ganrac

import (
	"testing"
)

func TestSimplComm1(t *testing.T) {
	x := NewPolyCoef(0, 1, 2)
	y := NewPolyCoef(1, 2, 3)
	z := NewPolyCoef(2, 4, -3)
	w := NewPolyCoef(3, 4, -1)
	X := NewAtom(x, LT)
	Y := NewAtom(y, LE)
	Z := NewAtom(z, EQ)
	W := NewAtom(w, GE)

	for ii, ss := range []struct {
		input  Fof
		expect Fof // simplified input
	}{
		{ // 0
			NewFmlAnd(NewFmlOr(X, Y), NewFmlOr(X, Y)),
			NewFmlOr(X, Y),
		}, { // 1
			NewFmlAnd(NewFmlOr(X, Y), newFmlOrs(X, Y, Z)),
			newFmlOrs(X, Y, Z),
		}, { // 2
			NewFmlAnd(newFmlOrs(X, Y, W, Z), newFmlOrs(X, Y, Z)),
			newFmlOrs(X, Y, W, Z),
		}, { // 3
			NewFmlAnd(newFmlOrs(W, X, Y, Z), newFmlOrs(X, Y, Z)),
			newFmlOrs(X, Y, W, Z),
		}, { // 4
			NewFmlAnd(newFmlOrs(X, Y, NewAtom(z, GT)), newFmlOrs(X, Y, NewAtom(z, GE))),
			newFmlOrs(X, Y, NewAtom(z, GT)),
		}, {
			NewFmlAnd(X, newFmlOrs(Z, NewFmlAnd(X, Y))),
			NewFmlAnd(X, newFmlOrs(Z, Y)),
		}, {
			NewFmlAnd(X, NewFmlOr(X, Y)),
			X,
		}, {
			NewFmlAnd(X, newFmlOrs(X, Y, Z)),
			X,
		},
	} {
		if ss.expect == nil {
			ss.expect = ss.input
		}
		for i, s := range []struct {
			input  Fof
			expect Fof // simplified input
		}{
			{ss.input, ss.expect},
			{ss.input.Not(), ss.expect.Not()},
		} {
			output := s.input.simplComm()
			if testSameFormAndOr(output, s.expect) {
				continue
			}

			out2 := output.simplBasic(trueObj, falseObj)
			if testSameFormAndOr(out2, s.expect) {
				continue
			}

			t.Errorf("%d/%d: not same form:\ninput =`%v`\noutput=`%v`\noutput=`%v`\nexpect=`%v`", ii, i, s.input, output, out2, s.expect)
			return
		}
	}
}
