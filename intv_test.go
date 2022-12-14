package ganrac

import (
	"math/big"
	"testing"
)

func TestIntvOp(t *testing.T) {

	pairs := [][]int{{0, 2}, {0, 3}, {1, 2}, {1, 3}}
	u := new(big.Float)
	u.SetPrec(50)

	for _, s := range [][]int64{
		{+1931, 1951, +733, +755},
		{+1931, 1951, -733, +755},
		{+1931, 1951, -755, -733},
		{-1931, 1951, -733, +755},
	} {
		for _, prec := range []uint{5, 8, 10} {
			x := newInterval(prec)
			x.inf.SetInt64(s[0])
			x.sup.SetInt64(s[1])

			y := newInterval(prec)
			y.inf.SetInt64(s[2])
			y.sup.SetInt64(s[3])

			for _, z := range []*Interval{
				x.Add(y).(*Interval),
				y.Add(x).(*Interval),
				x.Neg().Add(y.Neg()).Neg().(*Interval),
				y.Neg().Add(x.Neg()).Neg().(*Interval)} {
				for _, idx := range pairs {
					u.SetMode(big.ToPositiveInf)
					u.SetInt64(s[idx[0]] + s[idx[1]])
					if u.Cmp(z.sup) > 0 {
						t.Errorf("add.sup s=%v, idx=%v, x=%v, y=%v, z=%v\n", s, idx, x, y, z)
						break
					}

					u.SetMode(big.ToNegativeInf)
					u.SetInt64(s[idx[0]] + s[idx[1]])
					if u.Cmp(z.inf) < 0 {
						t.Errorf("add.inf s=%v, idx=%v, x=%v, y=%v, z=%v\n", s, idx, x, y, z)
						break
					}
				}
			}

			for _, z := range []*Interval{
				x.Mul(y).(*Interval),
				y.Mul(x).(*Interval),
				x.Mul(y.Neg()).Neg().(*Interval),
				y.Mul(x.Neg()).Neg().(*Interval),
				x.Neg().Mul(y).Neg().(*Interval),
				y.Neg().Mul(x).Neg().(*Interval),
				x.Neg().Mul(y.Neg()).(*Interval),
				y.Neg().Mul(x.Neg()).(*Interval),
			} {
				for _, idx := range pairs {
					u.SetMode(big.ToPositiveInf)
					u.SetInt64(s[idx[0]] * s[idx[1]])
					if u.Cmp(z.sup) > 0 {
						t.Errorf("mul.sup s=%v, idx=%v, x=%v, y=%v, z=%v\n", s, idx, x, y, z)
						break
					}

					u.SetMode(big.ToNegativeInf)
					u.SetInt64(s[idx[0]] * s[idx[1]])
					if u.Cmp(z.inf) < 0 {
						t.Errorf("mul.inf s=%v, idx=%v, x=%v, y=%v, z=%v\n", s, idx, x, y, z)
						break
					}
				}
			}

			for _, z := range []struct {
				v      *Interval
				i0, i1 int
			}{
				{x.Sub(y).(*Interval), 0, 1},
				{y.Sub(x).(*Interval), 1, 0},
				{x.Neg().Sub(y.Neg()).(*Interval), 1, 0},
				{y.Neg().Sub(x.Neg()).(*Interval), 0, 1},
			} {
				for _, idx := range pairs {
					u.SetMode(big.ToPositiveInf)
					u.SetInt64(s[idx[z.i0]] - s[idx[z.i1]])
					if u.Cmp(z.v.sup) > 0 {
						t.Errorf("sub.sup s=%v, idx=%v, x=%v, y=%v, z=%v\n", s, idx, x, y, z)
						break
					}

					u.SetMode(big.ToNegativeInf)
					u.SetInt64(s[idx[z.i0]] - s[idx[z.i1]])
					if u.Cmp(z.v.inf) < 0 {
						t.Errorf("sub.inf s=%v, idx=%v, x=%v, y=%v, z=%v\n", s, idx, x, y, z)
						break
					}
				}
			}
		}
	}
}

func TestIntvSubst(t *testing.T) {
	prec := uint(20)
	lv := Level(0)

	// [10] F;
	// -x^3-1301*x^2-301300*x-300000
	// [11] subst(F,x,x-1001.778320);
	// -x^3+1704.33*x^2-705352*x+1.24896e+06
	// [12] subst(F,x,3.595703*x-1001.778320);
	// -46.4891*x^3+22035.5*x^2-2.53624e+06*x+1.24896e+06

	p := NewPolyCoef(lv, -300000, -301300, -1301, -1).toIntv(prec).(*Poly)

	low := newInterval(prec)
	low.SetFloat64(-1001.778320)

	oneone := newInterval(prec)
	oneone.SetInt64(1)

	q := p.Subst(NewPolyCoef(lv, low, oneone), lv).(*Poly)
	if q.c[0].Sign() <= 0 {
		t.Errorf("why?")
		return
	}
}
