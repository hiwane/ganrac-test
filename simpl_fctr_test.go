package ganrac

import (
	"fmt"
	"testing"
)

func testSameFormAndOrFctr(output, expect Fof) bool {
	// 形として同じか.
	// QE によるチェックでは等価性は確認できるが簡単化はわからない
	if output.fofTag() != expect.fofTag() {
		return false
	}
	if !output.IsQff() {
		return false
	}
	if !expect.IsQff() {
		return false
	}

	var oofmls, eefmls []Fof
	switch oo := output.(type) {
	case *FmlAnd:
		oofmls = oo.fml
		ee := expect.(*FmlAnd)
		eefmls = ee.fml
	case *FmlOr:
		oofmls = oo.fml
		ee := expect.(*FmlOr)
		eefmls = ee.fml
	default:
		return oo.Equals(expect)
	}

	if len(oofmls) != len(eefmls) {
		return false
	}

	for i := 0; i < len(oofmls); i++ {
		vars := make([]bool, 5)
		oofmls[i].Indets(vars)
		lv := Level(0)
		for j := 0; j < len(vars); j++ {
			if vars[j] {
				lv = Level(j)
			}
		}

		m := 0
		for j := 0; j < len(eefmls); j++ {
			vars = make([]bool, 5)
			eefmls[j].Indets(vars)
			if vars[lv] {
				m += 1
				if !oofmls[i].Equals(eefmls[j]) {
					return false
				}
			}
		}
		if m != 1 {
			fmt.Printf("gege %d\n", m)
			return false
		}
	}
	return true
}

func TestSimplFctr(t *testing.T) {

	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		return
	}
	defer connc.Close()
	defer connd.Close()

	x := NewPolyVar(0)
	y := NewPolyVar(1)
	z := NewPolyVar(2)

	for i, s := range []struct {
		input  Fof
		expect Fof
	}{
		{
			NewAtom(x.powi(3), EQ),
			NewAtom(x, EQ),
		}, {
			NewAtom(x.powi(2).Mul(NewInt(5)), LE),
			NewAtom(x, EQ),
		}, {
			NewAtom(x.powi(2).Mul(NewInt(5)), LT),
			falseObj,
		}, {
			NewAtom(x.powi(2).Mul(NewInt(5)), GE),
			trueObj,
		}, {
			NewAtom(x.powi(2).Mul(NewInt(5)), GT),
			NewAtom(x, NE),
		}, {
			NewAtom(x.powi(2).Mul(NewInt(5)), NE),
			NewAtom(x, NE),
		}, {
			NewAtom(x.powi(2).Mul(NewInt(5)), EQ),
			NewAtom(x, EQ),
		}, {
			NewAtom(x.powi(3).Mul(NewInt(5)), LE),
			NewAtom(x, LE),
		}, {
			NewAtom(x.powi(3).Mul(NewInt(5)), LT),
			NewAtom(x, LT),
		}, {
			NewAtom(x.powi(3).Mul(NewInt(5)), GE),
			NewAtom(x, GE),
		}, {
			NewAtom(x.powi(3).Mul(NewInt(5)), GT),
			NewAtom(x, GT),
		}, {
			NewAtom(x.powi(3).Mul(NewInt(5)), NE),
			NewAtom(x, NE),
		}, {
			NewAtom(x.powi(3).Mul(NewInt(5)), EQ),
			NewAtom(x, EQ),
		}, { // 13
			NewAtom(x.powi(3).Mul(y.powi(4)), LE),
			NewFmlOr(NewAtom(y, EQ), NewAtom(x, LE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)), GE),
			NewFmlOr(NewAtom(y, EQ), NewAtom(x, GE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)), LT),
			NewFmlAnd(NewAtom(y, NE), NewAtom(x, LT)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)), GT),
			NewFmlAnd(NewAtom(y, NE), NewAtom(x, GT)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)), EQ),
			NewFmlOr(NewAtom(y, EQ), NewAtom(x, EQ)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)), NE),
			NewFmlAnd(NewAtom(y, NE), NewAtom(x, NE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z).Mul(NewInt(5)), LE),
			NewFmlOr(NewAtom(y, EQ), NewAtoms([]RObj{x, z}, LE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z).Mul(NewInt(5)), GE),
			NewFmlOr(NewAtom(y, EQ), NewAtoms([]RObj{x, z}, GE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z).Mul(NewInt(5)), LT),
			NewFmlAnd(NewAtom(y, NE), NewAtoms([]RObj{x, z}, LT)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z).Mul(NewInt(5)), GT),
			NewFmlAnd(NewAtom(y, NE), NewAtoms([]RObj{x, z}, GT)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z).Mul(NewInt(5)), EQ),
			newFmlOrs(NewAtom(y, EQ), NewAtom(z, EQ), NewAtom(x, EQ)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z).Mul(NewInt(5)), NE),
			newFmlAnds(NewAtom(y, NE), NewAtom(z, NE), NewAtom(x, NE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z.powi(4)).Mul(NewInt(5)), LE),
			newFmlOrs(NewAtom(y, EQ), NewAtom(z, EQ), NewAtom(x, LE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z.powi(4)).Mul(NewInt(5)), GE),
			newFmlOrs(NewAtom(y, EQ), NewAtom(z, EQ), NewAtom(x, GE)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z.powi(4)).Mul(NewInt(5)), LT),
			newFmlAnds(NewAtom(y, NE), NewAtom(z, NE), NewAtom(x, LT)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z.powi(4)).Mul(NewInt(5)), GT),
			newFmlAnds(NewAtom(y, NE), NewAtom(z, NE), NewAtom(x, GT)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z.powi(4)).Mul(NewInt(5)), EQ),
			newFmlOrs(NewAtom(y, EQ), NewAtom(z, EQ), NewAtom(x, EQ)),
		}, {
			NewAtom(x.powi(3).Mul(y.powi(4)).Mul(z.powi(4)).Mul(NewInt(5)), NE),
			newFmlAnds(NewAtom(y, NE), NewAtom(z, NE), NewAtom(x, NE)),
		},
	} {
		var output Fof
		output = s.input.simplFctr(g)

		if !testSameFormAndOrFctr(output, s.expect) {
			fmt.Printf("expect %V\n", s.expect)
			fmt.Printf("output %V\n", output)
			t.Errorf("i=%d\ninput =%v\nexpect=%v\nactual=%v", i, s.input, s.expect, output)
			continue
		}
	}
}
