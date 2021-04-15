package ganrac

import (
	"fmt"
	"math/big"
	"testing"
)

func TestImprove(t *testing.T) {
	for _, s := range []struct {
		p        []*big.Float
		inf, sup float64
		prec     uint
		expect   *big.Float
	}{
		{
			[]*big.Float{
				big.NewFloat(+2),
				big.NewFloat(-4),
				big.NewFloat(1)},
			3.30078125, 3.421875, 10,
			big.NewFloat(3.41421356),
		},
	} {
		x := newInterval(s.prec)
		x.inf.SetFloat64(s.inf)
		x.sup.SetFloat64(s.sup)
		fmt.Printf("x=%v\n", x)

		kraw := newIKraw(s.prec, s.p)
		o := kraw.improve(x)
		fmt.Printf("o=%v\n", o)

		if o.inf.Cmp(s.expect) > 0 || o.sup.Cmp(s.expect) < 0 {
			t.Errorf("x=%v, p=%v\nexpect=%v\nactual=%v\n", x, s.p, s.expect, o)
		}
	}
}

func TestFRealRoot(t *testing.T) {
	for _, s := range []struct {
		p    []*big.Float
		prec uint
		root []*big.Float
	}{
		{
			[]*big.Float{
				big.NewFloat(-2),
				big.NewFloat(0),
				big.NewFloat(1)},
			10,
			[]*big.Float{
				big.NewFloat(-1.41421356),
				big.NewFloat(+1.41421356)},
		}, {
			[]*big.Float{
				big.NewFloat(+2),
				big.NewFloat(-4),
				big.NewFloat(1)},
			10,
			[]*big.Float{
				big.NewFloat(2 - 1.41421356),
				big.NewFloat(2 + 1.41421356)},
		}, {
			[]*big.Float{
				big.NewFloat(+6),
				big.NewFloat(0),
				big.NewFloat(-5),
				big.NewFloat(0),
				big.NewFloat(1)},
			50,
			[]*big.Float{
				big.NewFloat(1.414213562373095048801688),
				big.NewFloat(+1.7320508075688772935274463415058723669428),
				big.NewFloat(-1.7320508075688772935274463415058723669428),
				big.NewFloat(-1.414213562373095048801688),
			},
		}, {
			[]*big.Float{
				big.NewFloat(-6),
				big.NewFloat(11),
				big.NewFloat(-6),
				big.NewFloat(1)},
			50,
			[]*big.Float{
				big.NewFloat(1),
				big.NewFloat(2),
				big.NewFloat(3)},
		}, {
			[]*big.Float{
				big.NewFloat(12),
				big.NewFloat(-4),
				big.NewFloat(-3),
				big.NewFloat(1)},
			10,
			[]*big.Float{
				big.NewFloat(-2),
				big.NewFloat(2),
				big.NewFloat(3)},
		}, {
			[]*big.Float{
				big.NewFloat(-120),
				big.NewFloat(154),
				big.NewFloat(49),
				big.NewFloat(-140),
				big.NewFloat(70),
				big.NewFloat(-14),
				big.NewFloat(1)},
			50,
			[]*big.Float{
				big.NewFloat(-1),
				big.NewFloat(1),
				big.NewFloat(2),
				big.NewFloat(3),
				big.NewFloat(4),
				big.NewFloat(5)},
		},
	} {
		fmt.Printf("########## goo! %v, roots=%v\n", s.p, s.root)
		o := FRealRoot(s.prec, 100000, s.p, 0.1234)
		if o == nil || len(o) != len(s.root) {
			t.Errorf("prec=%d\ninput=%v\nexpect=%v\nactual=%v", s.prec, s.p, s.root, o)
			return
		}

		// 根はそれぞれの区間に一度だけ含まれる.
		for _, r := range s.root {
			cnt := 0
			for _, oo := range o {
				if oo.inf.Cmp(r) <= 0 && r.Cmp(oo.sup) <= 0 {
					cnt++
				}
			}
			if cnt != 1 {
				t.Errorf("prec=%d, r=%v, cnt=%d\ninput=%v\nexpect=%v\nactual=%v", s.prec, r, cnt, s.p, s.root, o)
				return
			}
		}
	}
}
