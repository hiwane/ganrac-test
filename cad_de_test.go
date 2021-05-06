package ganrac

import (
	"fmt"
	"testing"
)

func TestSymSqfr(t *testing.T) {
	// ox は必要ないのだけど．
	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		return
	}
	defer connc.Close()
	defer connd.Close()

	fof := NewQuantifier(false, []Level{3}, NewAtom(NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 1), NewInt(1)), NewInt(1)), NewInt(1)), GT))
	cad, err := NewCAD(fof, g)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	cad.initProj(0)
	g.ox = nil
	cell := NewCell(cad, cad.root, 1)
	cell.defpoly = NewPolyCoef(0, -2, 0, 1)
	for ii, s := range []struct {
		p *Poly
		r []int8
		d []int
	}{
		{
			NewPolyCoef(3, -3, 7, -5, 1),
			[]int8{1, 2},
			[]int{1, 1},
		}, {
			NewPolyCoef(3, -5, 15, -16, 8, -3, 1),
			[]int8{1, 3},
			[]int{2, 1},
		}, {
			NewPolyCoef(3, NewInt(2), NewPolyCoef(0, 0, -2), NewInt(1)),
			[]int8{2},
			[]int{1},
		}, {
			NewPolyCoef(3, NewPolyCoef(0, 0, 0, 0, 0, -1), NewPolyCoef(0, 0, 10), NewInt(-12), NewPolyCoef(0, 0, -4), NewInt(8)),
			[]int8{1, 3},
			[]int{1, 1},
		},
	} {
		for _, ppp := range []*Poly{
			s.p, s.p.Neg().(*Poly),
			s.p.subsXinv(), s.p.subsXinv().Neg().(*Poly)} {

			sqfr := cad.sym_sqfr(ppp, cell)
			var q RObj = one
			for i, sq := range sqfr {
				fmt.Printf("<%d> [%v]^%d\n", i, sq.p, sq.r)
				if sq.r != s.r[i] {
					t.Errorf("<%d> r[%d] expect=%d actual=%d\nret=%v", ii, i, s.r[i], sq.r, sq.p)
					break
				}
				if len(sq.p.c)-1 != s.d[i] {
					t.Errorf("<%d> d[%d] expect=%d\nactual=%d\nret=%v", ii, i, s.d[i], sq.p, sq)
					break
				}

				qq := sq.p.Pow(NewInt(int64(sq.r)))
				q = Mul(qq, q)
			}
			qq := Sub(Mul(q, ppp.lc()), Mul(ppp, q.(*Poly).lc()))
			if !qq.IsZero() {
				if qx, ok := qq.(*Poly); ok {
					flag := true
					if qx.lv == ppp.lv {
						for _, qc := range qx.c {
							switch qcc := qc.(type) {
							case *Poly:
								if !cad.sym_zero_chk(qcc, cell) {
									flag = false
								}
							case NObj:
								if !qcc.IsZero() {
									flag = false
								}
							}
						}
						if flag {
							continue
						}
					} else if cad.sym_zero_chk(qx, cell) {
						continue
					}
				}
				t.Errorf("<%d>\ninput =%v\noutput=%v\nret=%v\nqq=%v", ii, ppp, q, sqfr, qq)
			}
		}
	}
}
