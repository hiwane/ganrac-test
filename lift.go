package ganrac

import (
	"fmt"
	"math/big"
	"os"
)

func (cell *Cell) set_truth_value_from_children(cad *CAD) {
	// 子供の真偽値から自分の真偽値を決める.
	if cad.q[cell.lv+1] < 0 {
		panic("gao")
	}

	cell.truth = 1 - cad.q[cell.lv+1]
	for _, c := range cell.children {
		if cad.q[cell.lv+1] == c.truth {
			cell.truth = c.truth
			break
		}
	}
}

func (cell *Cell) set_truth_other() {
	// 自分の真偽値が確定したから，
	// 子供の真偽値を other に設定してしまう.
	for _, c := range cell.children {
		if c.truth < 0 {
			c.truth = t_other
			if c.children != nil {
				c.set_truth_other()
			}
		}
	}

}

func (cell *Cell) set_parent_and_truth_other(cad *CAD) {
	// 自分の真偽値が確定したから，
	// 決められるなら親と
	// 子供の真偽値を other に設定してしまう.

	// 親の真偽値設定
	c := cell
	for ; c.lv >= 0 && cad.q[c.lv] == cell.truth; c = c.parent {
		c.parent.truth = cell.truth
	}

	c.set_truth_other()
}

func (cad *CAD) Lift(index ...int) error {
	fmt.Printf("cad.Lift %v\n", index)
	if len(index) == 0 {
		for !cad.stack.empty() {
			cell := cad.stack.pop()
			if cell.truth >= 0 {
				continue
			} else if cell.children != nil {
				// 子供の真偽値が確定した.
				cell.set_truth_value_from_children(cad)
				continue
			} else {
				cell.lift(cad)
			}
		}
		return nil
	}
	c := cad.root
	if len(index) == 1 && index[0] == -1 {
		if c.children == nil {
			c.lift(cad)
		} else {
			return fmt.Errorf("already lifted %v", index)
		}
		return nil
	}

	for _, idx := range index {
		if c.children == nil || idx < 0 || idx >= len(c.children) {
			return fmt.Errorf("invalid index %v", index)
		}
		c = c.children[idx]
	}
	if c.children == nil {
		c.lift(cad)
		return nil
	} else {
		return fmt.Errorf("already lifted %v", index)
	}
}

func (cell *Cell) Index() []uint {
	idx := make([]uint, cell.lv+1)
	c := cell
	for c.lv >= 0 {
		idx[c.lv] = c.index
		c = c.parent
	}
	return idx
}

func (cell *Cell) lift(cad *CAD) error {
	fmt.Printf("lift (%v)\n", cell.Index())
	cad.stat.lift++
	ciso := make([][]*Cell, len(cad.proj[cell.lv+1].pf))
	signs := make([]sign_t, len(ciso))
	cs := make([]*Cell, 0)
	for i := 0; i < len(ciso); i++ {
		ciso[i], signs[i] = cell.make_cells(cad, cad.proj[cell.lv+1].pf[i])
		cs = append(cs, ciso[i]...)
	}

	// merge して
	cs = cad.cellsort(cs, true)

	// sector 作って
	cs = cad.addSector(cell, cs)

	// for i := 0; i < len(cs); i++ {
	// 	fmt.Printf("cs[%d]=<%e,%e> <%v,%v> %v\n", i, cs[i].intv.l.Float(), cs[i].intv.u.Float(),
	// 		cs[i].intv.l, cs[i].intv.u, cs[i].defpoly)
	// }

	// signature 設定して
	c := cs[len(cs)-1]
	for j := 0; j < len(c.signature); j++ {
		c.signature[j] = signs[j]
		for i := len(cs) - 2; i > 0; i -= 2 {
			if cs[i].multiplicity[j] > 0 {
				cs[i].signature[j] = 0
				if cs[i].multiplicity[j]%2 == 0 {
					cs[i-1].signature[j] = cs[i+1].signature[j]
				} else {
					cs[i-1].signature[j] = -cs[i+1].signature[j]
				}
			} else {
				cs[i-0].signature[j] = cs[i+1].signature[j]
				cs[i-1].signature[j] = cs[i+1].signature[j]
			}
		}
	}
	cell.children = cs

	undefined := false
	for _, c = range cs {
		switch c.evalTruth(cad.fml, cad).(type) {
		case *AtomT:
			cad.stat.true_cell++
			c.truth = t_true
		case *AtomF:
			cad.stat.false_cell++
			c.truth = t_false
		default:
			undefined = true
		}
		cad.stat.cell++
	}
	if cad.q[cell.lv+1] >= 0 {
		qx := cad.q[cell.lv+1]
		for _, c = range cs {
			if c.truth == qx {
				// exists なら true があった.
				// forall なら false があった
				cell.truth = qx

				//.... さらに親に伝播?
				cell.set_parent_and_truth_other(cad)
				return nil
			}
		}
		if !undefined {
			// 全ての子供の真偽値が決まっていた
			cell.truth = 1 - qx
			cell.set_parent_and_truth_other(cad)
			return nil
		}

		// quantifier なら親の真偽値に影響する
		cad.stack.push(cell)
	}

	// section を追加
	// @TODO ほんとは拡大次数が高い=計算量が大きそうなものからいれたい
	for i := 1; i < len(cs); i += 2 {
		if cs[i].truth < 0 {
			cad.stack.push(cs[i])
		}
	}
	// sector をあとで．
	for i := 0; i < len(cs); i += 2 {
		if cs[i].truth < 0 {
			cad.stack.push(cs[i])
		}
	}

	return nil
}

func (cell *Cell) evalTruth(formula Fof, cad *CAD) Fof {
	switch fml := formula.(type) {
	case *FmlAnd:
		var t Fof = trueObj
		for _, f := range fml.fml {
			t = NewFmlAnd(t, cell.evalTruth(f, cad))
		}
		return t
	case *FmlOr:
		var t Fof = falseObj
		for _, f := range fml.fml {
			t = NewFmlOr(t, cell.evalTruth(f, cad))
		}
		return t
	case *AtomProj:
		sgn, b := fml.pl.evalSign(cell)
		if !b {
			return fml
		}
		if sgn < 0 && (fml.op&LT) != 0 ||
			sgn == 0 && (fml.op&EQ) != 0 ||
			sgn > 0 && (fml.op&GT) != 0 {
			return trueObj
		} else {
			return falseObj
		}
	}
	panic("stop")
}

func (cad *CAD) midSamplePoint(c, d *Cell) NObj {
	if c.intv.inf != nil && d.intv.inf != nil {
		return c.intv.sup.Add(d.intv.inf).Div(two).(NObj)
	}

	var di, ci *Interval
	if c.nintv == nil {
		ci = c.getNumIsoIntv(50)
	} else {
		ci = c.nintv
	}
	if d.nintv == nil {
		di = d.getNumIsoIntv(50)
	} else {
		di = d.nintv
	}

	fmt.Printf("ci=%f, di=%f\n", ci, di)
	f := new(big.Float)
	f.Add(ci.sup, di.inf)
	f.Quo(f, big.NewFloat(2))

	rat := newRat()
	f.Rat(rat.n)

	fmt.Printf("c:%v, %v -> %v\n", c.Index(), c.intv.sup, ci.sup)
	fmt.Printf("d:%v, %v -> %v\n", d.Index(), d.intv.inf, di.inf)
	fmt.Printf("m:%v, %v\n", f, rat)

	return rat
}

func (cad *CAD) addSector(parent *Cell, cs []*Cell) []*Cell {
	// @TODO 整数を優先するとか
	// @TODO 葉である場合には，サンプル点不要
	ret := make([]*Cell, len(cs)*2+1)
	ret[0] = NewCell(cad, parent, 0)
	if len(cs) == 0 {
		ret[0].intv.inf = zero
		ret[0].intv.sup = zero
		return ret
	}

	if cs[0].intv.inf != nil {
		ret[0].intv.inf = cs[0].intv.inf.Sub(one).(NObj)
	} else {
		f_one := big.NewFloat(1)
		f := new(big.Float)
		f.Sub(cs[0].nintv.inf, f_one)
		ii := newInt()
		f.Int(ii.n)
		ret[0].intv.inf = ii
	}
	ret[0].intv.sup = ret[0].intv.inf
	ret[1] = cs[0]
	cs[0].index = 1
	for i := 1; i < len(cs); i++ {
		ret[2*i] = NewCell(cad, parent, uint(2*i))
		m := cad.midSamplePoint(cs[i-1], cs[i])
		ret[2*i].intv.inf = m
		ret[2*i].intv.sup = m
		cs[i].index = uint(2*i + 1)
		ret[2*i+1] = cs[i]
	}
	n := len(ret) - 1
	ret[n] = NewCell(cad, parent, uint(n))

	if cs[len(cs)-1].intv.inf != nil {
		ret[n].intv.inf = cs[len(cs)-1].intv.sup.Add(one).(NObj)
	} else {
		f_one := big.NewFloat(1)
		f := new(big.Float)
		f.Add(cs[len(cs)-1].nintv.sup, f_one)
		ii := newInt()
		f.Int(ii.n)
		ret[n].intv.inf = ii
	}
	ret[n].intv.sup = ret[n].intv.inf
	return ret
}

func (cad *CAD) cellcmp(c, d *Cell) (int, bool) {
	if c.nintv != nil {
		if d.nintv != nil {
			if c.nintv.sup.Cmp(d.nintv.inf) < 0 {
				return -1, true
			} else if d.nintv.sup.Cmp(c.nintv.inf) < 0 {
				return +1, true
			}
		} else {
			fmt.Printf("c=%v: %v, %v %v\n", c.Index(), c.defpoly, c.nintv, c.intv.inf)
			fmt.Printf("d=%v: %v, %v %v\n", d.Index(), d.defpoly, d.nintv, d.intv.inf)
			panic("nooo")
		}
	} else {
		if d.nintv != nil {
			cc := c.getNumIsoIntv(d.nintv.Prec())
			if cc.sup.Cmp(d.nintv.inf) < 0 {
				return -1, true
			} else if d.nintv.sup.Cmp(cc.inf) < 0 {
				return +1, true
			}
		} else {
			if c.intv.sup.Cmp(d.intv.inf) < 0 {
				return -1, true
			} else if d.intv.sup.Cmp(c.intv.inf) < 0 {
				return +1, true
			}
		}
	}
	return 0, false
}

func (cell *Cell) fusion(c *Cell) {
	for k := 0; k < len(cell.multiplicity); k++ {
		cell.multiplicity[k] += c.multiplicity[k]
	}
	// cs[j] はもういらない
	c.multiplicity = nil
}

func (cad *CAD) cellsort(cs []*Cell, dup bool) []*Cell {
	// dup: 同じ根を表現する可能性がある場合は true
	// @TODO とりま bubble sort.
	for i := len(cs) - 1; i >= 0; i-- {
		for j := 0; j < i; j++ {
			for {
				if s, ok := cad.cellcmp(cs[j], cs[i]); s < 0 {
					break
				} else if s > 0 {
					// cs[i] < cs[j]
					c := cs[i]
					cs[i] = cs[j]
					cs[j] = c
					break
				} else if ok {
					// 一致した
				}

				// 共通根をもつ可能性があるか?
				if dup && cad.proj[cs[i].lv].hasCommonRoot(cs[i].parent, cs[j].index, cs[i].index) {
					if cs[i].defpoly == nil || cs[j].defpoly == nil {
						if cs[i].defpoly == nil && cs[j].defpoly == nil && cs[i].intv.inf.Equals(cs[j].intv.inf) {
							cs[i].fusion(cs[j])
							c := cs[j]
							for k := j + 1; k < len(cs); k++ {
								cs[k-1] = cs[k]
							}
							cs[len(cs)-1] = c
							goto _RETRY

						}
					} else if cs[i].defpoly.Equals(cs[j].defpoly) {
						// 一致した.
						cs[i].fusion(cs[j])
						c := cs[j]
						for k := j + 1; k < len(cs); k++ {
							cs[k-1] = cs[k]
						}
						cs[len(cs)-1] = c
						goto _RETRY
					}
					cs[i].parent.Print(os.Stdout)
					cs[i].parent.Print(os.Stdout, "signature")
					cs[i].Print(os.Stdout)
					cs[j].Print(os.Stdout)
					panic("not implemented")
					// multiplicity のマージ
				}

				// 区間を改善
				cs[i].improveIsoIntv()
				cs[j].improveIsoIntv()
			}
		}
	_RETRY:
	}

	for i := len(cs) - 1; i >= 0; i-- {
		if cs[i].multiplicity != nil {
			return cs[:i+1]
		}
	}
	return cs

	// @Print
	// for i := 0; i < len(cs); i++ {
	// 	fmt.Printf("cs[%d]=<%e,%e> <%v,%v>\n", i, cs[i].intv.l.Float(), cs[i].intv.u.Float(),
	// 		cs[i].intv.l, cs[i].intv.u)
	// }

}

func (cell *Cell) root_iso_q(cad *CAD, pf *ProjFactor, p *Poly) []*Cell {
	// returns (roots, sign(lc(p)))
	cs := make([]*Cell, 0, len(p.c)-1)
	fctrs := cad.g.ox.Factor(p)
	cad.stat.fctr++
	fmt.Printf("root_iso(%v,%d): %v -> %v\n", cell.Index(), pf.index, p, fctrs)
	for i := fctrs.Len() - 1; i > 0; i-- {
		ff := fctrs.getiList(i)
		q := ff.getiPoly(0)
		r := int8(ff.getiInt(1).Int64())
		if len(q.c) == 2 {
			c := NewCell(cad, cell, pf.index)
			rat := NewRatFrac(q.c[0].(*Int), q.c[1].(*Int).Neg().(*Int))
			if rat.n.IsInt() {
				ci := new(Int)
				ci.n = rat.n.Num()
				c.intv.inf = ci
				c.intv.sup = ci
			} else {
				c.intv.inf = rat
				c.intv.sup = rat
			}
			c.multiplicity[pf.index] = r
			cs = append(cs, c)
		} else {
			roots := q.realRootIsolation(-30) // @TODO DEBUG用に大きい値を設定中
			cad.stat.qrealroot++
			sgn := sign_t(1)
			if len(q.c)%2 == 0 {
				sgn = -1
			}
			if q.Sign() < 0 {
				sgn *= -1
			}
			for i := 0; i < len(roots); i++ {
				rr := roots[i]
				c := NewCell(cad, cell, pf.index)
				c.intv.inf = rr.low
				if rr.point { // fctr しているからこのルートは通らないはず.
					c.intv.sup = rr.low
				} else {
					c.intv.sup = rr.low.upperBound()
				}
				c.sgn_of_left = sgn
				sgn *= -1
				c.multiplicity[pf.index] = r
				c.defpoly = q
				c.ex_deg = int(r)
				cs = append(cs, c)
			}
		}
	}
	if fctrs.Len() == 2 {
		return cs
	}

	// sort...
	cs = cad.cellsort(cs, false)

	return cs
}

func (cell *Cell) root_iso_i(cad *CAD, pf *ProjFactor, p *Poly) []*Cell {
	cs := make([]*Cell, 0, len(p.c)-1)
	return cs
}

func (cell *Cell) reduce(p *Poly) RObj {
	// 次数を下げる.
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly != nil {
			if _, ok := c.defpoly.c[len(c.defpoly.c)-1].(NObj); !ok {
				// 正規化されていない
				continue
			}
			q := p.reduce(c.defpoly)
			if qq, ok := q.(*Poly); ok {
				p = qq
			} else {
				return q
			}
		}
	}
	return p
}

func (cell *Cell) make_cells_try1(cad *CAD, pf *ProjFactor, pr RObj) (*Poly, []*Cell, sign_t) {
	// returns (p, c, s)
	// 子供セルが作れたら， p=nil, s=pfに cell 代入したときの主係数の符号
	// 子供セルが作れなかったら p != nil, (c,s) は使わない
	if pf.discrim != nil {
		dd, db := pf.discrim.evalSign(cell)
		fmt.Printf("make_cells_try1(%v, %d) pr=%v, discrim=%d:%v\n", cell.Index(), pf.index, pr, dd, db)
	} else {
		fmt.Printf("make_cells_try1(%v, %d) pr=%v, p=%v\n", cell.Index(), pf.index, pr, pf.p)
	}
	switch p := pr.(type) {
	case *Poly:
		if p.isUnivariate() && p.lv == pf.p.lv {
			// 他の変数が全部消えた.
			return nil, cell.root_iso_q(cad, pf, p), sign_t(p.Sign())
		} else if p.lv != pf.p.lv {
			// 主変数が消えて定数になった.
			if pf.coeff[0] != nil {
				s, _ := pf.coeff[0].evalSign(cell)
				return nil, []*Cell{}, s
			}

			cell.reduce(p)

			// 定数の符号を決定する.
			panic("not implemented")
		}
		return p, nil, 0
	case NObj:
		return nil, []*Cell{}, sign_t(p.Sign())
	}
	panic("invalid")
}

func (cell *Cell) make_cells(cad *CAD, pf *ProjFactor) ([]*Cell, sign_t) {

	p := pf.p

	if cell.de || cell.defpoly != nil {
		// projection factor の情報から，ゼロ値を決める
		pp := NewPoly(p.lv, len(p.c))
		up := false
		for i := 0; i < len(p.c); i++ {
			if pf.coeff[i] == nil {
				pp.c[i] = p.c[i]
			} else if s, _ := pf.coeff[i].evalSign(cell); s == 0 {
				pp.c[i] = zero
				up = true
			} else {
				pp.c[i] = p.c[i]
			}
		}
		if up {
			switch pq := pp.normalize().(type) {
			case NObj:
				return []*Cell{}, sign_t(pq.Sign())
			case *Poly:
				p = pq
			}
		}
	}

	for c := cell; c != cad.root; c = c.parent {
		// 有理数代入
		if c.defpoly == nil {
			pp := c.intv.inf.subst_poly(p, Level(c.lv))
			switch px := pp.(type) {
			case *Poly:
				p = px
			case NObj:
				return []*Cell{}, sign_t(px.Sign())
			}
		}
	}

	p, c, s := cell.make_cells_try1(cad, pf, p)
	if p == nil {
		return c, s
	}

	// とりあえず簡単化してみる
	p, c, s = cell.make_cells_try1(cad, pf, cell.reduce(p))
	if p == nil {
		return c, s
	}

	for prec := uint(53); ; prec += uint(53) {
		c, s, err := cell.make_cells_i(cad, pf, p, prec, 1)
		if err == nil {
			fmt.Printf("c=%v, s=%v\n", c, s)
			return c, s
		}

		// 係数区間にゼロを含むなら.... 記号演算でチェック

		// セルの分離区間を改善
		cell.Print(os.Stdout)
		panic("not implemented")

		// 精度が足りなかったか無平方でなかったか.
	}
}

func (cell *Cell) getNumIsoIntv(prec uint) *Interval {
	// isolating interval を *Interval に変換する
	if cell.nintv != nil && cell.nintv.Prec() >= prec {
		return cell.nintv.clonePrec(prec)
	}
	if cell.defpoly == nil {
		return cell.intv.inf.toIntv(prec).(*Interval)
	}
	if cell.intv.inf != nil {
		// binary interval
		z := newInterval(prec)
		cell.intv.inf.(*BinInt).setToBigFloat(z.inf)
		cell.intv.sup.(*BinInt).setToBigFloat(z.sup)
		cell.nintv = z
		return z
	}

	panic("unimplemented")
}

func (cell *Cell) make_cells_i(cad *CAD, pf *ProjFactor, p *Poly, prec uint, multiplicity int8) ([]*Cell, sign_t, error) {
	pp := p.toIntv(prec).(*Poly)
	for c := cell; c != cad.root; c = c.parent {
		if !p.hasVar(Level(c.lv)) {
			continue
		}
		x := cell.getNumIsoIntv(prec)
		pp = pp.subst1(x, Level(c.lv)).(*Poly)
	}

	if !pp.isUnivariate() {
		panic("invalid")
	}

	for i, c := range pp.c {
		if c.(*Interval).ContainsZero() && !c.(*Interval).IsZero() {
			fmt.Printf("coef[%d] contains zero: %v\n", i, c)
			return nil, 0, fmt.Errorf("coef contains zero.")
		}
	}

	// 定数項のゼロは取り除く.
	zeros := 0
	for i, c := range pp.c {
		if !c.IsZero() {
			if i > 0 {
				qq := NewPoly(pp.lv, len(pp.c)-i)
				copy(qq.c, pp.c[i:])
				pp = qq
				zeros = i
			}
			break
		}
	}

	ans, err := pp.iRealRoot(prec, 1000)
	cad.stat.irealroot++
	if err != nil {
		fmt.Printf("irealroot failed: %s\n", err.Error())
		return nil, 0, err
	}
	cad.stat.irealroot_ok++

	sgn := pp.Sign()
	cells := make([]*Cell, len(ans), len(ans)+1)
	for i := len(ans) - 1; i >= 0; i-- {
		sgn *= -1
		c := NewCell(cad, cell, pf.index)
		c.nintv = ans[i]
		c.sgn_of_left = sign_t(sgn)
		c.de = true
		c.multiplicity[pf.index] = multiplicity
		c.defpoly = p
		cells[i] = c
	}
	if zeros > 0 {
		// x=0 を追加
		c := NewCell(cad, cell, pf.index)
		c.de = cell.de
		c.multiplicity[pf.index] = int8(zeros)
		c.intv.inf = zero
		c.intv.sup = zero
		for i := 0; i < len(cells); i++ {
			if c.nintv.inf.Sign() >= 0 {
				cs := cells[:i]
				cs = append(cs, c)
				cs = append(cs, cells[i:]...)
				cells = cs
				zeros = 0
				break
			}
		}
		if zeros > 0 {
			cells = append(cells, c)
		}
	}

	return cells, sign_t(pp.Sign()), nil
}

func (pl *ProjLink) evalSign(cell *Cell) (sign_t, bool) {
	sgn := pl.sgn
	fix := true
	for i := 0; i < len(pl.multiplicity); i++ {
		pf := pl.projs.pf[i]
		if Level(cell.lv) < pf.p.lv {
			fix = false
			continue
		}

		c := cell
		for Level(c.lv) != pf.p.lv {
			c = c.parent
		}
		s := c.signature[pf.index]
		if s == 0 {
			return 0, true
		} else if s < 0 && pl.multiplicity[i]%2 == 1 {
			sgn *= -1
		}
	}
	return sgn, fix
}

func (cell *Cell) improveIsoIntv() {
	// 分離区間の改善
	if cell.defpoly == nil {
		return
	}

	switch l := cell.intv.inf.(type) {
	case *BinInt:
		// binint ということは, realroot の出力であり，１変数多項式
		m := l.midBinIntv()
		v := m.subst_poly(cell.defpoly, cell.defpoly.lv)
		if v.Sign() < 0 && cell.sgn_of_left < 0 || v.Sign() > 0 && cell.sgn_of_left > 0 {
			cell.intv.inf = m
			cell.intv.sup = m.upperBound()
		} else if v.Sign() != 0 {
			cell.intv.inf = l.halveIntv()
			cell.intv.sup = m
		} else {
			cell.defpoly = nil
			cell.intv.inf = l
			cell.intv.sup = m
		}
		return
	}

	panic("unimplemented")
}
