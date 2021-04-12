package ganrac

import (
	"fmt"
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
		if c.children != nil {
			c.lift(cad)
		}
		return nil
	}

	for _, idx := range index {
		if c.children == nil || idx < 0 || idx >= len(c.children) {
			return fmt.Errorf("invalid index %v", index)
		}
		c = c.children[idx]
	}
	if c.children != nil {
		c.lift(cad)
	}
	return nil
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
	ciso := make([][]*Cell, len(cad.proj[cell.lv+1].pf))
	signs := make([]sign_t, len(ciso))
	cs := make([]*Cell, 0)
	for i := 0; i < len(ciso); i++ {
		ciso[i], signs[i] = cell.make_cells(cad, cad.proj[cell.lv+1].pf[i])
		cs = append(cs, ciso[i]...)
	}

	// merge して
	cad.cellsort(cs, true)

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

	for _, c = range cs {
		switch c.evalTruth(cad.fml, cad).(type) {
		case *AtomT:
			c.truth = t_true
		case *AtomF:
			c.truth = t_false
		}
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

func (cad *CAD) addSector(parent *Cell, cs []*Cell) []*Cell {
	// @TODO 整数を優先するとか
	ret := make([]*Cell, len(cs)*2+1)
	ret[0] = NewCell(cad, parent, 0)
	if len(cs) == 0 {
		ret[0].intv.l = zero
		ret[0].intv.u = zero
		return ret
	}

	ret[0].intv.l = cs[0].intv.l.Sub(one).(NObj)
	ret[0].intv.u = ret[0].intv.l
	ret[1] = cs[0]
	cs[0].index = 1
	for i := 1; i < len(cs); i++ {
		ret[2*i] = NewCell(cad, parent, uint(2*i))
		m := cs[i-1].intv.u.Add(cs[i].intv.l).Div(two).(NObj)
		ret[2*i].intv.l = m
		ret[2*i].intv.u = m
		cs[i].index = uint(2*i + 1)
		ret[2*i+1] = cs[i]
	}
	n := len(ret) - 1
	ret[n] = NewCell(cad, parent, uint(n))
	ret[n].intv.l = cs[len(cs)-1].intv.u.Add(one).(NObj)
	ret[n].intv.u = ret[n].intv.l
	return ret
}

func (cad *CAD) cellsort(cs []*Cell, dup bool) {
	// dup: 同じ根を表現する可能性がある場合は true
	// @TODO とりま bubble sort.
	for i := 0; i < len(cs); i++ {
		for j := 0; j < i; j++ {
			for {
				if cs[j].intv.u.Cmp(cs[i].intv.l) < 0 {
					break
				} else if cs[i].intv.u.Cmp(cs[j].intv.l) < 0 {
					// cs[i] < cs[j]
					c := cs[i]
					cs[i] = cs[j]
					cs[j] = c
					break
				}

				// 共通根をもつ可能性があるか?
				if dup && cad.proj[cs[i].lv].hasCommonRoot(cs[i].parent, cs[j].index, cs[i].index) {
					cs[i].parent.Print(os.Stdout)
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
	}

	// @Print
	// for i := 0; i < len(cs); i++ {
	// 	fmt.Printf("cs[%d]=<%e,%e> <%v,%v>\n", i, cs[i].intv.l.Float(), cs[i].intv.u.Float(),
	// 		cs[i].intv.l, cs[i].intv.u)
	// }

}

func (cell *Cell) root_iso_q(cad *CAD, pf *ProjFactor, p *Poly) []*Cell {
	// returns (roots, sign(lc(p)))
	cs := make([]*Cell, 0, len(p.c)-1)
	fmt.Printf("root_iso: %v\n", p)
	fctrs := cad.g.ox.Factor(p)
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
				c.intv.l = ci
				c.intv.u = ci
			} else {
				c.intv.l = rat
				c.intv.u = rat
			}
			c.multiplicity[pf.index] = r
			cs = append(cs, c)
		} else {
			roots, _ := q.RealRootIsolation(+30) // DEBUG用に大きい値を設定中
			sgn := sign_t(1)
			if len(q.c)%2 == 0 {
				sgn = -1
			}
			if q.Sign() < 0 {
				sgn *= -1
			}
			for i := 0; i < roots.Len(); i++ {
				rr := roots.getiList(i)
				c := NewCell(cad, cell, pf.index)
				c.intv.l = rr.geti(0).(NObj)
				c.intv.u = rr.geti(1).(NObj)
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
	cad.cellsort(cs, false)

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
	fmt.Printf("make_cells_try1() pr=%v, pf=%v\n", pr, pf.p)
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
			pp := c.intv.l.subst_poly(p, Level(c.lv))
			fmt.Printf("pq=%v\n", pp)
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
		c, s, err := cell.make_cells_i(cad, pf, p, prec)
		if err != nil && false {
			return c, s
		}

		// 区間にゼロを含むなら.... 記号演算でチェック

		// セルの分離区間を改善
		cell.Print(os.Stdout)
		panic("not implemented")

		// 精度が足りなかったか無平方でなかったか.
	}
}

func (cell *Cell) getNumIsoIntv(prec uint) *Interval {
	if cell.nintv != nil && cell.nintv.Prec() >= prec {
		return cell.nintv.clonePrec(prec)
	}
	if cell.defpoly == nil {
		return cell.intv.l.toIntv(prec).(*Interval)
	}
	if cell.intv.l != nil {
		// binary interval
		z := newInterval(prec)
		cell.intv.l.(*BinInt).setToBigFloat(z.lv)
		cell.intv.u.(*BinInt).setToBigFloat(z.uv)
		cell.nintv = z
		return z
	}

	panic("unimplemented")
}

func (cell *Cell) make_cells_i(cad *CAD, pf *ProjFactor, p *Poly, prec uint) ([]*Cell, sign_t, error) {
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

	for _, c := range pp.c {
		if c.(*Interval).ContainsZero() && !c.(*Interval).IsZero() {
			return nil, 0, fmt.Errorf("coef contains zero.")
		}
	}

	_, err := pp.iRealRoot(prec)
	if err != nil {
		return nil, 0, err
	}

	return []*Cell{}, sign_t(pp.Sign()), nil
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

	switch l := cell.intv.l.(type) {
	case *BinInt:
		// binint ということは, realroot の出力であり，１変数多項式
		m := l.midBinIntv()
		v := m.subst_poly(cell.defpoly, cell.defpoly.lv)
		if v.Sign() < 0 && cell.sgn_of_left < 0 || v.Sign() > 0 && cell.sgn_of_left > 0 {
			cell.intv.l = m
			cell.intv.u = m.upperBound()
		} else if v.Sign() != 0 {
			cell.intv.l = l.halveIntv()
			cell.intv.u = m
		} else {
			cell.defpoly = nil
			cell.intv.u = l
			cell.intv.u = m
		}
		return
	}

	panic("unimplemented")
}
