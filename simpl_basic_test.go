package ganrac

import (
	"fmt"
	"testing"
)

func testSameFormAndOr(output, expect Fof) bool {
	// 形として同じか.
	// QE によるチェックでは等価性は確認できるが簡単化はわからない
	output = output.normalize()
	expect = expect.normalize()

	if output.fofTag() != expect.fofTag() {
		return false
	}
	if !output.IsQff() {
		return false
	}
	if !expect.IsQff() {
		return false
	}

	switch oo := output.(type) {
	case lener:
		switch ee := expect.(type) {
		case lener:
			if oo.Len() != ee.Len() {
				return false
			}
		}
	}

	switch oo := output.(type) {
	case *FmlAnd:
		ee := expect.(*FmlAnd)
		for i := 0; i < len(oo.fml); i++ {
			if !testSameFormAndOr(oo.fml[i], ee.fml[i]) {
				return false
			}
		}
	case *FmlOr:
		ee := expect.(*FmlOr)
		for i := 0; i < len(oo.fml); i++ {
			if !testSameFormAndOr(oo.fml[i], ee.fml[i]) {
				return false
			}
		}
	}

	return true
}

func TestSimplBasicAndOr2(t *testing.T) {
	c := int64(0)
	d := int64(1)

	for i, s := range []struct {
		a, b   Fof
		expect Fof // simplified a /\ b
	}{
		{
			NewAtom(NewPolyCoef(0, 0, 1), LE),
			NewAtom(NewPolyCoef(1, 0, 1), LE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), LE), // x <= 0
			NewAtom(NewPolyCoef(0, 0, 1), LE),
			NewAtom(NewPolyCoef(0, 0, 1), LE),
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), LE), // x <= 0
			NewAtom(NewPolyCoef(0, 1, 1), LE), // x <= -1
			NewAtom(NewPolyCoef(0, 1, 1), LE),
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), LE),
			NewAtom(NewPolyCoef(0, 1, 1), LT),
			NewAtom(NewPolyCoef(0, 1, 1), LT),
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), LE), // x <= 0
			NewAtom(NewPolyCoef(0, 1, 1), GT), // x >= -1
			nil,
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), LE),  // x <= 0
			NewAtom(NewPolyCoef(0, -1, 1), GT), // x >= 1
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), LE),  // x <= 0
			NewAtom(NewPolyCoef(0, -1, 1), NE), // x != 1
			NewAtom(NewPolyCoef(0, 0, 1), LE),  // x <= 0
		}, {
			NewAtom(NewPolyCoef(0, 0, 1), EQ),  // x = 0
			NewAtom(NewPolyCoef(0, -1, 1), NE), // x != 1
			NewAtom(NewPolyCoef(0, 0, 1), EQ),  // x = 0
		}, {
			// Table 2 additive smart simplification assuming c < d
			NewAtom(NewPolyCoef(0, c, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), LE),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), GE),
			NewAtom(NewPolyCoef(0, c, 1), EQ),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), NE),
			NewAtom(NewPolyCoef(0, c, 1), EQ),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), LT),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), GT),
			NewAtom(NewPolyCoef(0, c, 1), EQ),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), LE),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), GE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), NE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), LT),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), GT),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GE),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GE),
			NewAtom(NewPolyCoef(0, d, 1), LE),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GE),
			NewAtom(NewPolyCoef(0, d, 1), GE),
			NewAtom(NewPolyCoef(0, c, 1), GE),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GE),
			NewAtom(NewPolyCoef(0, d, 1), NE),
			NewAtom(NewPolyCoef(0, c, 1), GE),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GE),
			NewAtom(NewPolyCoef(0, d, 1), LT),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GE),
			NewAtom(NewPolyCoef(0, d, 1), GT),
			NewAtom(NewPolyCoef(0, c, 1), GE),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), NE),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), NE),
			NewAtom(NewPolyCoef(0, d, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), LE),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), NE),
			NewAtom(NewPolyCoef(0, d, 1), GE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), NE),
			NewAtom(NewPolyCoef(0, d, 1), NE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), NE),
			NewAtom(NewPolyCoef(0, d, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), LT),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), NE),
			NewAtom(NewPolyCoef(0, d, 1), GT),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), LE),
			NewAtom(NewPolyCoef(0, d, 1), LE),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), GE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), NE),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), LT),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), LT),
			NewAtom(NewPolyCoef(0, d, 1), GT),
			nil,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GT),
			NewAtom(NewPolyCoef(0, d, 1), EQ),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GT),
			NewAtom(NewPolyCoef(0, d, 1), LE),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GT),
			NewAtom(NewPolyCoef(0, d, 1), GE),
			NewAtom(NewPolyCoef(0, c, 1), GT),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GT),
			NewAtom(NewPolyCoef(0, d, 1), NE),
			NewAtom(NewPolyCoef(0, c, 1), GT),
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GT),
			NewAtom(NewPolyCoef(0, d, 1), LT),
			falseObj,
		}, {
			NewAtom(NewPolyCoef(0, c, 1), GT),
			NewAtom(NewPolyCoef(0, d, 1), GT),
			NewAtom(NewPolyCoef(0, c, 1), GT),
		},
	} {

		for ii, input := range []Fof{
			NewFmlAnd(s.a, s.b),
			NewFmlAnd(s.b, s.a),
			newFmlAnds(s.b, s.a, s.b),
			newFmlAnds(s.a, s.b, s.a, s.b, s.a, s.a, s.b),
		} {
			var q, output, expect Fof

			if s.expect == nil {
				expect = NewFmlAnd(s.a, s.b)
			} else {
				expect = s.expect
			}

			output = input.simplBasic(trueObj, falseObj)
			// 項の数が一致する.
			if !testSameFormAndOr(output, expect) {
				t.Errorf("%d:%d: not same form:\ninput =`%v`\noutput=`%v`\nexpect=`%v`", i, ii, input, output, expect)
				return
			}

			lllv := []Level{0, 1, 2, 3, 4, 5, 6}

			// 等価である.
			q = NewQuantifier(true, lllv, newFmlEquiv(output, expect))
			for _, llv := range lllv {
				qnew := vsLinear(q, llv)
				if qnew.hasVar(llv) {
					t.Errorf("%d: variable %d is not eliminated: X1\nin =`%v`\nout=`%s`", i, llv, q, qnew)
					return
				}
				q = qnew
			}
			if _, ok := q.(*AtomT); !ok {
				fmt.Printf("q=%v\n", q)
				t.Errorf("%d: qe failed\ninput =%v\nexpect=%v\nactual=%v", i, input, expect, output)
				return
			}

			input = input.Not()
			expect = expect.Not()

			output = input.simplBasic(trueObj, falseObj)

			if !testSameFormAndOr(output, expect) {
				t.Errorf("%d: not same form:\ninput =`%v`\noutput=`%v`\nexpect=`%v`", i, input, output, expect)
				return
			}
			q = NewQuantifier(true, lllv, newFmlEquiv(output, expect))
			for _, llv := range lllv {
				q = vsLinear(q, llv)
				if q.hasVar(llv) {
					t.Errorf("%d: variable %d is not eliminated: X1 out=%s", i, llv, q)
					return
				}
			}
			if _, ok := q.(*AtomT); !ok {
				fmt.Printf("q=%v\n", q)
				t.Errorf("%d: qe failed input=%v: expect=%v, actual=%v", i, input, expect, output)
				return
			}
		}
	}
}

func newAtomVar(lv Level, op OP) Fof {
	return NewAtom(NewPolyVar(lv), op)
}

func TestSimplSmartAndOrn(t *testing.T) {
	// 5.3 illustrating examples
	for i, s := range []struct {
		input, expect Fof
	}{
		{
			NewFmlAnd(newAtomVar(0, EQ),
				NewFmlOr(newAtomVar(1, NE),
					NewFmlAnd(newAtomVar(2, LE),
						NewFmlOr(newAtomVar(3, GT), newAtomVar(0, EQ))))),
			NewFmlAnd(newAtomVar(0, EQ),
				NewFmlOr(newAtomVar(1, NE), newAtomVar(2, LE))),
		}, {
			newFmlAnds(
				newAtomVar(0, EQ),
				NewFmlOr(newAtomVar(1, EQ), NewFmlAnd(newAtomVar(2, EQ), newAtomVar(3, GE))),
				NewFmlOr(newAtomVar(3, NE), newAtomVar(0, NE))),
			newFmlAnds(
				newAtomVar(0, EQ),
				NewFmlOr(newAtomVar(1, EQ),
					NewFmlAnd(newAtomVar(2, EQ), newAtomVar(3, GT))),
				newAtomVar(3, NE)),
		}, {
			newFmlAnds(
				newAtomVar(0, GT),
				NewAtom(NewPolyCoef(0, -1, 2), GT),
				NewAtom(NewPolyCoef(0, 5, 3), NE)),
			NewAtom(NewPolyCoef(0, -1, 2), GT),
		}, {
			NewFmlOr(
				NewAtom(NewPolyCoef(1, NewPolyCoef(0, 4, 1), 0, 1), GE),
				NewAtom(NewPolyCoef(1, NewPolyCoef(0, 4, 7), 0, 7), LE)),
			trueObj,
		},
	} {
		output := s.input.simplBasic(trueObj, falseObj)
		if !testSameFormAndOr(output, s.expect) {
			t.Errorf("%d: not same form:\ninput =`%v`\noutput=`%v`\nexpect=`%v`", i, s.input, output, s.expect)
			continue
		}

		var q Fof
		lllv := []Level{0, 1, 2, 3, 4, 5, 6}

		// 等価である.
		q = NewQuantifier(true, lllv, newFmlEquiv(output, s.expect))
		for _, llv := range lllv {
			qnew := vsLinear(q, llv)
			if qnew.hasVar(llv) {
				t.Errorf("%d: variable %d is not eliminated: X1\nin =`%v`\nout=`%s`", i, llv, q, qnew)
				continue
			}
			q = qnew
		}
		if _, ok := q.(*AtomT); !ok {
			fmt.Printf("q=%v\n", q)
			t.Errorf("%d: qe failed\ninput =%v\nexpect=%v\nactual=%v", i, s.input, s.expect, output)
			continue
		}
	}
}
