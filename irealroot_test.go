package ganrac

import (
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func TestIRootCountRand(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		deg := rand.Int31()%10 + 5
		p := NewPoly(0, int(deg)+1)
		for i := 0; i < len(p.c); i++ {
			p.c[i] = NewInt((rand.Int63() % 2000) - 1000)
		}
		p = p.normalize().(*Poly)
		bq, err := p.RootBound()
		if err != nil {
			t.Errorf("Q rootbound error: %v", err)
		}

		prec := uint(5)
		q := p.toIntv(prec).(*Poly)
		bi, _, _ := q.iRootBound(prec)

		bqf := new(big.Float)
		bqf.SetPrec(100)
		bqf.SetMode(big.ToPositiveInf)
		switch bq_ := bq.(type) {
		case *Int:
			bqf.SetInt(bq_.n)
		case *Rat:
			bqf.SetRat(bq_.n)
		default:
			t.Errorf("unexpected return `%v`", bq)
			continue
		}

		if bi.Cmp(bqf) < 0 {
			t.Errorf("smallllll q=`%v`=`%v`, i=%v\ninput=%v\n     =%v", bq, bqf, bi, p, q)
		}
	}
}

func TestIRealRoot1(t *testing.T) {
	NewGANRAC()
	for i, s := range []struct {
		p      *Poly
		prec   uint
		expect []float64
	}{
		{
			NewPolyInts(0, -300000, 301300, -1301, 1),
			10,
			[]float64{1, 1000, 300},
		}, {
			NewPolyInts(0, -6, 12, -1, -4, 1),
			10,
			[]float64{1.7320508, -1.7320508, 2 + 1.41421356, 2 - 1.41421356},
		},
	} {
		for j, pp := range []struct {
			p   *Poly
			sgn float64
		}{
			{s.p, 1},
			{s.p.subst1(NewPolyInts(s.p.lv, 0, -1), s.p.lv).(*Poly), -1},
			{s.p.Neg().(*Poly), 1},
			{s.p.subst1(NewPolyInts(s.p.lv, 0, -1), s.p.lv).Neg().(*Poly), -1},
		} {
			p := pp.p.toIntv(s.prec).(*Poly)
			o, err := p.iRealRoot(s.prec + 10)
			if err != nil {
				t.Errorf("[%d,%d] err=%v\ninput =%v\nexpect=%v\nactual=%v\n", i, j, err, pp.p, s.expect, o)
				return
			}
			if len(o) != len(s.expect) {
				t.Errorf("\ninput =%v\nexpect=%v\nactual=%v\n", pp.p, s.expect, o)
				return
			}

			for _, root := range s.expect {
				f := big.NewFloat(root * pp.sgn)
				cnt := 0
				for _, oo := range o {
					if oo.inf.Cmp(f) <= 0 && f.Cmp(oo.sup) <= 0 {
						cnt++
					}
				}
				if cnt != 1 {
					t.Errorf("cnt=%d, root=%f\ninput =%v\nexpect=%v*%f\nactual=%v\n", cnt, root, pp.p, s.expect, pp.sgn, o)
					return
				}
			}
		}
	}

}
