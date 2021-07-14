package ganrac

// real root isolation of a univariate polynomial with interval coefficient
import (
	"fmt"
	"math/big"
)

func (p *Poly) iRootBound(prec uint) (*big.Float, uint, uint) {
	// p: univariate poly. with interval coeff.
	// returns (bound, # of positive roots, # of negative roots)
	posnum := uint(0)
	negnum := uint(0)

	sgnp := p.c[len(p.c)-1].Sign()
	sgnn := sgnp
	if len(p.c)%2 == 0 { // 奇数次数
		sgnn = -sgnp
	}

	bound := new(big.Float)
	for i := len(p.c) - 2; i >= 0; i-- {
		c := p.c[i].(*Interval)
		s := c.Sign()

		if s > 0 {
			if bound.Cmp(c.sup) < 0 {
				bound = c.sup
			}
		} else {
			vv := new(big.Float)
			vv.SetPrec(prec)
			vv.SetMode(big.ToPositiveInf)
			vv.Neg(c.inf)
			if bound.Cmp(vv) < 0 {
				bound = vv
			}
		}

		if s*sgnp < 0 {
			posnum++
			sgnp = s
		}
		if i%2 == 1 {
			s = -s
		}
		if s*sgnn < 0 {
			negnum++
			sgnn = s
		}
	}

	c := p.c[len(p.c)-1].(*Interval)
	boundx := new(big.Float)
	boundx.SetPrec(prec)
	boundx.SetMode(big.ToPositiveInf)
	if c.Sign() > 0 {
		bound = boundx.Quo(bound, c.inf)
	} else {
		lc := new(big.Float)
		lc.SetPrec(prec)
		lc.SetMode(big.ToPositiveInf)
		lc.Neg(c.sup)
		bound = boundx.Quo(bound, lc)
	}

	bf_one := new(big.Float)
	bf_one.SetInt64(1)
	bound.Add(bound, bf_one)

	return bound, posnum, negnum
}

func (p *Poly) has_only_one_root(x *Interval, prec uint) bool {
	// 区間 x に含まれる区間多項式 p の数が唯一であるか確かめる.
	// assume: deg(p) > 1

	// 微分一定なら....
	v := p.c[len(p.c)-1].(*Interval).MulInt64(int64(len(p.c) - 1))
	for i := len(p.c) - 2; i > 0; i-- {
		v = Add(Mul(v, x), p.c[i].(*Interval).MulInt64(int64(i))).(*Interval)
	}
	if !v.ContainsZero() {
		return true
	}

	// デカルト
	// (inf, sup) -> (0, sup-inf) -> (0, 1)
	a := newInterval(prec)
	a.inf.Set(x.inf)
	a.sup.Set(x.inf)
	b := newInterval(prec)

	b.inf.Set(x.sup)
	b.sup.Set(x.sup)

	b.inf.Sub(b.inf, x.inf)
	b.sup.Sub(b.sup, x.inf)

	cc := newInterval(prec)
	cc.inf.SetInt64(1)
	cc.sup.SetInt64(1)

	q := p.Subst(NewPolyCoef(p.lv, a, b), p.lv).(*Poly)
	if len(q.c) != len(p.c) {
		return false
	}

	// (0, 1) -> (1, +inf)
	for i := 0; i < len(q.c)/2; i++ {
		c := q.c[i]
		q.c[i] = q.c[len(p.c)-i-1]
		q.c[len(p.c)-i-1] = c
	}

	a = newInterval(prec)
	a.sup.SetInt64(1)
	a.inf.SetInt64(1)
	q = q.Subst(NewPolyCoef(p.lv, a, a), p.lv).(*Poly)
	if len(q.c) != len(p.c) {
		return false
	}

	// 符号変化の数を数える.
	c := q.c[0]
	s := c.Sign()
	if s == 0 {
		// 定数項が 0 を保持.
		return false
	}

	n := 0
	undetermined := false
	for i := 1; i < len(q.c); i++ {
		sn := q.c[i].Sign()
		if sn == 0 { // 符号不明
			if undetermined {
				return false
			}
			undetermined = false
		} else {
			if s*sn < 0 {
				// [+ ? -], [+ -] で 1増える.
				n++
				undetermined = false
			} else if undetermined {
				// [+ ? +] とかで，符号変化の数が確定できない.
				return false
			}
			s = sn
		}
	}

	return n == 1 && !undetermined
}

func (p *Poly) iRealRoot(prec uint, lmax int) ([]*Interval, error) {
	// return error if
	// - p の区間が広いか
	// - p が重複因子をもつか
	// - 精度がたりないか (prec が小さい)
	if err := p.valid(); err != nil {
		fmt.Printf("iRealRoot() input is invalid %v\n", err)
		panic("stop")
	}
	// fmt.Printf("iRealRoot(%v) start!\n", p)

	bound, posp, posn := p.iRootBound(prec)

	sup := make([]*big.Float, len(p.c))
	inf := make([]*big.Float, len(p.c))

	for i, c := range p.c {
		sup[i] = c.(*Interval).sup
		inf[i] = c.(*Interval).inf
	}

	roots := make([][]*Interval, 2)

	x := newInterval(prec)
	x.sup.Set(bound)

	ret := make([]*Interval, 0, posn+posp)
	if posp > 0 {
		// 正の根を探す.
		kraw := newIKraw(prec, sup)
		x := newInterval(prec)
		x.sup.Set(bound)
		roots[0] = kraw.fRealRoot(x, lmax)
		if roots[0] == nil {
			return nil, fmt.Errorf("pos0 prec")
		}

		kraw = newIKraw(prec, inf)
		x = newInterval(prec)
		x.sup.Set(bound)
		roots[1] = kraw.fRealRoot(x, lmax)

		if roots[1] == nil {
			return nil, fmt.Errorf("pos1 prec")
		}
		if len(roots[0]) != len(roots[1]) || int(posp) < len(roots[0]) || int(posp)%2 != len(roots[0])%2 {
			return nil, fmt.Errorf("pos root=(%d,%d)/%d", len(roots[0]), len(roots[1]), posp)
		}

		w := 0
		if p.c[len(p.c)-1].Sign() > 0 {
			w = 1
		}

		retp := make([]*Interval, len(roots[0]))
		for i := len(roots[0]) - 1; i >= 0; i-- {
			vv := newInterval(prec)
			vv.sup.Set(roots[0+w][i].sup)
			vv.inf.Set(roots[1-w][i].inf)
			if err := vv.valid(); err != nil { // @TODO
				fmt.Printf("roots[%d] 0=%v, 1=%v -> %v\n", i, roots[0][i], roots[1][i], vv)
				return nil, err
			}
			retp[i] = vv
			w = 1 - w
		}

		// 区間が重複していないこと.
		for i := 1; i < len(retp); i++ {
			if retp[i-1].sup.Cmp(retp[i].inf) >= 0 {
				return nil, fmt.Errorf("pos root overlap. [%d,%d] %v <%v,%v>", i-1, i, retp,
					retp[i-1].sup, retp[i].inf)
			}
		}

		// 区間に根が唯一しかないこと
		if posp > 2 && len(retp) < int(posp) {
			for i, rr := range retp {
				if !p.has_only_one_root(rr, prec) {
					return nil, fmt.Errorf("pos: not one root %d:%v", i, retp)
				}
			}
		}
		ret = retp
	}

	if posn > 0 {
		// 負の根を探す.
		for i := 1; i < len(p.c); i += 2 {
			f := new(big.Float)
			f.SetPrec(prec)
			f.Neg(p.c[i].(*Interval).inf)
			sup[i] = f

			f = new(big.Float)
			f.SetPrec(prec)
			f.Neg(p.c[i].(*Interval).sup)
			inf[i] = f
		}

		kraw := newIKraw(prec, sup)
		roots[0] = kraw.fRealRoot(x, lmax)
		if roots[0] == nil {
			return nil, fmt.Errorf("neg0 prec")
		}

		kraw = newIKraw(prec, inf)
		roots[1] = kraw.fRealRoot(x, lmax)
		if roots[1] == nil {
			return nil, fmt.Errorf("neg1 prec")
		}

		if len(roots[0]) != len(roots[1]) || int(posn) < len(roots[0]) || int(posn)%2 != len(roots[0])%2 {
			return nil, fmt.Errorf("neg root=(%d,%d)/%d", len(roots[0]), len(roots[1]), posn)
		}

		w := 0
		if (len(p.c)%2*2-1)*p.c[len(p.c)-1].Sign() > 0 {
			w = 1
		}

		retp := make([]*Interval, len(roots[0]))
		for i := len(roots[0]) - 1; i >= 0; i-- {
			if err := roots[0][i].valid(); err != nil { // @TODO
				fmt.Printf("roots[0][%d]=%v\n", i, roots[0][i])
				panic(err.Error())
			}
			if err := roots[1][i].valid(); err != nil { // @TODO
				fmt.Printf("roots[1][%d]=%v\n", i, roots[1][i])
				panic(err.Error())
			}

			vv := newInterval(prec)
			vv.inf.Neg(roots[0+w][i].sup)
			vv.sup.Neg(roots[1-w][i].inf)
			retp[len(roots[0])-i-1] = vv
			w = 1 - w
			if err := vv.valid(); err != nil { // @TODO
				fmt.Printf("P1 neg. roots[%d/%d,%d:%d] %v\n", i, len(roots[w]), p.c[len(p.c)-1].Sign(), len(p.c), err)
				fmt.Printf("%%e: %e %e\n", roots[w][i], roots[1-w][i])
				fmt.Printf("%%f: %f %f\n", roots[w][i], roots[1-w][i])
				fmt.Printf("%%f: %.30f\n", vv)
				fmt.Printf("%%f: %.30f\n", roots[1-w][i])
				panic(err)
			}
		}

		// 区間が重複していないこと.
		for i := 1; i < len(retp); i++ {
			if retp[i-1].sup.Cmp(retp[i].inf) >= 0 {
				return nil, fmt.Errorf("neg root overlap. [%d,%d] %v <%v,%v>", i-1, i, retp,
					retp[i-1].sup, retp[i].inf)
			}
		}

		// 区間に根が唯一しかないこと
		if posn > 2 && len(retp) < int(posn) {
			for i, rr := range retp {
				if !p.has_only_one_root(rr, prec) {
					return nil, fmt.Errorf("neg: not one root %d:%v", i, retp)
				}
			}
		}

		if ret == nil {
			ret = retp
		} else {
			ret = append(retp, ret...)
		}
	}

	// fmt.Printf("iRealRoot(%v) end!\n", p)
	return ret, nil
}

/*
has_only_one_root() x=[998.1826172,1001.77832]
  p[3]=[1,1]*x^3[-1302,-1300]*x^2+[301056,301568]*x[-300032,-299520]
  q1[4]=[46.4375,46.5625]*x^3+[21824,21984]*x^2+[2097152,8388608]*x[-524288,-262144]
  q2[4]=[-524288,-262144]*x^3+[-0,8388608]*x^2+[524288,16777216]*x+[32,64]
  sign(q[0])=1
  descates x=[998.1826172,1001.77832]
    i=1, sn=1->1, n=0, m=false
    i=2, sn=1->0, n=0, m=false
    i=3, sn=1->-1, n=0, m=false
    n=1, q=[-524288,-262144]*x^3+[-0,8388608]*x^2+[524288,16777216]*x+[32,64]


has_only_one_root() x=[-1001.77832,-998.1826172]
  p[3]=[-1,-1]*x^3[-1302,-1300]*x^2[-301568,-301056]*x[-300032,-299520]
  q1[4]=[-46.5625,-46.4375]*x^3+[21984,22112]*x^2[-8388608,-2097152]*x[-524288,-262144]
  q2[4]=[-524288,-262144]*x^3[-67108864,-2097152]*x^2[-33554432,-1048576]*x[-64,-32]
  sign(q[0])=-1
  descates x=[-1001.77832,-998.1826172]
    i=1, sn=-1->-1, n=0, m=false
    i=2, sn=-1->-1, n=0, m=false
    i=3, sn=-1->-1, n=0, m=false
    n=0, q=[-524288,-262144]*x^3[-67108864,-2097152]*x^2[-33554432,-1048576]*x[-64,-32]
*/
