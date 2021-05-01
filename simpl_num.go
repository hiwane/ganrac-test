package ganrac

// a symbolic-numeric method for formula simplification
// 数値数式手法による論理式の簡単化, 岩根秀直, JSSAC 2017

// NOTE:
// F=example("adam2-1")[0]; F;simpl(F); で落ちる

import (
	"fmt"
	"sort"
)

// open/closed interval
type ninterval struct {
	// inf=nil => -infinity
	inf NObj
	// sup=nil => +infinity
	sup NObj
}

// nil は empty を表す.
type NumRegion struct {
	// union of closed interval
	r map[Level][]*ninterval
}

////////////////////////////////////////////////////////////////////////
// ninterval
////////////////////////////////////////////////////////////////////////

func (x *ninterval) Format(s fmt.State, format rune) {
	fmt.Fprintf(s, "[")
	if x.inf == nil {
		fmt.Fprintf(s, "-inf")
	} else {
		x.inf.Format(s, format)
	}
	fmt.Fprintf(s, ",")
	if x.sup == nil {
		fmt.Fprintf(s, "+inf")
	} else {
		x.sup.Format(s, format)
	}
	fmt.Fprintf(s, "]")
}

////////////////////////////////////////////////////////////////////////
// NumRegion
////////////////////////////////////////////////////////////////////////

func newNumRegion() *NumRegion {
	f := new(NumRegion)
	f.r = make(map[Level][]*ninterval, 0)
	return f
}

func (m *NumRegion) add(inf, sup NObj, lv Level) {
	m.r[lv] = append(m.r[lv], &ninterval{inf, sup})
}

func (m *NumRegion) Format(s fmt.State, format rune) {
	if m == nil {
		fmt.Fprintf(s, "{}")
		return
	}
	fmt.Fprintf(s, "{")
	for lv, vv := range m.r {
		fmt.Fprintf(s, "%d: [", lv)
		for i, v := range vv {
			if i != 0 {
				fmt.Fprintf(s, ",")
			}
			v.Format(s, format)
		}
		fmt.Fprintf(s, "],")
	}
	fmt.Fprintf(s, "}")
}

// 左端点が小さい方，または
// 左端点が同じかつ，右端点が小さい方
func (_ *NumRegion) less(n, m *ninterval) bool {
	if n.inf == nil {
		if m.inf != nil {
			return true
		}
	} else {
		c := n.inf.Cmp(m.inf)
		if c < 0 {
			return true
		} else if c > 0 {
			return false
		}
	}
	// n.inf == m.inf

	if n.sup == nil {
		return false
	}
	if m.sup == nil {
		return true
	}
	return n.sup.Cmp(m.sup) < 0
}

func (nr *NumRegion) sort(nm []*ninterval) {
	sort.Slice(nm, func(i, j int) bool {
		return nr.less(nm[i], nm[j])
	})
}

func (_ *NumRegion) max(n, m int) int {
	if n < m {
		return m
	} else {
		return n
	}
}

// lv 限定で intersect 計算
func (m *NumRegion) intersect_lv(n *NumRegion, lv Level) []*ninterval {
	mx, ok := m.r[lv]
	if !ok {
		return nil
	}
	nx, ok := n.r[lv]
	if !ok {
		return nil
	}

	ret := make([]*ninterval, 0, m.max(len(mx), len(nx)))
	i := 0
	j := 0
	for i < len(mx) && j < len(nx) {
		if m.less(mx[i], nx[j]) {
			// [m .... m]
			//                 [n ..... n]
			if mx[i].sup != nil && mx[i].sup.Cmp(nx[j].inf) <= 0 {
				// 重複なし.
				i++
				continue
			}

			x := new(ninterval)
			x.inf = nx[j].inf
			if mx[i].sup == nil || mx[i].sup.Cmp(nx[j].sup) > 0 {
				// [m ....... m]
				//   [n...n]
				x.sup = nx[j].sup
				j++
			} else {
				// [m .... m]
				//     [n.......n]
				x.sup = mx[i].sup
				i++
			}
			ret = append(ret, x)
		} else {
			if nx[j].sup != nil && nx[j].sup.Cmp(mx[i].inf) <= 0 {
				// 重複なし.
				j++
				continue
			}

			x := new(ninterval)
			x.inf = nx[j].inf
			if nx[j].sup == nil || nx[j].sup.Cmp(mx[i].sup) > 0 {
				//   [m...m]
				// [n ....... n]
				x.sup = mx[i].sup
				j++
			} else {
				//     [m.......m]
				// [n .... n]
				x.sup = nx[j].sup
				i++
			}
			ret = append(ret, x)
		}
	}
	return ret
}

// 2つの領域の intersect を計算する
func (m *NumRegion) intersect(n *NumRegion) *NumRegion {
	if n == nil || m == nil {
		return nil
	}
	u := newNumRegion()
	for lv, _ := range m.r {
		v := m.intersect_lv(n, lv)
		if v != nil {
			u.r[lv] = v
		}
	}
	return u
}

// lv 限定で union 計算
func (m *NumRegion) union_lv(n *NumRegion, lv Level) []*ninterval {
	// assume: m != nil && n != nil
	mx, okm := m.r[lv]
	nx, okn := n.r[lv]
	if !okn && !okm {
		return nil
	} else if !okn {
		return mx
	} else if !okm {
		return nx
	}

	nm := make([]*ninterval, len(mx)+len(nx))
	copy(nm, mx)
	copy(nm[len(mx):], nx)
	sort.Slice(nm, func(i, j int) bool {
		return m.less(nm[i], nm[j])
	})

	r := nm[0]
	ret := make([]*ninterval, 0, len(nm))
	for _, a := range nm[1:] {
		// 重なりがあるか.
		if r.sup.Cmp(a.inf) <= 0 {
			// [r ..... r]
			//              [a  ..... a]
			ret = append(ret, r)
			r = a
		} else {
			// 合体.
			// [r ..... r]
			//       [a  ..... a]
			v := new(ninterval)
			v.inf = r.inf
			v.sup = a.sup
			r = v
		}
	}
	return ret
}

func (m *NumRegion) union(n *NumRegion) *NumRegion {
	if n == nil {
		return m
	}
	if m == nil {
		return n
	}
	u := newNumRegion()
	for lv := range m.r {
		u.r[lv] = m.union_lv(n, lv)
	}
	for lv, nx := range n.r {
		if _, ok := u.r[lv]; !ok {
			u.r[lv] = nx
		}
	}

	return u
}

func (m *NumRegion) del(qq []Level) *NumRegion {
	for _, q := range qq {
		delete(m.r, q)
	}
	return m
}

// trueRegion と falseRegion 以外の領域を返す.
func (m *NumRegion) getU(n *NumRegion, lv Level) []*Interval {
	prec := uint(53)

	var mn []*ninterval
	if m == nil && n == nil {
		// goto _R させてくれない
		mn = nil
	} else if n == nil {
		mn = m.r[lv]
	} else if m == nil {
		mn = n.r[lv]
	} else {
		// 補集合を広くとるので， union は狭く.
		// このときの union は境界は残さないといけない. (m.sup=n.inf)
		// [m ... m]
		//         [n .... n] なら
		// [m ....m] [n ...n] と別区間扱い
		mn = m.union_lv(n, lv)
	}
	if mn == nil {
		mn = []*ninterval{}
	}

	// mn の補集合
	var inf NObj
	inf = nil
	xs := make([]*Interval, 0, len(mn)+2)
	for _, y := range mn {
		if y.inf != nil {
			x := newInterval(prec)
			if inf == nil {
				x.inf.SetInf(true)
			} else {
				xi := inf.toIntv(prec).(*Interval)
				x.inf.Set(xi.inf)
			}
			xi := y.inf.toIntv(prec).(*Interval)
			x.sup = xi.sup
			xs = append(xs, x)
		}
		inf = y.sup
	}
	if inf != nil {
		x := newInterval(prec)
		xi := inf.toIntv(prec).(*Interval)
		x.inf.Set(xi.inf)
		x.sup.SetInf(false)
		xs = append(xs, x)
	} else if len(xs) == 0 {
		x := newInterval(prec)
		x.inf.SetInf(true)
		x.sup.SetInf(false)
		xs = append(xs, x)
	}
	return xs
}

////////////////////////////////////////////////////////////////////////
// simplNumPoly
////////////////////////////////////////////////////////////////////////

// assume: poly is univariate
// returns (OP, pos, neg)
// OP = (t,f) 以外で取りうる符号
func (poly *Poly) simplNumUni(op OP, t, f *NumRegion) (OP, *NumRegion, *NumRegion) {
	// 重根を持っていたら...?
	fmt.Printf("simplNumUni(%v) t=%v, f=%v\n", poly, t, f)
	roots := poly.realRootIsolation(-30)
	if len(roots) == 0 {
		if poly.Sign() > 0 {
			return GT, t, f
		} else {
			return LT, t, f
		}
	}
	xs := t.getU(f, poly.lv)
	prec := uint(53)

	// 根が， unknown 領域に含まれていない，かつ
	// 同じ領域に入るなら符号一定が確定する
	if len(xs) > 0 { // T union F に対してやったほうが楽か.
		idx := -1
		rooti := roots[0].toIntv(prec)
		fmt.Printf("xs=%v\n", xs)
		fmt.Printf("root=%v\n", roots)
		if xs[0].inf != nil && rooti.sup.Cmp(xs[0].inf) < 0 {
			idx = 0
		} else {
			for i := 1; i < len(xs); i++ {
				if xs[i-1].sup.Cmp(rooti.inf) < 0 && rooti.sup.Cmp(xs[i].inf) < 0 {
					idx = i
					break
				}
			}
			if idx < 0 && xs[len(xs)-1].sup.Cmp(rooti.inf) < 0 {
				idx = len(xs)
			}
		}
		fmt.Printf("rooti=%v, idx=%d\n", rooti, idx)

		if idx >= 0 {
			for _, root := range roots {
				rooti := root.toIntv(prec)
				if idx == 0 {
					//  root  ... xs[0]
					if rooti.sup.Cmp(xs[idx].inf) >= 0 {
						idx = -1
						break
					}
				} else if idx == len(xs) {
					// xs[-1] ... root
					if xs[idx-1].sup.Cmp(rooti.inf) >= 0 {
						idx = -1
						break
					}
				} else {
					// xs[idx-1] ... root ... xs[idx]
					if !(xs[idx-1].sup.Cmp(rooti.inf) < 0 && rooti.sup.Cmp(xs[idx].inf) < 0) {
						idx = -1
						break
					}
				}
			}

			if idx >= 0 {
				// 根が連結した known 領域のみに含まれていて,
				// unknown 領域での符号が一定であることが確定した
				var sgn int
				if idx == len(xs)-1 || idx == 0 && len(poly.c)%2 == 1 {
					sgn = poly.Sign()
				} else if idx == 0 {
					sgn = -poly.Sign()
				} else {
					x := roots[idx-1].upperBound().Add(roots[idx].low).Div(two)
					x = poly.subst1(x, poly.lv)
					sgn = x.Sign()
				}
				if sgn > 0 {
					return GT, nil, nil
				} else if sgn < 0 {
					return LT, nil, nil
				}
				panic("?")
			}
		}
	}

	if poly.Sign() < 0 {
		op = op.neg()
	}
	var nr []*NumRegion
	if op == EQ || op == NE {
		nr = []*NumRegion{newNumRegion()}
	} else {
		nr = []*NumRegion{newNumRegion(), newNumRegion()}
	}
	var inf NObj
	for i, intv := range roots {
		nr[i%len(nr)].add(inf, intv.low, poly.lv)
		if intv.point {
			inf = intv.low
		} else {
			inf = intv.low.upperBound()
		}
	}
	nr[len(roots)%len(nr)].add(inf, nil, poly.lv)
	switch op {
	case EQ:
		return OP_TRUE, nil, nr[0]
	case NE:
		return OP_TRUE, nr[0], nil
	case GT, GE:
		return OP_TRUE, nr[len(roots)%2], nr[1-len(roots)%2]
	case LT, LE:
		return OP_TRUE, nr[1-len(roots)%2], nr[len(roots)%2]
	default:
		panic("unknown")
	}
}

func (poly *Poly) simplNumNvar(op OP, t, f *NumRegion, dv Level) (OP, *NumRegion, *NumRegion) {
	prec := uint(30)
	p := poly.toIntv(prec).(*Poly)
	for lv := poly.lv; lv >= 0; lv-- {
		if lv != dv {
			continue
		}
		xs := t.getU(f, lv)
		x := xs[0]
		x.sup = xs[len(xs)-1].sup
		switch pp := p.SubstIntv(x, lv, prec).(type) {
		case *Poly:
			p = pp
		case *Interval:
			// 区間 u での符号がきまった
			fmt.Printf("simplNumNvar() interval=%f\n", pp)
			if ss := pp.Sign(); ss > 0 {
				goto _GT
			} else if ss < 0 {
				goto _LT
			} else if pp.inf.Sign() == 0 {
				goto _GE
			} else if pp.sup.Sign() == 0 {
				goto _LE
			}
			return OP_TRUE, t, f
		}
	}

	fmt.Printf("simplNumNvar() p[%d]=%f\n", dv, p)
	if len(p.c) == 2 { // linear
		if p.c[1].(*Interval).ContainsZero() {
			return OP_TRUE, t, f
		}
		x := p.c[0].Div(p.c[1].Neg().(NObj)).(*Interval)
		if ss := x.Sign(); ss > 0 {
			goto _GT
		} else if ss < 0 {
			goto _LT
		} else if x.inf.Sign() == 0 {
			goto _GE
		} else if x.sup.Sign() == 0 {
			goto _LE
		} else {
			return OP_TRUE, t, f
		}
	}

	return OP_TRUE, t, f
_GT:
	return GT, t, f
_GE:
	return GE, t, f
_LE:
	return LE, t, f
_LT:
	return LT, t, f
}

func (poly *Poly) simplNumPoly(op OP, t, f *NumRegion, dv Level) (OP, *NumRegion, *NumRegion) {
	if poly.isUnivariate() {
		return poly.simplNumUni(op, t, f)
	}
	tret := newNumRegion()
	fret := newNumRegion()
	for v := poly.lv; v >= 0; v-- {
		deg := poly.Deg(v)
		if deg == 0 {
			continue
		}
		s, t, f := poly.simplNumNvar(op, t, f, v)
		if s != OP_TRUE {
			if s&op == 0 {
				return s, nil, nil
			} else if s&op.not() == 0 {
				return s, nil, nil
			}
		}
		tret = tret.union(t)
		fret = fret.union(f)
		if deg != 2 || v > dv {
			// \sum_i x_i^2 = 1 のようなケースで, 判別式爆発が起こる.
			// v > dv は，多少でも削減するために限定する
			continue
		}

		// 2次であれば，判別式が負なら符号が主係数の符号と一致することを利用する.
		c2 := poly.Coef(v, 2)
		if cp, ok := c2.(*Poly); ok {
			s, _, _ = cp.simplNumPoly(NE, t, f, dv)
			if s != LT && s != GT {
				continue
			}
		} else if c2.Sign() > 0 {
			s = GT
		} else {
			s = LT
		}
		c1 := poly.Coef(v, 1)
		c0 := poly.Coef(v, 0)
		discrim := Sub(c1.Mul(c1), Mul(c2, c0).Mul(four)).(*Poly) // b^2-4ac
		fmt.Printf("in=%v\n", poly)
		fmt.Printf("discrim[%d]=%v\n", v, discrim)
		sx, t, f := discrim.simplNumPoly(LE, t, f, v)
		fmt.Printf("discrim[%d]=%x\n", v, sx)
		if sx == LT {
			// poly の符号が確定した.
			return s, nil, nil
		}
	}

	return OP_TRUE, tret, fret
}

////////////////////////////////////////////////////////////////////////
// simplNum
////////////////////////////////////////////////////////////////////////

func (p *AtomT) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	return p, t, f
}

func (p *AtomF) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	return p, t, f
}

func (atom *Atom) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	// simplFctr 通過済みと仮定したいところだが.
	if atom.op == NE || atom.op == GE || atom.op == GT {
		p := atom.Not()
		p, f, t = p.simplNum(t, f)
		fmt.Printf("Atm.simplNum(NEG) p=%v\n", p)
		return p.Not(), t, f
	}

	ts := make([]*NumRegion, 0, len(atom.p))
	fs := make([]*NumRegion, 0, len(atom.p))
	ps := make([]*Poly, 0, len(atom.p))
	op := atom.op
	if atom.op == EQ {
		up := false
		for _, p := range atom.p {
			s, tt, ff := p.simplNumPoly(NE, t, f, p.lv)
			if s == GT || s == LT {
				up = true
				continue
			}
			ts = append(ts, tt)
			fs = append(fs, ff)
			ps = append(ps, p)
		}
		if len(ps) == 0 {
			return falseObj, t, f
		}
		for i, _ := range ps {
			t = t.intersect(ts[i])
			f = f.union(fs[i])
		}
		if up {
			atom = newAtoms(ps, op)
		}
		return atom, t, f
	}
	if len(atom.p) > 1 {
		// 一変数じゃないとやってられない....
		var pmul RObj = one

		for _, p := range atom.p {
			s, _, _ := p.simplNumPoly(NE, t, f, p.lv)
			if s != GT && s != LT {
				ps = append(ps, p)
				pmul = p.Mul(pmul)
			} else if s == LT {
				op = op.neg()
			}
		}
		a, t, f := NewAtom(pmul, op).simplNum(t, f)
		switch a.(type) {
		case *AtomT, *AtomF:
			return a, t, f
		default:
			if len(ps) == len(atom.p) {
				return atom, t, f
			} else if len(ps) == 1 {
				return a, t, f
			} else {
				return newAtoms(ps, op), t, f
			}
		}
	}
	s, t, f := atom.p[0].simplNumPoly(op, t, f, atom.p[0].lv)
	fmt.Printf("atm.simplNum(): s=%v, t=%f, f=%f\n", s, t, f)
	if s == OP_TRUE {
		return atom, t, f
	} else if s&atom.op == 0 {
		return falseObj, t, f
	} else if s&atom.op.not() == 0 {
		return trueObj, t, f
	} else {
		return atom, t, f
	}
}

func (p *FmlAnd) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	ts := make([]*NumRegion, 0, len(p.fml))
	fs := make([]*NumRegion, 0, len(p.fml))
	fmls := make([]Fof, 0, len(p.fml))
	fmt.Printf("And.simplNum And=%v\n", p)
	for i := range p.fml {
		fml, tt, ff := p.fml[i].simplNum(t, f)
		fmt.Printf("And.simplNum[%d] %v -> %v\n", i, p.fml[i], fml)
		if _, ok := fml.(*AtomF); ok {
			return falseObj, nil, nil
		}
		if _, ok := fml.(*AtomT); ok {
			continue
		}
		fmls = append(fmls, fml)
		ts = append(ts, tt)
		fs = append(fs, ff)
	}
	fmt.Printf("And.simplNum fmls=%v\n", fmls)
	if len(fmls) <= 1 {
		if len(fmls) == 0 {
			return trueObj, nil, nil
		}
		return fmls[0], ts[0], fs[0]
	}
	for i, fml := range fmls {
		ff := f
		tt := t
		for j := 0; j < len(fmls); j++ {
			if j != i {
				ff = ff.union(fs[j])
				tt = tt.intersect(fs[j])
			}
		}
		tt = tt.union(t)
		fmt.Printf("And.simplNum(%d) tt=%e, ff=%e\n", i, tt, ff)
		fmls[i], ts[i], fs[i] = fml.simplNum(tt, ff)
		if _, ok := fmls[i].(*AtomF); ok {
			return falseObj, nil, nil
		}
	}
	tt := t
	ff := f
	for i, _ := range fmls {
		tt = tt.intersect(ts[i])
		ff = ff.union(fs[i])
	}
	fml := newFmlAnds(fmls...)
	return fml, tt, ff
}

func (p *FmlOr) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	// @TODO サボり
	q := p.Not()
	q, f, t = q.simplNum(f, t)
	return q.Not(), t, f
}

func (p *ForAll) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	fml, t, f := p.fml.simplNum(t, f)
	return NewQuantifier(true, p.q, fml), t.del(p.q), f.del(p.q)
}

func (p *Exists) simplNum(t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	fml, t, f := p.fml.simplNum(t, f)
	return NewQuantifier(false, p.q, fml), t.del(p.q), f.del(p.q)
}
