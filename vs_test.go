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
		}{		// 再帰表現なので，自由変数と束縛変数のレベルの大小で動きが異なる
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
				t.Errorf("%d: eval failed input=%s: eval=%v, err:%s", i, s.qff, fof, err)
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
