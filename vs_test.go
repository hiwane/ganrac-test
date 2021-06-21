package ganrac

import (
	"fmt"
	"strings"
	"testing"
)

func TestVsLin(t *testing.T) {

	g := NewGANRAC()

	for i, ss := range []struct {
		qff    string
		expect string
	}{
		{"x*y >0", "x != 0;"},
		{"y>=0", "true;"},
		{"x+y >=0", "true;"},
		{"x+y >0", "true;"},
		{"x*y <0", "x != 0;"},
		{"x*y >=0", "true;"},
		{"(x+1)*y <0", "x != -1;"},
		{"(2-x)*y >0", "x != 2;"},
		{"x+y <0", "true;"},
		{"x+y >=0", "true;"},
		{"x+y >=0 && x + y + 2 <= 0", "false;"},
		{"(y+1)*x-3 >=0", "x!=0;"},
		{"(x+1)*y-3 >=0", "x!=-1;"},
		{"y > a && y <= b", "a < b;"},
		{"y == a && y != b && a > 0", "a != b && a > 0;"},
		{"y == x && y != b && x > 0", "x != b && x > 0;"},
	} {
		for j, s := range []struct {
			qff    string
			expect string
		}{ // 再帰表現なので，自由変数と束縛変数のレベルの大小で動きが異なる
			{ss.qff, ss.expect},
			{strings.ReplaceAll(ss.qff, "x", "z"),
				strings.ReplaceAll(ss.expect, "x", "z")}} {

			//fmt.Printf("==============[%d,%d]==================================\n%v\n", i, j, s.qff)
			_fof, err := g.Eval(strings.NewReader(fmt.Sprintf("ex([y], %s);", s.qff)))
			if err != nil {
				t.Errorf("%d-%d: eval failed input=`%s`: err:`%s`", i, j, s.qff, err)
				return
			}
			fof, ok := _fof.(Fof)
			if !ok {
				t.Errorf("%d: eval failed\ninput=%s\neval=%v\nerr:%s", i, s.qff, fof, err)
				return
			}

			ans, err := g.Eval(strings.NewReader(s.expect))
			if err != nil {
				t.Errorf("%d: eval failed input=%s: err:%s", i, s.expect, err)
				return
			}

			lv := Level(1)
			qff := vsLinear(fof, lv)
			if err = qff.valid(); err != nil {
				t.Errorf("%d: formula is broken input=`%s`: out=`%s`, %v", i, s.qff, qff, err)
				return
			}

			if qff.hasVar(lv) {
				t.Errorf("%d: variable %d is not eliminated input=%s: out=%s", i, lv, s.qff, qff)
				return
			}

			if !qff.Equals(ans) {
				// qff.dump(os.Stdout)
				// fmt.Printf("\n")
				//
				// fmt.Printf("----------------------------\n")
				var q Fof
				lllv := []Level{0, 2, 3, 4, 5, 6}
				q = NewQuantifier(true, lllv, newFmlEquiv(qff, ans.(Fof)))
				for _, llv := range lllv {
					q = vsLinear(q, llv)
					if q.hasVar(llv) {
						t.Errorf("%d: variable %d is not eliminated: X1 out=%s", i, llv, q)
						return
					}
				}
				if _, ok := q.(*AtomT); !ok {
					fmt.Printf("q=%v\n", q)
					t.Errorf("%d: qe failed\ninput =%s\nexpect=%v\nactual=%v", i, s.qff, s.expect, qff)
					return
				}
			}

			sqff := qff.simplBasic(trueObj, falseObj)
			if err = sqff.valid(); err != nil {
				t.Errorf("%d: formula is broken input=`%v`: out=`%v`, %v", i, qff, sqff, err)
				return
			}

			if !sqff.Equals(ans) {
				var q Fof
				lllv := []Level{0, 2, 3, 4, 5, 6}
				q = NewQuantifier(true, lllv, newFmlEquiv(sqff, ans.(Fof)))
				for _, llv := range lllv {
					q = vsLinear(q, llv)
					if q.hasVar(llv) {
						t.Errorf("%d: variable %d is not eliminated: X1 out=%s", i, llv, q)
						return
					}
				}
				if _, ok := q.(*AtomT); !ok {
					fmt.Printf("q=%v\n", q)
					t.Errorf("%d: qe failed\ninput =%s\nexpect=%v\nactual=%v", i, qff, s.expect, s.qff)
					return
				}
			}
		}
	}
}

func TestVsLin2(t *testing.T) {

	for ii, ss := range []struct {
		lv     Level
		p      Fof
		expect Fof
	}{
		{2, // ex([x], a*x == b && 3*x > 1);
			NewQuantifier(false, []Level{2}, newFmlAnds(
				NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, -1), NewPolyCoef(0, 0, 1)), EQ),
				NewAtom(NewPolyCoef(2, -1, 3), GT))),
			newFmlOrs( // [ a = 0 /\ 3 b - a = 0 ] \/ [ a > 0 /\ 3 b - a > 0 ] \/ [ a < 0 /\ 3 b - a < 0 ]
				newFmlAnds(
					NewAtom(NewPolyCoef(0, 0, 1), EQ),
					NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 3), EQ)),
				newFmlAnds(
					NewAtom(NewPolyCoef(0, 0, 1), GT),
					NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 3), GT)),
				newFmlAnds(
					NewAtom(NewPolyCoef(0, 0, 1), LT),
					NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 3), LT))),
		}, {2, // ex([x], a*x < b && 3*x > 1);
			NewQuantifier(false, []Level{2}, newFmlAnds(
				NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, -1), NewPolyCoef(0, 0, 1)), LT),
				NewAtom(NewPolyCoef(2, -1, 3), GT))),
			newFmlOrs(
				NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 3), GT),
				NewAtom(NewPolyCoef(0, 0, 1), LT)),
		}, {2, // ex([x], a*x > b && 3*x > 1);
			NewQuantifier(false, []Level{2}, newFmlAnds(
				NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, -1), NewPolyCoef(0, 0, 1)), GT),
				NewAtom(NewPolyCoef(2, -1, 3), GT))),
			newFmlOrs(
				NewAtom(NewPolyCoef(0, 0, 1), GT),
				NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 3), LT)),
		}, {2,
			// ex([x], a*x+b >= 0 && 3*x+1 > 0)
			NewQuantifier(false, []Level{2}, newFmlAnds(
				NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 1)), GE),
				NewAtom(NewPolyCoef(2, 1, 3), GT))),
			// b >= 0 \/ a > 0 \/ 3 b - a > 0
			newFmlOrs(
				NewAtom(NewPolyCoef(1, 0, 1), GE),
				NewAtom(NewPolyCoef(0, 0, 1), GT),
				NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, -1), 3), GT)),
		}, {4,
			// ex([x], a*x+b >= 0 && c*x+d > 0)
			NewQuantifier(false, []Level{4}, newFmlAnds(
				NewAtom(NewPolyCoef(4, NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 1)), GE),
				NewAtom(NewPolyCoef(4, NewPolyCoef(3, 0, 1), NewPolyCoef(2, 0, 1)), GT))),
			//  b >= 0 && d > 0  ||  a >= 0 && c > 0 && a*d - b*c <= 0  ||  a <= 0 && c < 0 && a*d - b*c >= 0  ||  a > 0 && a*d - b*c > 0  ||  a < 0 && a*d - b*c < 0
			newFmlOrs(
				newFmlAnds(NewAtom(NewPolyCoef(1, 0, 1), GE), NewAtom(NewPolyCoef(3, 0, 1), GT)),
				newFmlAnds(NewAtom(NewPolyCoef(0, 0, 1), GE), NewAtom(NewPolyCoef(2, 0, 1), GT), NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, NewPolyCoef(1, 0, -1)), NewPolyCoef(0, 0, 1)), LE)),
				newFmlAnds(NewAtom(NewPolyCoef(0, 0, 1), LE), NewAtom(NewPolyCoef(2, 0, 1), LT), NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, NewPolyCoef(1, 0, -1)), NewPolyCoef(0, 0, 1)), GE)),
				newFmlAnds(NewAtom(NewPolyCoef(0, 0, 1), GT), NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, NewPolyCoef(1, 0, -1)), NewPolyCoef(0, 0, 1)), GT)),
				newFmlAnds(NewAtom(NewPolyCoef(0, 0, 1), LT), NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, NewPolyCoef(1, 0, -1)), NewPolyCoef(0, 0, 1)), LT))),
		},
	} {
		for jj, sss := range []struct {
			p      Fof
			expect Fof
		}{
			{ss.p, ss.expect},
			{ss.p.Not(), ss.expect.Not()},
		} {
			f := vsLinear(sss.p, ss.lv)
			f = f.simplBasic(trueObj, falseObj)

			q := make([]Level, ss.lv)
			for i := Level(0); i < ss.lv; i++ {
				q[i] = i
			}
			g := NewQuantifier(true, q, FofEquiv(f, sss.expect))
			for i := Level(0); i < ss.lv; i++ {
				g = vsLinear(g, i)
				switch g.(type) {
				case *AtomF:
					t.Errorf("invalid %d, %d, F\n in=%v\nexp=%v\nout=%v", ii, jj, sss.p, sss.expect, f)
					return
				case FofQ:
					continue
				case *AtomT:
					break
				default:
					t.Errorf("invalid %d, %d, 2, %v\n in=%v\nexp=%v\nout=%v", ii, jj, g, sss.p, sss.expect, f)
					return
				}
				g = g.simplBasic(trueObj, falseObj)
			}
			if _, ok := g.(*AtomT); !ok {
				t.Errorf("invalid %d, %d, 3, %v\n in=%v\nexp=%v\nout=%v", ii, jj, g, sss.p, sss.expect, f)
				return
			}
		}
	}
}
