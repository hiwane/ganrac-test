package ganrac

import (
	"fmt"
)

func (cad *CAD) Lift(index ...int) error {
	if len(index) == 0 {
		for !cad.stack.empty() {
			cell := cad.stack.pop()
			cell.lift(cad)
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
			c.truth = 1
		case *AtomF:
			c.truth = 0
		}
	}
	if cad.q[cell.lv+1] >= 0 {
		qx := cad.q[cell.lv+1]
		for _, c = range cs {
			if c.truth == qx {
				cell.truth = qx

				for {
					cell = cell.parent
					if cad.q[cell.lv+1] != qx {
						break
					}
					cell.truth = qx
				}

				return nil
			}
		}
	}

	// section から
	// @TODO ほんとは拡大次数が高いものからいれたい
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
	fctrs := cad.g.ox.Factor(p)
	for i := fctrs.Len() - 1; i > 0; i-- {
		ff := fctrs.getiList(i)
		q := ff.getiPoly(0)
		r := int8(ff.getiInt(1).Int64())
		if len(q.c) == 2 {
			c := NewCell(cad, cell, pf.index)
			rat := NewRatFrac(q.c[0].(*Int), q.c[1].(*Int).Neg().(*Int))
			c.intv.l = rat
			c.intv.u = rat
			c.multiplicity[pf.index] = r
			cs = append(cs, c)
		} else {
			roots, _ := q.RealRootIsolation(+30)
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

func (cell *Cell) make_cells(cad *CAD, pf *ProjFactor) ([]*Cell, sign_t) {
	pp, isQ, sgn := cell.subst_sample_points(cad, pf)
	switch p := pp.(type) {
	case *Poly:
		if isQ {
			// fctr.
			return cell.root_iso_q(cad, pf, p), sgn
		}
	case NObj:
		return make([]*Cell, 0, 0), sgn
	}
	panic("unimplemented")
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
		m := l.midBinIntv()
		v := cell.defpoly.subst_noden(m, cell.defpoly.lv)
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

func (cell *Cell) subst_sample_points(cad *CAD, pf *ProjFactor) (RObj, bool, sign_t) {
	p := pf.p
	c := cell
	is_q := true
	for c != cad.root {
		if c.ex_deg == 1 || c.defpoly == nil {
			pp := p.subst1(c.intv.l, Level(c.lv))
			switch px := pp.(type) {
			case *Poly:
				p = px
			case NObj:
				return px, is_q, sign_t(px.Sign())
			}
		} else {
			panic("unimplemented")
		}
		c = c.parent
	}

	return p, is_q, sign_t(p.Sign())
}
