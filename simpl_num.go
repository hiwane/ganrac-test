package ganrac

// a symbolic-numeric method for formula simplification
// 数値数式手法による論理式の簡単化, 岩根秀直, JSSAC 2017

// @TODO y^2-x<=0 && z^2+2*x*z+(-x+1)*y^2-2*x*y-x<=0 && x-1>=0 で -1<y<1 は真

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

// nil は empty set を表す.
type NumRegion struct {
	// len(r[lv]) == 0 は  -inf <= x[lv] <= +inf を表す.
	// もし，要素があれば, union of closed interval
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
	sep := ""
	for lv, vv := range m.r {
		fmt.Fprintf(s, "%s%s:[", sep, varstr(lv))
		for i, v := range vv {
			if i != 0 {
				fmt.Fprintf(s, ",")
			}
			v.Format(s, format)
		}
		fmt.Fprintf(s, "]")
		sep = ", "
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
	} else if m.inf == nil {
		return false
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

func (_ *NumRegion) maxInt(n, m int) int {
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
		return n.r[lv]
	}
	nx, ok := n.r[lv]
	if !ok {
		return mx
	}

	ret := make([]*ninterval, 0, m.maxInt(len(mx), len(nx)))
	i := 0
	j := 0
	for i < len(mx) && j < len(nx) {
		if m.less(mx[i], nx[j]) {
			// [m .... m]
			//                 [n ..... n]
			if mx[i].sup != nil && nx[j].inf != nil && mx[i].sup.Cmp(nx[j].inf) <= 0 {
				// 重複なし.
				i++
				continue
			}

			x := new(ninterval)
			x.inf = nx[j].inf
			if mx[i].sup == nil || (nx[j].sup != nil) && mx[i].sup.Cmp(nx[j].sup) > 0 {
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
			if nx[j].sup != nil && mx[i].inf != nil && nx[j].sup.Cmp(mx[i].inf) <= 0 {
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
	} else if !okn || len(nx) == 0 {
		return mx
	} else if !okm || len(mx) == 0 {
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
		if a.inf != nil && r.sup != nil && r.sup.Cmp(a.inf) <= 0 {
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
	return append(ret, r)
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
	if m == nil {
		return m
	}
	for _, q := range qq {
		if _, ok := m.r[q]; ok {
			delete(m.r, q)
		}
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
func (poly *Poly) simplNumUniPoly(t, f *NumRegion) (OP, *NumRegion, *NumRegion) {
	// 重根を持っていたら...?
	roots := poly.realRootIsolation(-30)
	// fmt.Printf("   simplNumUniPoly(%v) t=%v, f=%v, #root=%v\n", poly, t, f, len(roots))
	if len(roots) == 0 { // 符号一定
		if poly.Sign() > 0 {
			return GT, newNumRegion(), nil
		} else {
			return LT, nil, newNumRegion()
		}
	}
	xs := t.getU(f, poly.lv)
	prec := uint(53)

	// 根が， unknown 領域に含まれていない，かつ
	// 同じ領域に入るなら符号一定が確定する
	if len(xs) > 0 { // T union F に対してやったほうが楽か.
		idx := -1
		rooti := roots[0].toIntv(prec)
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
				// 重根がないと仮定しているので，idx で符号が確定できる.
				if 0 < idx && idx < len(xs) {
					if poly.deg()%2 == 0 {
						// 偶数次
						if poly.Sign() > 0 {
							return GT, newNumRegion(), nil
						} else {
							return LT, nil, newNumRegion()
						}
					} else {
						// 奇数次
						pinf := newNumRegion()
						pinf.r[poly.lv] = append(pinf.r[poly.lv], &ninterval{roots[len(roots)-1].low.upperBound(), nil})
						ninf := newNumRegion()
						ninf.r[poly.lv] = append(ninf.r[poly.lv], &ninterval{nil, roots[0].low})
						if poly.Sign() > 0 {
							return OP_TRUE, pinf, ninf
						} else {
							return OP_TRUE, ninf, pinf
						}
					}
				}

				sgn := poly.Sign()
				if idx != 0 {
					// x = -inf
					sgn *= 2*(len(poly.c)%2) - 1
				}
				if sgn > 0 {
					return GT, newNumRegion(), nil
				} else if sgn < 0 {
					return LT, nil, newNumRegion()
				}
				panic("?")
			}
		}
	}

	// gray region を代入
	x := xs[0]
	x.sup = xs[len(xs)-1].sup

	p := poly.toIntv(prec).(*Poly)
	pp := p.SubstIntv(x, p.lv, prec).(*Interval)
	if ss := pp.Sign(); ss > 0 {
		return GT, newNumRegion(), nil
	} else if ss < 0 {
		return LT, nil, newNumRegion()
	}

	nr := []*NumRegion{newNumRegion(), newNumRegion()}

	var inf NObj
	for i, intv := range roots {
		nr[i%2].add(inf, intv.low, poly.lv)
		if intv.point {
			inf = intv.low
		} else {
			inf = intv.low.upperBound()
		}
	}
	nr[len(roots)%len(nr)].add(inf, nil, poly.lv)

	sgn := poly.Sign()
	sgn *= (len(poly.c)%2)*2 - 1 // x=-inf での poly の符号
	if sgn > 0 {
		return OP_TRUE, nr[0], nr[1]
	} else {
		return OP_TRUE, nr[1], nr[0]
	}
}

func (poly *Poly) simplNumNvar(g *Ganrac, t, f *NumRegion, dv Level) (OP, *NumRegion, *NumRegion) {
	prec := uint(30)
	p := poly.toIntv(prec).(*Poly)

	for lv := poly.lv; lv >= 0; lv-- {
		if lv == dv {
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

	// fmt.Printf("   simplNumNvar() p[%d]=%f\n", dv, p)
	if len(p.c) == 2 { // linear
		if p.c[1].(*Interval).ContainsZero() {
			return OP_TRUE, t, f
		}
		// x := p.c[0].Div(p.c[1].Neg().(NObj)).(*Interval)
		// if ss := x.Sign(); ss > 0 {
		// } else if ss < 0 {
		// } else if x.inf.Sign() == 0 {
		// } else if x.sup.Sign() == 0 {
		// } else {
		// }
		return OP_TRUE, t, f
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

func (poly *Poly) simplNumPoly(g *Ganrac, t, f *NumRegion, dv Level) (OP, *NumRegion, *NumRegion) {
	if poly.isUnivariate() {
		return poly.simplNumUniPoly(t, f)
	}
	pret := newNumRegion()
	nret := newNumRegion()
	op_ret := OP_TRUE
	for v := poly.lv + 1; v >= 0; v-- {
		deg := poly.Deg(v)
		if deg == 0 && v < poly.lv {
			continue
		}
		s, pos, neg := poly.simplNumNvar(g, t, f, v)
		op_ret &= s
		if op_ret == GT || op_ret == LT {
			return op_ret, pos, neg
		}
		pret = pret.union(pos)
		nret = nret.union(neg)
		if deg != 2 || v > dv {
			// \sum_i x_i^2 = 1 のようなケースで, 判別式爆発が起こる.
			// v > dv は，多少でも削減するために限定する
			continue
		}

		// 2次であれば，判別式が負なら符号が主係数の符号と一致することを利用する.
		c2 := poly.Coef(v, 2)
		if cp, ok := c2.(*Poly); ok {
			var fml Fof
			atom := NewAtom(cp, GT).simplFctr(g)
			fml, pos, _ = atom.simplNum(g, t, f)
			if _, ok := fml.(*AtomT); ok {
				s = GT
				neg = nil
			} else {
				atom = NewAtom(cp, LT).simplFctr(g)
				fml, neg, _ = atom.simplNum(g, t, f)
				if _, ok := fml.(*AtomT); ok {
					s = LT
				} else {
					s = OP_TRUE
				}
			}
		} else if c2.Sign() > 0 {
			s = GT
			pos = newNumRegion()
			neg = nil
		} else {
			s = LT
			pos = nil
			neg = newNumRegion()
		}
		c1 := poly.Coef(v, 1)
		c0 := poly.Coef(v, 0)
		d := Sub(c1.Mul(c1), Mul(c2, c0).Mul(four))
		var fml Fof
		var neg2 *NumRegion
		if dd, ok := d.(NObj); ok {
			if dd.Sign() >= 0 {
				continue
			}
			fml = trueObj
			neg2 = newNumRegion()
		} else {
			discrim := Sub(c1.Mul(c1), Mul(c2, c0).Mul(four)).(*Poly) // b^2-4ac
			atom := NewAtom(discrim, LT)
			atom = atom.simplFctr(g)

			fml, neg2, _ = atom.simplNum(g, t, f)
			// fmt.Printf("simnum: discrim[%d]=%v: %v: neg=%v: %v\n", dv, d, atom, neg2, fml)
		}
		switch fml.(type) {
		case *AtomT: // 符号一定が確定
			if s == GT {
				return s, newNumRegion(), nil
			} else if s == LT {
				return s, nil, newNumRegion()
			}
		case *AtomF:
			continue
		}

		pos = neg2.intersect(pos) // 判別式が負 かつ 主係数が正
		pret = pret.union(pos)

		neg = neg2.intersect(neg) // 判別式が負 かつ 主係数が負
		nret = nret.union(neg)
		// fmt.Printf("-- neg2=%v\n", neg2)
		// fmt.Printf("-- pos=%v -> %v\n", pos, pret)
		// fmt.Printf("-- neg=%v -> %v\n", neg, nret)
	}

	return OP_TRUE, pret, nret
}

////////////////////////////////////////////////////////////////////////
// simplNum
////////////////////////////////////////////////////////////////////////

func (p *AtomT) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	return p, t, f
}

func (p *AtomF) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	return p, t, f
}

func (atom *Atom) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	// simplFctr 通過済みと仮定したいところだが.
	if atom.op == NE || atom.op == GE || atom.op == GT {
		p := atom.Not()
		p, f, t = p.simplNum(g, f, t)
		return p.Not(), t, f
	}

	ps := make([]*Poly, 0, len(atom.p))
	op := atom.op
	if atom.op == EQ {
		up := false
		var neg *NumRegion
		for _, p := range atom.p {
			s, pp, nn := p.simplNumPoly(g, t, f, p.lv)
			if s == GT || s == LT {
				up = true
				continue
			}
			neg = neg.union(pp)
			neg = neg.union(nn)
			ps = append(ps, p)
		}
		if len(ps) == 0 {
			return falseObj, t, f
		}
		if up {
			atom = newAtoms(ps, op)
		}
		return atom, nil, neg
	}
	if len(atom.p) > 1 {
		// 一多項式じゃないとやってられない....
		var pmul RObj = one

		for _, p := range atom.p {
			s, _, _ := p.simplNumPoly(g, t, f, p.lv)
			if s != GT && s != LT {
				ps = append(ps, p)
				pmul = p.Mul(pmul)
			} else if s == LT {
				op = op.neg()
			}
		}
		a, t, f := NewAtom(pmul, op).simplNum(g, t, f)
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
	pol := atom.getPoly()
	s, pp, nn := pol.simplNumPoly(g, t, f, pol.lv+1)
	// fmt.Printf("   atm.simplNum(): s=%v|%v, atom=%v, pos=%f, neg=%f\n", s, atom.op, pol, pp, nn)

	// op は　LT or LE
	if s == OP_TRUE {
		// 簡単化できず
		return atom, nn, pp
	} else if s&atom.op == 0 {
		return falseObj, nil, newNumRegion()
	} else if s|atom.op == atom.op {
		fmt.Printf("true!\n")
		return trueObj, newNumRegion(), nil
	} else {
		return atom, nn, pp
	}
}

func (p *FmlAnd) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	ts := make([]*NumRegion, 0, len(p.fml))
	fs := make([]*NumRegion, 0, len(p.fml))
	fmls := make([]Fof, 0, len(p.fml))
	// fmt.Printf("   And.simplNum And=%v\n", p)
	for i := range p.fml {
		fml, tt, ff := p.fml[i].simplNum(g, t, f)
		// fmt.Printf("@@ And.simplNum[1st,%d/%d] %v -> %v, %v\n", i+1, len(p.fml), p.fml[i], fml, ff)
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
	// fmt.Printf("   And.simplNum fmls=%v\n", fmls)
	if len(fmls) <= 1 {
		if len(fmls) == 0 {
			return trueObj, nil, nil
		}
		return fmls[0], ts[0], fs[0]
	}

	var tret *NumRegion
	fret := f
	for i, fml := range fmls {
		ff := f
		var tt *NumRegion
		for j := 0; j < len(fmls); j++ {
			if j != i {
				ff = ff.union(fs[j])
				//		tt = tt.intersect(ts[j])
			}
		}

		fmls[i], tt, ff = fml.simplNum(g, t, ff)
		if _, ok := fmls[i].(*AtomF); ok {
			return falseObj, nil, newNumRegion()
		}
		tret = tret.intersect(tt)
		fret = fret.union(ff)
		// fmt.Printf("@@ And.simplNum[2nd,%d/%d] %v, Fi=%v, Fret=%v\n", i+1, len(fmls), fmls[i], ff, fret)
	}
	tret = tret.union(t)
	fml := newFmlAnds(fmls...)
	// fmt.Printf("## And.simplNum[end,%d] %v\n", len(fmls), fret)
	return fml, tret, fret
}

func (p *FmlOr) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	// @TODO サボり
	q := p.Not()
	q, f, t = q.simplNum(g, f, t)
	return q.Not(), t, f
}

func (p *ForAll) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	fml, t, f := p.fml.simplNum(g, t, f)
	return NewQuantifier(true, p.q, fml), t.del(p.q), f.del(p.q)
}

func (p *Exists) simplNum(g *Ganrac, t, f *NumRegion) (Fof, *NumRegion, *NumRegion) {
	fml, t, f := p.fml.simplNum(g, t, f)
	return NewQuantifier(false, p.q, fml), t.del(p.q), f.del(p.q)
}
