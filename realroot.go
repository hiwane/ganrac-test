package ganrac

// real root isolation by using Descartes rule of signs
import (
	"fmt"
	"sort"
)

type dcsr struct {
	lb, rb NObj
	m      int
	p      *dcsr // parent
}

func (p *dcsr) String() string {
	return fmt.Sprintf("[%d:%f, %f]", p.m, p.lb.Float(), p.rb.Float())
}

type dcsr_stack struct {
	v   []*dcsr
	n   int // len of self.v
	ret []*dcsr
}

func newDcsrStack(deg int) *dcsr_stack {
	v := new(dcsr_stack)
	v.v = make([]*dcsr, 0, 1)
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

func (s *dcsr_stack) addret(lb, rb NObj, n int) {
	s.ret = append(s.ret, &dcsr{lb, rb, n, nil})
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

func (p *Poly) convertRange(lb, ub NObj) *Poly {
	var q *Poly
	q = p.subst1(NewPolyCoef(p.lv, lb, one), p.lv).(*Poly)
	q = q.subst1(NewPolyCoef(p.lv, zero, ub.Sub(lb)), p.lv).(*Poly)
	for i := 0; i < len(q.c)/2; i++ {
		t := q.c[i]
		q.c[i] = q.c[len(q.c)-i-1]
		q.c[len(q.c)-i-1] = t
	}
	q = q.subst1(NewPolyCoef(p.lv, one, one), p.lv).(*Poly)
	return q
}

func evalRange(stack *dcsr_stack, sp *dcsr, p *Poly, lb, ub NObj) int {
	q := p.convertRange(lb, ub)
	n := q.descartesSignRules()
	if len(q.c) != len(p.c) { // 左端点がゼロ点だった
		n++
	}
	if n > 1 {
		dc := &dcsr{lb, ub, n, sp}
		stack.push(dc)
	} else if n == 1 {
		if len(q.c) != len(p.c) {
			stack.addret(lb, lb, n)
		} else {
			stack.addret(lb, ub, n)
		}
	}
	return n
}

func realRootImprove(p *Poly, sp *dcsr) {
	if sp.lb == sp.rb {
		return
	}
	var m NObj
	m = sp.lb.Add(sp.rb).Div(two).(NObj) // 中点
	v := p.subst1(m, p.lv)
	sgn := v.Sign()
	if sgn == sp.m {
		sp.lb = m
	} else if sgn != 0 {
		sp.rb = m
	} else {
		sp.lb = m
		sp.rb = m
	}
}

func (p *Poly) RealRootIsolation(prec int) (*List, error) {
	if !p.isUnivariate() {
		return nil, fmt.Errorf("not a unirariate polynomial")
	}

	rb := p.rootBound2Exp() // bint に変更したい @TODO

	stack := newDcsrStack(len(p.c))
	// x=0 は事前に取り除く.
	i := 0
	for ; p.c[i].IsZero(); i++ {
	}
	if i > 0 {
		stack.addret(zero, zero, i)
		q := NewPoly(p.lv, len(p.c)-i)
		copy(q.c, p.c[i:])
		p = q
	}

	// p = sqrt(p)

	// x < 0 の処理
	q := p.subst1(NewPolyInts(p.lv, 0, -1), p.lv).(*Poly)
	nn := q.descartesSignRules()
	if nn > 0 {
		if nn == 1 {
			stack.addret(rb.Neg().(NObj), zero, nn)
		} else {
			dc := &dcsr{rb.Neg().(NObj), zero, nn, nil}
			stack.push(dc)
		}
	}

	// x > 0 の処理
	np := p.descartesSignRules()
	if np > 0 {
		if np == 1 {
			stack.addret(zero, rb, np)
		} else {
			dc := &dcsr{zero, rb, np, nil}
			stack.push(dc)
		}
	}

	counter := 0
	for stack.n > 0 { // 各区間の実根の個数が 1 になるまで分割
		counter += 1
		sp := stack.pop()
		mid := sp.lb.Add(sp.rb).Div(NewInt(2)).(NObj)

		// 区間の左半分
		n := evalRange(stack, sp, p, sp.lb, mid)
		if n == 1 && sp.m == 2 {
			// 確定
			if p.subst1(mid, p.lv).Sign() == 0 {
				stack.addret(mid, mid, n)
			} else {
				stack.addret(mid, sp.rb, n)
			}
		} else {
			// 区間の右半分
			evalRange(stack, sp, p, mid, sp.rb)
		}
	}

	// 端点が 0 は困る
	for _, r := range stack.ret {
		if r.lb.Equals(r.rb) {
			// ゼロ点
			r.m = 0
			r.lb = r.rb
			continue
		}

		sgn := p.subst1(r.lb, p.lv).Sign()
		sgnr := p.subst1(r.rb, p.lv).Sign()
		var lb NObj
		if sgn != 0 {
			r.m = sgn // 左端点の符号を設定
			if r.lb.IsZero() {
				lb = r.lb
			}
		} else {
			r.m = -sgnr
			lb = r.lb
		}
		for r.lb == lb {
			realRootImprove(p, r)
		}

		for (sgnr == 0 || r.rb.IsZero()) && r.lb != r.rb {
			realRootImprove(p, r)
			sgnr = p.subst1(r.rb, p.lv).Sign()
		}
	}

	ret := stack.ret
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].lb.Cmp(ret[j].lb) < 0
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
			if !r.lb.IsZero() {
				sgn *= -1
			}
		}
	}

	// 重複を除去
	for i := 1; i < len(ret); i++ {
		for ret[i-1].rb.Cmp(ret[i].lb) >= 0 {
			realRootImprove(p, ret[i-1])
			realRootImprove(p, ret[i])
		}
	}

	r := make([]interface{}, len(ret))
	for i := 0; i < len(ret); i++ {
		r[i] = NewList([]interface{}{ret[i].lb, ret[i].rb})
	}

	return NewList(r), nil
}
