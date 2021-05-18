package ganrac

import (
	"fmt"
	"testing"
)

// func TestSymSqfr2(t *testing.T) {
// 	p := NewPolyCoef(2,
// 			NewPolyCoef(1, ParseInt("-4722201209543931678482839880788832023380506408746984383", 10), ParseInt("7037081672937279457168302335815099996936095844515446784", 10), ParseInt("-7370402196928139846235864997969418656470294922939858944", 10)),
// 			NewPolyCoef(1, ParseInt("-937089791168204322416009923114557457813156603499642880", 10), ParseInt("695112178837912870216157370077032489892039416212357120", 10), ParseInt("-59907021258380514474371077442594431218361575250329600", 10)),
// 			NewPolyCoef(1, ParseInt("-46489827399607932944452513320345263121254652785459200", 10), ParseInt("1412477194662381313033576520811520736137630973952000", 10), ParseInt("30100594890350423012826167438885134580526627998924800", 10)))
// 	d := NewPolyCoef(1,
// 			ParseInt("189708409272553183839474655", 10),
// 			ParseInt("-1436744321078070999056602079", 10),
// 			ParseInt("-5184691718935390049781415936", 10),
// 			ParseInt("5196819548962240103538753536", 10))
// }

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
	cell.defpoly = NewPolyCoef(0, -2, 0, 1) // x^2-2
	for ii, s := range []struct {
		p *Poly
		r []mult_t
		d []int
	}{
		{ // 0
			NewPolyCoef(3, -3, 7, -5, 1),
			[]mult_t{1, 2},
			[]int{1, 1},
		}, { // 1
			NewPolyCoef(3, -5, 15, -16, 8, -3, 1),
			[]mult_t{1, 3},
			[]int{2, 1},
		}, { // 2
			NewPolyCoef(3, 2, NewPolyCoef(0, 0, -2), 1),
			[]mult_t{2},
			[]int{1},
		}, { // 3
			NewPolyCoef(3, NewPolyCoef(0, 0, -2), 6, NewPolyCoef(0, 0, -3), 1),
			[]mult_t{3},
			[]int{1},
		}, { // 4
			// x^4-a*x^3-3*a^2*x^2+5*a^3*x-2*a^4
			NewPolyCoef(1, -8, NewPolyCoef(0, 0, 0, 0, 5), NewPolyCoef(0, 0, 0, -3), NewPolyCoef(0, 0, -1), 1),
			[]mult_t{1, 3},
			[]int{1, 1},
		}, { // 5
			NewPolyCoef(3, NewPolyCoef(0, 0, 0, 0, 0, -1), NewPolyCoef(0, 0, 10), -12, NewPolyCoef(0, 0, -4), 8),
			[]mult_t{1, 3},
			[]int{1, 1},
		},
	} {

		for jj, ppp := range []*Poly{
			s.p, s.p.Neg().(*Poly),
			s.p.subsXinv(), s.p.subsXinv().Neg().(*Poly)} {

			// fmt.Printf("TestSymSqfr(%d,%d) ppp=%v ===========================\n", ii, jj, ppp)

			sqfr := cad.sym_sqfr2(ppp, cell)
			var q RObj = one
			hasErr := false
			for i, sq := range sqfr {
				//fmt.Printf("<%d> [%v]^%d\n", i, sq.p, sq.r)
				if sq.r != s.r[i] {
					t.Errorf("<%d,%d> r[%d] expect=%d actual=%d\nret=%v", ii, jj, i, s.r[i], sq.r, sq.p)
					hasErr = true
					return
				}
				if len(sq.p.c)-1 != s.d[i] {
					t.Errorf("<%d> d[%d] expect=%d\nactual=%d\nret=%v", ii, i, s.d[i], sq.p, sq)
					hasErr = true
					break
				}

				qq := sq.p.Pow(NewInt(int64(sq.r)))
				q = Mul(qq, q)
			}
			if hasErr {
				break
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

func TestSymGcdMod(t *testing.T) {
	cad := new(CAD)
	cad.root = NewCell(cad, nil, 0)
	cad.rootp = NewCellmod(cad.root)
	cell0 := NewCell(cad, nil, 1)
	cell0.lv = 0
	cell0.parent = cad.root
	cell0.defpoly = NewPolyCoef(0, -2, 0, 1) // x^2-2
	cell1 := NewCell(cad, nil, 1)
	cell1.lv = 1
	cell1.parent = cell0
	cell1.defpoly = NewPolyCoef(1, -1, -2, 1) // y^2-2*y-1

	for ii, s := range []struct {
		f, g   *Poly
		expect *Poly
		celf   bool
		cell   *Cell
		p      Uint
	}{
		{
			NewPolyCoef(2, -5, 0, 1),
			NewPolyCoef(2, NewPolyVar(0), -1),
			nil, false,
			cell0, 151,
		}, {
			NewPolyCoef(2, 1, 0, -5),
			NewPolyCoef(2, NewPolyVar(0), -1),
			nil, false,
			cell0, 151,
		}, {
			NewPolyCoef(2, 1, 0, -5),
			NewPolyCoef(2, NewPolyVar(0), -1),
			nil, false,
			cell1, 151,
		}, {
			NewPolyCoef(2, -2, 0, 1),
			NewPolyCoef(2, NewPolyVar(0), -1),
			NewPolyCoef(2, NewPolyVar(0).Neg(), +1), false,
			cell0, 151,
		}, {
			NewPolyCoef(2, 5, NewPolyCoef(0, 1, 1), 1),
			NewPolyCoef(2, 5, NewPolyVar(1), 1),
			nil, true,
			cell0, 151,
		}, {
			NewPolyCoef(2, 5, NewPolyCoef(0, 1, 1), 3),
			NewPolyCoef(2, 5, NewPolyVar(1), 3),
			nil, true,
			cell0, 151,
		}, {
			NewPolyCoef(2, -2, 0, NewPolyCoef(0, 1, 1)),
			NewPolyCoef(2, -2, 0, NewPolyVar(1)),
			nil, true,
			cell0, 151,
		},
	} {
		fp := s.f.mod(s.p).(*Poly)
		gp := s.g.mod(s.p).(*Poly)
		fmt.Printf("<%d>===TestSymGcdMod() ======================================\nf=%v\ng=%v\n", ii, s.f, s.g)
		cellp, ok := cell1.mod(cad, s.p)
		if !ok {
			t.Errorf("not ok ii=%d", ii)
			continue
		}

		gcd, a, b := cad.symde_gcd_mod(fp, gp, cellp, s.p, true)
		if (gcd == nil) != (s.expect == nil) {
			t.Errorf("invalid gcd <a1, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\na=%v\nb=%v\n",
				ii, s.p, s.f, s.g, s.expect, gcd, a, b)
			continue
		}
		if gcd != nil {
			if gcd.lv != fp.lv || gcd.deg() != s.expect.deg() {
				t.Errorf("invalid gcd <a2, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\na=%v\nb=%v\n",
					ii, s.p, s.f, s.g, s.expect, gcd, a, b)
				continue
			}

			dega := a.deg()
			degb := b.deg()

			if degb > fp.deg()-gcd.deg() || dega > gp.deg()-gcd.deg() {
				t.Errorf("invalid gcd <a3, %d, %d>: %d, %d\nf=%v -> %v\ng=%v -> %v\nexpect=%v\nactual=%v\na=%v\nb=%v\n",
					ii, s.p, dega, degb, s.f, fp, s.g, gp, s.expect, gcd, a, b)
				return
			}

			gg := fp.mul_mod(a, s.p).add_mod(gp.mul_mod(b, s.p), s.p)
			if !gcd.Equals(gg) {
				t.Errorf("invalid gcd <a4, %d, %d>: %d, %d\nf=%v, g=%v\nfp=%v\ngp=%v\nexpect=%v\nactual=%v\ngg    =%v\nfpa=%v\ngpb=%v -> %v\na=%v\nb=%v\n",
					ii, s.p, dega, degb, s.f, s.g, fp, gp, s.expect, gcd, gg,
					fp.mul_mod(a, s.p), gp.mul_mod(b, s.p),
					fp.mul_mod(a, s.p).add_mod(gp.mul_mod(b, s.p), s.p),
					a, b)
				return
			}

		}
		if (a == nil) != s.celf {
			t.Errorf("invalid gcd <a4, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\na=%v\nb=%v\n",
				ii, s.p, s.f, s.g, s.expect, gcd, a, b)
			continue
		}
		if a == nil {
			if cellp.factor1 == nil {
				t.Errorf("invalid gcd <a5, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\na=%v\nb=%v\n",
					ii, s.p, s.f, s.g, s.expect, gcd, a, b)
				continue
			}
			if cellp.factor2 != nil && (cellp.factor1.lv != cellp.factor2.lv || cellp.factor1.deg() != cellp.factor2.deg()) {
				t.Errorf("invalid gcd <a6, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\na=%v\nb=%v\n",
					ii, s.p, s.f, s.g, s.expect, gcd, a, b)
				continue
			}
		}

		gcd3, a3, b3 := cad.symde_gcd_mod(gp, fp, cellp, s.p, true)
		if gcd3 == nil {
			if gcd != nil {
				t.Errorf("invalid gcd <b1, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v :: %v\na=%v\nb=%v\n",
					ii, s.p, s.f, s.g, s.expect, gcd3, gcd, a3, b3)
				continue

			}

		} else if (gcd == nil) != (gcd3 != nil) && !gcd3.Equals(gcd) && !gcd3.add_mod(gcd, s.p).IsZero() {
			t.Errorf("invalid gcd <b2, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v :: %v\na=%v\nb=%v\n",
				ii, s.p, s.f, s.g, s.expect, gcd3, gcd, a3, b3)
			continue
		}
		if a3 != nil && (b3 == nil || a3.deg() != b.deg() || b3.deg() != a.deg()) {
			t.Errorf("invalid gcd <b3, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v :: %v\na=%v :: %v\nb=%v :: %v\n",
				ii, s.p, s.f, s.g, s.expect, gcd3, gcd, a3, b, b3, a)
			continue

		}

		gcd2, a2, _ := cad.symde_gcd_mod(fp, gp, cellp, s.p, false)
		if (gcd == nil) != (gcd2 == nil) {
			t.Errorf("invalid gcd <c1, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\n",
				ii, s.p, s.f, s.g, s.expect, gcd)
			continue
		}
		if gcd != nil && !gcd.Equals(gcd2) {
			t.Errorf("invalid gcd <c2, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\n",
				ii, s.p, s.f, s.g, s.expect, gcd)
			continue
		}
		if a != nil && !a.Equals(a2) {
			t.Errorf("invalid gcd <c3, %d, %d>\nf=%v, g=%v\nexpect=%v\nactual=%v\n",
				ii, s.p, s.f, s.g, s.expect, gcd)
			continue
		}
	}
}

func TestSymGcd(t *testing.T) {
	cad := new(CAD)
	cad.root = NewCell(cad, nil, 0)
	cad.rootp = NewCellmod(cad.root)
	cell0 := NewCell(cad, nil, 1)
	cell0.lv = 0
	cell0.parent = cad.root
	cell0.defpoly = NewPolyCoef(0, -2, 0, 1) // x^2-2
	cell1 := NewCell(cad, nil, 1)
	cell1.lv = 1
	cell1.parent = cell0
	cell1.defpoly = NewPolyCoef(1, -1, -2, 1) // y^2-2*y-1: y = 1 +- x

	for ii, s := range []struct {
		f, g   *Poly
		expect *Poly
	}{
		{
			NewPolyCoef(2, NewPolyCoef(0, -2, 1), -1, 1), // z^2-z+x-2
			NewPolyCoef(2, NewPolyCoef(0, -2, 3), -3, 1), // z^2-3*z+3*x-2
			NewPolyCoef(2, NewPolyCoef(0, 0, -1), 1),
		},
	} {
		fmt.Printf("<%d>===TestSymGcd() ======================================\nf=%v\ng=%v\n", ii, s.f, s.g)
		cell0 := NewCell(cad, nil, 1)
		cell0.lv = 0
		cell0.parent = cad.root
		cell0.defpoly = NewPolyCoef(0, -2, 0, 1) // x^2-2
		cell1 := NewCell(cad, nil, 1)
		cell1.lv = 1
		cell1.parent = cell0
		cell1.defpoly = NewPolyCoef(1, -1, -2, 1) // y^2-2*y-1: y = 1 +- x

		gcd, _ := cad.symde_gcd2(s.f, s.g, cell1, 0)

		if !gcd.Equals(s.expect) {
			t.Errorf("i=%d\nf  =%v\ng  =%v\nexp=%v\nact=%v", ii, s.f, s.g, s.expect, gcd)
		}
	}
}
