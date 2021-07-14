package ganrac

// real root isolation by using Descartes rule of signs
//
// user binary interval
// lower bound = (n + 0) * 2^m
// upper bound = (n + 1) * 2^m

import (
	"fmt"
	"sort"
)

type dcsr struct {
	low   *BinInt
	m     int
	p     *dcsr // parent
	point bool
}

func (p *dcsr) toIntv(prec uint) *Interval {
	x := p.low.toIntv(prec).(*Interval)
	if !p.point {
		xi := p.low.upperBound().toIntv(prec).(*Interval)
		x.sup = xi.sup
	}
	return x
}

func (p *dcsr) String() string {
	upp := p.low.upperBound()
	if p.point {
		return fmt.Sprintf("[%d:%f*]", p.m, p.low.Float())
	} else {
		return fmt.Sprintf("[%d:%f, %f]", p.m, p.low.Float(), upp.Float())
	}
}

func (p *dcsr) upperBound() *BinInt {
	if p.point {
		return p.low
	} else {
		return p.low.upperBound()
	}
}

type dcsr_stack struct {
	v   []*dcsr
	n   int // len of self.v
	ret []*dcsr
}

func newDcsrStack(deg int) *dcsr_stack {
	v := new(dcsr_stack)
	v.v = make([]*dcsr, 0)
	v.ret = make([]*dcsr, 0, deg)
	return v
}

func (s *dcsr_stack) pop() *dcsr {
	s.n -= 1
	p := s.v[s.n]
	s.v[s.n] = nil
	return p
}

func (s *dcsr_stack) push(p *dcsr) {
	if s.n < len(s.v) {
		s.v[s.n] = p
	} else {
		s.v = append(s.v, p)
	}
	s.n++
}

func (s *dcsr_stack) addret(lb *BinInt, n int, point bool) {
	s.ret = append(s.ret, &dcsr{lb, n, nil, point})
}

func (z *Poly) descartesSignRules() int {
	// z is univariate polynomial.
	// returns (# of sign variation of z(x), # ... of z(-x))
	d := len(z.c) - 1 // 次数-1
	sgnp := z.c[d].Sign()
	// sgnn := sgnp * (2 * (d % 2) - 1)
	np := 0

	for d--; d >= 0; d-- {
		sgn := z.c[d].Sign()
		if sgn == 0 {
			continue
		}
		if sgnp != sgn {
			sgnp = sgn
			np++
		}
	}
	return np
}

func (q *Poly) subsXinv() *Poly {
	// if f(0)!=0: return x^n f(1/x)
	// if f(0)==0: return x^(n-1) f(1/x)
	// where n = deg(q)
	m := 0
	if q.c[m].IsZero() {
		m = 1
	}
	qq := NewPoly(q.lv, len(q.c)-m)
	for i := m; i < len(q.c); i++ {
		qq.c[i-m] = q.c[len(q.c)-1-i+m]
	}
	return qq
}

func (p *Poly) convertRange(low *BinInt) *Poly {
	var q *Poly
	if low.m >= 0 {
		lb := newInt()
		lb.n.Lsh(low.n, uint(low.m))
		h := newInt()
		h.n.Lsh(one.n, uint(low.m))
		q = p.Subst(NewPolyCoef(p.lv, lb, h), p.lv).(*Poly)
	} else {
		c := new(Int)
		c.n = low.n
		m := uint(-low.m)
		q = p.subst_binint_1var(c, m).(*Poly)
		for i := 0; i < len(q.c)-1; i++ {
			q.c[i] = q.c[i].(NObj).mul_2exp(m * uint(len(q.c)-i-1))
		}

		if err := q.valid(); err != nil {
			panic(err)
		}
	}
	q = q.subsXinv()
	if err := q.valid(); err != nil {
		fmt.Printf("p=%v\n", p)
		fmt.Printf("q=%v\n", q)
		panic(err)
	}
	q = q.Subst(NewPolyCoef(p.lv, one, one), p.lv).(*Poly)
	// fmt.Printf("q_=%v\n", q)

	return q
}

func evalRange(stack *dcsr_stack, sp *dcsr, p *Poly, lb *BinInt) int {
	q := p.convertRange(lb)
	n := q.descartesSignRules()
	if len(q.c) != len(p.c) { // 左端点がゼロ点だった
		n++
	}
	if n > 1 {
		dc := &dcsr{lb, n, sp, false}
		stack.push(dc)
	} else if n == 1 {
		stack.addret(lb, n, len(q.c) != len(p.c))
	}
	return n
}

func realRootImprove(p *Poly, sp *dcsr) {
	if sp.point {
		return
	}
	m := sp.low.midBinIntv() // 中点
	v := p.Subst(m, p.lv)
	sgn := v.Sign()
	if sgn == sp.m {
		sp.low = m
	} else if sgn != 0 {
		sp.low = sp.low.halveIntv() // 左端点はそのままで区間幅を半分に
	} else {
		// 点になった.
		sp.low = m
		sp.point = true
	}
}

func (p *Poly) realRootIsolation(prec int) []*dcsr {
	rb := p.rootBoundBinInt()

	/////////////////////////////////////
	// 準備
	/////////////////////////////////////
	stack := newDcsrStack(len(p.c))
	// x=0 は事前に取り除く.
	i := 0
	for ; p.c[i].IsZero(); i++ {
	}
	if i > 0 {
		if len(p.c) == i+1 {
			// p=x^n
			ret := make([]*dcsr, 1)
			ret[0] = &dcsr{newBinInt(), i, nil, true}
			return ret
		}

		stack.addret(newBinInt(), i, true)
		q := NewPoly(p.lv, len(p.c)-i)
		copy(q.c, p.c[i:])
		p = q
	}

	// x < 0 の処理
	q := p.Subst(NewPolyCoef(p.lv, 0, -1), p.lv).(*Poly)
	nn := q.descartesSignRules()
	if nn > 0 {
		if nn == 1 {
			stack.addret(rb.Neg().(*BinInt), nn, false)
		} else {
			dc := &dcsr{rb.Neg().(*BinInt), nn, nil, false}
			stack.push(dc)
		}
	}

	// x > 0 の処理
	np := p.descartesSignRules()
	if np > 0 {
		z := newBinInt()
		z.m = rb.m
		if np == 1 {
			stack.addret(z, np, false)
		} else {
			dc := &dcsr{z, np, nil, false}
			stack.push(dc)
		}
	}

	/////////////////////////////////////
	// 本番
	/////////////////////////////////////
	counter := 0
	for stack.n > 0 { // 各区間の実根の個数が 1 になるまで分割
		counter += 1
		sp := stack.pop()
		mid := sp.low.midBinIntv()

		// 区間の左半分 (low, mid)
		low := newBinInt()
		low.n.Mul(sp.low.n, two.n)
		low.m = sp.low.m - 1
		n := evalRange(stack, sp, p, low)
		if n == 1 && sp.m == 2 {
			// 全体で 2 個だった, 左半分で 1個見つかったので右半分 1個確定
			stack.addret(mid, n, p.Subst(mid, p.lv).Sign() == 0)
		} else {
			// 区間の右半分
			evalRange(stack, sp, p, mid)
		}
	}

	/////////////////////////////////////
	// 後処理.
	/////////////////////////////////////
	// 端点がゼロ点は困る
	for _, r := range stack.ret {
		if r.point {
			// ゼロ点
			r.m = 0
			continue
		}

		sgn := p.Subst(r.low, p.lv).Sign()
		ub := r.low.upperBound()
		sgnr := p.Subst(ub, p.lv).Sign()
		var lb NObj
		if sgn != 0 {
			r.m = sgn // 左端点の符号を設定
			if r.low.IsZero() {
				lb = r.low
			}
		} else {
			r.m = -sgnr
			lb = r.low
		}
		for r.low == lb {
			realRootImprove(p, r)
		}

		// 右端がゼロ点か x=0
		for (sgnr == 0 || r.low.Sign() < 0 && r.low.n.BitLen() == 1) && !r.point {
			realRootImprove(p, r)
			ub := r.low.upperBound()
			sgnr = p.Subst(ub, p.lv).Sign()
		}

		// 精度十分?
		for !r.point && r.low.m > prec {
			realRootImprove(p, r)
		}
	}

	ret := stack.ret
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].low.Cmp(ret[j].low) < 0
	})

	// テスト
	if false {
		// 左端点の符号は根があるたびに入れ替わる
		sgn := p.c[len(p.c)-1].Sign()
		if len(p.c)%2 == 0 { // 奇数次
			sgn *= -1
		}
		fmt.Printf("p=%v, sgn=%d\n", p, sgn)
		for _, r := range ret {
			if sgn != r.m && r.m != 0 {
				fmt.Printf("sgn=%d, range=%v\n", sgn, r)
				fmt.Printf("p=%v\n", p)
				for _, r := range ret {
					fmt.Printf(" range=%v\n", r)
				}
				panic("gey")
			}
			if !r.low.IsZero() {
				sgn *= -1
			}
		}
	}

	// 重複を除去
	for i := 1; i < len(ret); i++ {
		ub := ret[i-1].upperBound()
		j := 0
		for ub.Cmp(ret[i].low) >= 0 {
			realRootImprove(p, ret[i-1])
			realRootImprove(p, ret[i])
			ub = ret[i-1].upperBound()
			j++
			if j >= 10 {
				panic("stop")
			}
		}
	}

	return ret
}

func (p *Poly) RealRootIsolation(prec int) (*List, error) {
	if !p.isUnivariate() {
		return nil, fmt.Errorf("not a unirariate polynomial")
	}

	ret := p.realRootIsolation(prec)

	r := NewList()
	for i := 0; i < len(ret); i++ {
		ub := ret[i].upperBound()
		r.Append(NewList(ret[i].low, ub))
	}

	return r, nil
}
