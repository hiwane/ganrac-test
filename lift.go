package ganrac

// H. Iwane, H. Yanami, H. Anai, K. Yokoyama
// An effective implementation of symbolic–numeric cylindrical algebraic decomposition for quantifier elimination
// symoblic numeric computation 2009.

import (
	"fmt"
	"math/big"
	"time"
)

func (cell *Cell) Precs() []uint {
	prec := make([]uint, cell.lv+1)
	c := cell
	for c.lv >= 0 {
		prec[c.lv] = c.Prec()
		c = c.parent
	}
	return prec
}

func (cell *Cell) Prec() uint {
	if cell.defpoly == nil {
		return 0
	}
	switch l := cell.intv.inf.(type) {
	case *BinInt:
		if l.m > 0 {
			return 0
		} else {
			return uint(-l.m)
		}
	}
	return cell.nintv.Prec()
}

func (cell *Cell) Index() []uint {
	// インデックスを返す. 論文は 1 始まりだが，0 始まりであることに注意
	// つまり， section は奇数である
	idx := make([]uint, cell.lv+1)
	c := cell
	for c.lv >= 0 {
		idx[c.lv] = c.index
		c = c.parent
	}
	return idx
}

func (cell *Cell) hasSection() bool {
	// 自分〜先祖に section がいるか.
	for c := cell; c.lv >= 0; c = c.parent {
		if c.isSection() {
			return true
		}
	}
	return false
}

func (cell *Cell) set_truth_value_from_children(cad *CAD) {
	// 子供の真偽値から自分の真偽値を決める.
	if cad.q[cell.lv+1] < 0 {
		cell.Print("cellp")
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
	for ; c.lv >= 0 && cad.q[c.lv] == c.truth; c = c.parent {
		c.parent.truth = cell.truth
		c.parent.set_truth_other()
	}
	if c.lv >= 0 && cad.q[c.lv] >= 0 {
		// 限量子の交代があった.
		for _, d := range c.parent.children {
			if d.truth != c.truth {
				return
			}
		}
		c.parent.truth = c.truth
		c.parent.set_parent_and_truth_other(cad)
	}
}

func (cad *CAD) Lift(index ...int) error {
	if cad.stage != 1 {
		return fmt.Errorf("invalid stage")
	}
	cad.log(2, "cad.Lift %v\n", index)
	if len(index) == 0 { // 指定なしなので，最後までやる.
		tm_start := time.Now()
		for !cad.stack.empty() {
			cell := cad.stack.pop()
			if cell.truth >= 0 {
				continue
			} else if cell.children != nil {
				// 子供の真偽値が確定した.
				cell.set_truth_value_from_children(cad)
				continue
			} else {
				if err := cell.lift(cad); err != nil {
					return err
				}
			}
		}
		err := cad.root.valid(cad)
		cad.stage = CAD_STAGE_LIFTED
		cad.stat.tm[1] = time.Since(tm_start)
		return err
	}
	c := cad.root
	if len(index) == 1 && index[0] == -1 { // 指定なしと区別するため，root は -1 で表現
		if c.children == nil {
			return c.lift(cad)
		} else {
			return fmt.Errorf("already lifted %v", index)
		}
	}

	for _, idx := range index {
		if c.children == nil {
			err := c.lift(cad)
			if err != nil {
				return err
			}
		}
		if idx < 0 || idx >= len(c.children) {
			return fmt.Errorf("invalid index %v", index)
		}
		c = c.children[idx]
	}
	if c.children == nil {
		return c.lift(cad)
	} else {
		return fmt.Errorf("already lifted %v", index)
	}
}

func (cell *Cell) setParentChildren() {
	for _, c := range cell.children {
		c.parent = cell
	}
}

func (cell *Cell) cloneSetParentChildren(parent *Cell) []*Cell {
	// rebuild CAD 用.
	// truth value が決定していないセルの子供のコピーを生成する.

	cs := make([]*Cell, len(cell.children))
	for i, c := range cell.children {
		cs[i] = new(Cell)
		*cs[i] = *c
		cs[i].parent = parent
		if c.truth < 0 {
			cs[i].children = c.cloneSetParentChildren(cs[i])
		}
	}
	return cs
}

func (cell *Cell) same_sig(old *Cell, num int) bool {
	if cell == old {
		return true
	}
	for i := 0; i < num; i++ {
		if cell.signature[i] != old.signature[i] {
			return false
		}
	}
	return true
}

func (cell *Cell) rlift(cad *CAD, lv Level, proj_num []int) error {
	// proj_num すでに構築済みの射影因子数
	if cell.lv+1 < lv {
		for _, c := range cell.children {
			c.rlift(cad, lv, proj_num)
		}
		return nil
	}
	if cell.truth >= 0 {
		return nil
	}
	// cell.lv + 1 == lv
	pfs := cad.proj[cell.lv+1]
	num := pfs.Len()
	if num == len(cell.children[0].signature) {
		// もう持ち上げ済み
		return nil
	}
	if num == proj_num[lv] {
		for _, c := range cell.children {
			c.rlift(cad, lv+1, proj_num)
		}
		return nil
	}
	cad.log(3, "rlift(%v, #child=%d) ... #proj=%d/%d/%d\n", cell.Index(), len(cell.children), proj_num[lv], num, len(cell.children[0].signature))
	cad.stat.rlift[cell.lv+1]++

	oldcs := cell.children
	ciso := make([][]*Cell, 0, num-proj_num[lv]+1)
	signs := make([]sign_t, proj_num[lv], num)
	copy(signs, cell.children[len(cell.children)-1].signature)

	for i := proj_num[lv]; i < num; i++ {
		pf := pfs.get(uint(i))
		c, s := cell.make_cells(cad, pf)
		ciso = append(ciso, c)
		signs = append(signs, s)
	}

	// すでに構築済みのものの signature/multiplicity 領域の拡張
	cs := make([]*Cell, 0, len(cell.children)/2)
	for i := 1; i < len(cell.children); i += 2 {
		c := cell.children[i]
		sig := c.signature
		c.signature = make([]sign_t, num)
		copy(c.signature, sig)
		mul := c.multiplicity
		c.multiplicity = make([]mult_t, num)
		copy(c.multiplicity, mul)
		cs = append(cs, c)

		for j := 0; j < len(mul); j++ {
			if mul[j] > 0 {
				c.index = uint(j)
			}
		}

		// 分離区間の更新が必要である.... @BUG
	}
	ciso = append(ciso, cs)

	// merge して
	cs = cad.cellmerge(ciso, cell.hasSection())

	// sector 作って
	cs = cad.addSector(cell, cs)

	// signature 設定して
	cell.set_signatures(cs, signs)

	cs = cell.children
	undefined := false

	cs[0].truth = oldcs[0].truth
	cs[0].children = oldcs[0].children
	cs[0].setParentChildren()
	cs[0].rlift(cad, lv+1, proj_num)

	if err := cs[0].valid(cad); err != nil {
		cs[0].Print("signatures")
		panic(err.Error())
	}

	m := 1
	for i := 1; i < len(cs); i += 2 {
		if m < len(oldcs) && cs[i].same_sig(oldcs[m], proj_num[lv]) {
			for mm := 0; mm < 2; mm++ {
				cs[i+mm].truth = oldcs[m+mm].truth
				cs[i+mm].children = oldcs[m+mm].children
				cs[i+mm].setParentChildren()
			}
			m += 2
		} else {
			for mm := 0; mm < 2; mm++ {
				cs[i+mm].truth = oldcs[m-1].truth
				cs[i+mm].children = oldcs[m-1].cloneSetParentChildren(cs[i+mm])
			}
		}
		for mm := 0; mm < 2; mm++ {
			cs[i+mm].rlift(cad, lv+1, proj_num)

			if err := cs[i+mm].valid(cad); err != nil {
				cs[i+mm].Print("signatures")
				panic(err.Error())
			}
		}
	}
	if m+2 < len(oldcs) {
		panic(fmt.Sprintf("m=%d, old=%d\n", m, len(oldcs)))
	}

	cell.lift_term(cad, undefined)

	if err := cell.valid(cad); err != nil {
		cell.Print("signatures")
		panic(err.Error())
	}
	return nil
}

func (cell *Cell) lift(cad *CAD) error {
	cad.log(2, "lift (%v)\n", cell.Index())
	cad.stat.lift[cell.lv+1]++
	ciso := make([][]*Cell, cad.proj[cell.lv+1].Len())
	signs := make([]sign_t, len(ciso))
	for i, pf := range cad.proj[cell.lv+1].gets() {
		ciso[i], signs[i] = cell.make_cells(cad, pf)

		if signs[i] == 0 {
			// vanish!
			if !pf.vanishChk(cad, cell) {
				return CAD_NO_WO
			}
		}

		if true {
			if pf.Index() != uint(i) { // @DEBUG
				panic("3?")
			}
			for _, c := range ciso[i] {
				if c.index != uint(i) { // @DEBUG
					panic("4?")
				}
				if c.parent != cell { // @DEBUG
					panic("5?")
				}
			}
		}
	}

	for i, cs := range ciso { // @DEBUG
		// @DEBUG. ciso[i] に属する cell の index には i が設定されていること.
		for j, c := range cs {
			if c.index != uint(i) {
				fmt.Printf("i=%d, j=%d, index=%d\n", i, j, c.index)
				c.Print()
				panic("ge")
			}
			if j != 0 {
				if s, b := cad.cellcmp(cs[j-1], c); !b || s >= 0 {
					panic("go")
				}
			}
		}
	}

	// merge して
	cs := cad.cellmerge(ciso, cell.hasSection())

	// sector 作って
	cs = cad.addSector(cell, cs)

	// for i := 0; i < len(cs); i++ {
	// 	fmt.Printf("cs[%d]=<%e,%e> <%v,%v> %v\n", i, cs[i].intv.l.Float(), cs[i].intv.u.Float(),
	// 		cs[i].intv.l, cs[i].intv.u, cs[i].defpoly)
	// }

	// signature 設定して
	cell.set_signatures(cs, signs)

	cs = cell.children
	undefined := false
	for _, c := range cs {
		switch c.evalTruth(cad.fml, cad).(type) {
		case *AtomT:
			cad.stat.true_cell[c.lv]++
			c.truth = t_true
		case *AtomF:
			cad.stat.false_cell[c.lv]++
			c.truth = t_false
		default:
			undefined = true
		}
		cad.stat.cell[c.lv]++
	}

	// 真偽値確認して
	cell.lift_term(cad, undefined)

	return nil
}

func (cell *Cell) lift_term(cad *CAD, undefined bool) {
	cs := cell.children
	if cad.q[cell.lv+1] >= 0 {
		qx := cad.q[cell.lv+1]
		for _, c := range cs {
			if c.truth == qx {
				// exists なら true があった.
				// forall なら false があった
				cell.truth = qx

				//.... さらに親に伝播?
				cell.set_parent_and_truth_other(cad)
				return
			}
		}
		if !undefined {
			// 全ての子供の真偽値が決まっていた
			cell.truth = 1 - qx
			cell.set_parent_and_truth_other(cad)
			return
		}

		// quantifier なら親の真偽値に影響する
		cad.stack.push(cell)
	}
	if !undefined {
		// 子供のセルの真偽値がすべて確定
		return
	}

	// section を追加
	// @TODO ほんとは拡大次数が高い=計算量が大きそうなものからいれたい
	// rebuild CAD のときに子供が既にいるかもしれないのでチェックが必要
	for i := 1; i < len(cs); i += 2 {
		if cs[i].truth < 0 && cs[i].children == nil {
			cad.stack.push(cs[i])
		}
	}
	// sector をあとで．
	for i := 0; i < len(cs); i += 2 {
		if cs[i].truth < 0 && cs[i].children == nil {
			cad.setSamplePoint(cs, i)
			cad.stack.push(cs[i])
		}
	}

	return
}

func (cell *Cell) set_signatures(cs []*Cell, signs []sign_t) {
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
}

func (cell *Cell) evalTruth(formula Fof, cad *CAD) Fof {
	// cell での formula の真偽値を評価してみる.
	// 確定しない場合は atom をそのまま返す
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
		s := fml.pl.evalSign(cell)
		if s == OP_TRUE {
			return fml
		} else if s&fml.op == 0 {
			return falseObj
		} else if (s & fml.op.not()) == 0 {
			return trueObj
		} else {
			return fml
		}
	}
	panic("stop")
}

func (cad *CAD) midSamplePoint(c, d *Cell) NObj {
	// @TODO 代入しやすいサンプル点を生成する.
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

	// ciz := new(big.Int)
	// diz := new(big.Int)
	// ci.sup.Int(ciz)
	// di.inf.Int(diz)
	// fmt.Printf("ci=%f, di=%f\n", ci, di)
	// fmt.Printf("ciz=%v, diz=%v\n", ciz, diz)
	// if ci.Cmp(di) <= 0 {
	// 	return NewIntZ(ciz)
	// }

	f := new(big.Float)
	f.Add(ci.sup, di.inf)
	f.Quo(f, big.NewFloat(2))

	rat := newRat()
	f.Rat(rat.n)

	return rat
}

func (cad *CAD) setSamplePoint(cells []*Cell, idx int) {
	// set a sample point to cs[idx] where cs[idx] is a sector
	// @TODO 整数を優先するとか
	if idx%2 != 0 {
		panic("invalid index")
	}
	c := cells[idx]
	if c.intv.inf != nil {
		return
	}
	if idx == 0 {
		if cells[1].intv.inf != nil {
			c.intv.inf = cells[1].intv.inf.Sub(one).(NObj)
		} else {
			f_one := big.NewFloat(1)
			f := new(big.Float)
			f.Sub(cells[1].nintv.inf, f_one)
			ii := newInt()
			f.Int(ii.n)
			c.intv.inf = ii
		}
	} else if idx < len(cells)-1 {

		m := cad.midSamplePoint(cells[idx-1], cells[idx+1])
		c.intv.inf = m
	} else {
		if cells[idx-1].intv.inf != nil {
			c.intv.inf = cells[idx-1].intv.sup.Add(one).(NObj)
		} else {
			f_one := big.NewFloat(1)
			f := new(big.Float)
			f.Add(cells[idx-1].nintv.sup, f_one)
			ii := newInt()
			f.Int(ii.n)
			c.intv.inf = ii
		}
	}
	c.intv.sup = c.intv.inf
}

func (cad *CAD) addSector(parent *Cell, cs []*Cell) []*Cell {
	ret := make([]*Cell, len(cs)*2+1)
	ret[0] = NewCell(cad, parent, 0)
	if len(cs) == 0 {
		ret[0].intv.inf = zero
		ret[0].intv.sup = zero
		return ret
	}

	ret[1] = cs[0]
	cs[0].index = 1
	for i := 1; i < len(cs); i++ {
		ret[2*i] = NewCell(cad, parent, uint(2*i))
		cs[i].index = uint(2*i + 1)
		ret[2*i+1] = cs[i]
	}
	n := len(ret) - 1
	ret[n] = NewCell(cad, parent, uint(n))

	return ret
}

func (cad *CAD) cellcmp(c, d *Cell) (int, bool) {
	// returns (s, b)
	//  一致するか分からない場合は, b=false
	if c.nintv != nil {
		if d.nintv != nil {
			if c.nintv.sup.Cmp(d.nintv.inf) < 0 {
				return -1, true
			} else if d.nintv.sup.Cmp(c.nintv.inf) < 0 {
				return +1, true
			}
		} else {
			dd := d.getNumIsoIntv(c.nintv.Prec())
			if c.nintv.sup.Cmp(dd.inf) < 0 {
				return -1, true
			} else if dd.sup.Cmp(c.nintv.inf) < 0 {
				return +1, true
			}
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
	// cell と c は同じものだったので，融合する
	for k := 0; k < len(cell.multiplicity); k++ {
		cell.multiplicity[k] += c.multiplicity[k]
	}
	// cs[j] はもういらない
	c.multiplicity = cell.multiplicity
}

func (cad *CAD) cellmerge(ciso [][]*Cell, dup bool) []*Cell {
	// dup: 同じ根を表現する可能性がある場合は true

	// for i, cs := range ciso { // @DEBUG
	// 	// @DEBUG. ciso[i] に属する cell の index には i が設定されていること.
	// 	for j, c := range cs {
	// 		if dup && c.index != uint(i) {
	// 			fmt.Printf("i=%d, j=%d, index=%d\n", i, j, c.index)
	// 			c.Print()
	// 			panic("ge")
	// 		}
	// 		if j != 0 {
	// 			if s, b := cad.cellcmp(cs[j-1], c); !b || s >= 0 {
	// 				panic("go")
	// 			}
	// 		}
	// 	}
	// }

	if len(ciso) == 0 {
		return []*Cell{}
	}
	cs := ciso[0]
	for i := 1; i < len(ciso); i++ {
		// cs < ciso[i]: c.index
		cs = cad.cellmerge2(cs, ciso[i], dup)
	}
	return cs
}

func (cad *CAD) cellmerge2(cis, cjs []*Cell, dup bool) []*Cell {
	// cis.index < cjs.index
	cret := make([]*Cell, 0, len(cis)+len(cjs))

	i := 0
	j := 0
	for i < len(cis) && j < len(cjs) {
		ci := cis[i]
		cj := cjs[j]
		fusion_improve := true
		if s, ok := cad.cellcmp(cj, ci); s < 0 {
			cret = append(cret, cj)
			j++
			continue
		} else if s > 0 {
			cret = append(cret, ci)
			i++
			continue
		} else if ok {
			// 一致したって........ 次数の低いほうを選ぼう.
			fusion_improve = true
		} else if !dup {
			fusion_improve = false
		} else {
			// 共通根をもつ可能性があるか?
			if ci.index == cj.index {
				ci.Print()
				cj.Print()
				panic("kk")
			}
			hcr := cad.proj[ci.lv].hasCommonRoot(cad, ci.parent, cj.index, ci.index)
			if hcr == PF_EVAL_NO {
				// 共通根は持たない.
				fusion_improve = false
				goto _FUSION_IMPROVE
			}
			if ci.defpoly == nil || cj.defpoly == nil {
				if ci.defpoly == nil && cj.defpoly == nil && ci.intv.inf.Equals(cj.intv.inf) {
					fusion_improve = true
				} else {
					var di, dj *Cell
					if ci.defpoly == nil {
						di = cj
						dj = ci
					} else {
						di = ci
						dj = cj
					}
					switch q := dj.intv.inf.subst_poly(di.defpoly, dj.lv).(type) {
					case NObj:
						fusion_improve = q.IsZero()
					case *Poly:
						fusion_improve = cad.sym_zero_chk(q, dj.parent)
					}
				}
				goto _FUSION_IMPROVE
			} else if ci.defpoly.Equals(cj.defpoly) {
				// 一致した.
				fusion_improve = true
				goto _FUSION_IMPROVE
			}

			if hcr == 1 {
				// @TODO 虚根を持たない，かつ，他に重複がなければ一致が確定する
			}

			fusion_improve = cad.sym_equal(ci, cj)
		}

	_FUSION_IMPROVE:
		if fusion_improve {
			// 一致した
			cj.fusion(ci)
			cret = append(cret, ci)
			i++
			j++
		} else {
			// 一致しないので区間を改善
			for k := 0; ; k++ {
				ci.improveIsoIntv(ci.defpoly, true)
				cj.improveIsoIntv(cj.defpoly, false)
				if s, _ := cad.cellcmp(cj, ci); s < 0 {
					cret = append(cret, cj)
					j++
					break
				} else if s > 0 {
					cret = append(cret, ci)
					i++
					break
				}
				if k > 50 {
					fmt.Printf("dup=%.1v, k=%d\n", dup, k)
					ci.Print()
					cj.Print("cellp")
					panic("dup51!")
				}
			}
		}
	}
	for ; i < len(cis); i++ {
		cret = append(cret, cis[i])
	}
	for ; j < len(cjs); j++ {
		cret = append(cret, cjs[j])
	}

	return cret
}

func (cell *Cell) root_iso_q(cad *CAD, pf ProjFactor, p *Poly) []*Cell {
	// returns (roots, sign(lc(p)))
	fctrs := cad.g.ox.Factor(p)
	cad.stat.fctr++
	// fmt.Printf("root_iso(%v,%d): %v -> %v\n", cell.Index(), pf.index, p, fctrs)
	ciso := make([][]*Cell, fctrs.Len()-1)
	for i := fctrs.Len() - 1; i > 0; i-- {
		ff := fctrs.getiList(i)
		q := ff.getiPoly(0)
		ciso[i-1] = make([]*Cell, 0, len(q.c)-1)
		r := mult_t(ff.getiInt(1).Int64())
		if len(q.c) == 2 {
			c := NewCell(cad, cell, pf.Index())
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
			c.multiplicity[pf.Index()] = r
			ciso[i-1] = append(ciso[i-1], c)
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
			for j := 0; j < len(roots); j++ {
				rr := roots[j]
				c := NewCell(cad, cell, pf.Index())
				c.intv.inf = rr.low
				if rr.point { // fctr しているからこのルートは通らないはず.
					c.intv.sup = rr.low
				} else {
					c.intv.sup = rr.low.upperBound()
				}
				c.sgn_of_left = sgn
				sgn *= -1
				c.multiplicity[pf.Index()] = r
				c.defpoly = q
				c.ex_deg = len(q.c) - 1
				ciso[i-1] = append(ciso[i-1], c)
			}
		}
	}

	// sort...
	return cad.cellmerge(ciso, false)
}

func (cell *Cell) reduce(pp RObj) RObj {
	// 次数を下げる.
	p, ok := pp.(*Poly)
	if !ok {
		return pp
	}
	for c := cell; c.lv >= 0; c = c.parent {
		var q RObj
		if c.defpoly != nil {
			if _, ok := c.defpoly.c[len(c.defpoly.c)-1].(NObj); !ok {
				// 正規化されていない
				continue
			}
			q = p.reduce(c.defpoly)
		} else {
			q = c.intv.inf.subst_poly(p, c.lv)
		}
		if qq, ok := q.(*Poly); ok {
			p = qq.primpart()
		} else {
			return q
		}
	}
	return p
}

func (cell *Cell) make_cells_try1(cad *CAD, pf ProjFactor, pr RObj) (*Poly, []*Cell, sign_t) {
	// returns (p, c, s)
	// 子供セルが作れたら， p=nil, s=pfに cell 代入したときの主係数の符号
	// 子供セルが作れなかったら p != nil, (c,s) は使わない
	// fmt.Printf("  make_cells_try1(%v, %d) pr=%v->%v\n", cell.Index(), pf.Index(), pf.P(), pr)

	switch p := pr.(type) {
	case *Poly:
		if p.isUnivariate() && p.lv == pf.Lv() {
			// 他の変数が全部消えた.
			return nil, cell.root_iso_q(cad, pf, p), sign_t(p.Sign())
		} else if p.lv != pf.Lv() {
			// 主変数が消えて定数になった.
			switch pf.evalCoeff(cad, cell, 0) {
			case LT:
				return nil, []*Cell{}, -1
			case GT:
				return nil, []*Cell{}, +1
			case EQ:
				return nil, []*Cell{}, 0
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

func (cell *Cell) make_cells(cad *CAD, pf ProjFactor) ([]*Cell, sign_t) {
	// cell: 持ち上げるセル
	// pf: 対象の射影因子
	// returns (children, sign(lc))

	p := pf.P()
	if sgn := pf.Sign(); sgn != 0 {
		return []*Cell{}, sgn
	}

	if cell.de || cell.defpoly != nil {
		// projection factor の情報から，ゼロ値を決める
		pp := NewPoly(p.lv, len(p.c))
		up := false
		for i := 0; i < len(p.c); i++ {
			if pf.evalCoeff(cad, cell, i) == EQ {
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
			pp := c.intv.inf.subst_poly(p, c.lv)
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

	q := cell.reduce(p)

	// @TODO ここで Z 上因数分解できるケースがある.
	// adam2-1, 98-23, 85-19.

	p, c, s = cell.make_cells_try1(cad, pf, q) // とりあえず簡単化してみる
	if p == nil {
		return c, s
	}
	return cell.make_cells_i(cad, pf, p)
}

func (cell *Cell) nintv_divide(prec uint) bool {
	ret := false
	for _, point := range []float64{0.5, 0.25, 0.75, 0.125, 0.875} {
		mid := cell.nintv.mid(point, prec)
		vv := NewIntervalFloat(mid, prec)
		pp := cell.defpoly.toIntv(prec).(*Poly)
		p2 := cell.parent.subst_intv(pp, prec).(*Poly)
		p3 := p2.Subst(vv, cell.lv).(*Interval)
		if !p3.ContainsZero() {
			iv := newInterval(prec)
			if p3.Sign()*int(cell.sgn_of_left) > 0 {
				iv.inf.Set(mid)
				iv.sup.Set(cell.nintv.sup)
			} else {
				iv.inf.Set(cell.nintv.inf)
				iv.sup.Set(mid)
			}
			cell.nintv = iv
			ret = true
		}
	}
	return ret
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

	cell.nintv_divide(prec)
	return cell.nintv
}

func (cell *Cell) _subst_intv(p *Poly, prec uint) RObj {
	for cell.lv > p.lv {
		cell = cell.parent
	}
	for i, _c := range p.c {
		if c, ok := _c.(*Poly); ok {
			p.c[i] = cell._subst_intv(c, prec)
		}
	}
	if err := p.valid(); err != nil {
		panic(err)
	}
	if p.lv != cell.lv {
		return p
	}
	x := cell.getNumIsoIntv(prec)
	qq := p.Subst(x, p.lv)
	return qq
}

func (cell *Cell) subst_intv(p *Poly, prec uint) RObj {
	pp := p.toIntv(prec).(*Poly)
	for cell.lv > p.lv {
		cell = cell.parent
	}

	return cell._subst_intv(pp, prec)
}

func (cell *Cell) subst_intv_nozero(cad *CAD, p *Poly) *Poly {
	prec := uint(53)
	pp := cell.subst_intv(p, prec).(*Poly)
	if !pp.isUnivariate() {
		panic("invalid")
	}

	for i, c := range pp.c {
		if c.(*Interval).ContainsZero() && !c.(*Interval).IsZero() {
			if cad.sym_zero_chk(p.c[i].(*Poly), cell) {
				pp.c[i] = zero.toIntv(prec)
				p.c[i] = zero
			} else {
				// 精度をあげる
				for prec = uint(53 * 2); prec < 1000; prec += 53 {
					cell.improveIsoIntv(p, true)
					pp.c[i] = cell.subst_intv(p.c[i].(*Poly), prec)
					if !pp.c[i].(*Interval).ContainsZero() {
						break
					}
				}
				if prec >= 1000 {
					panic(fmt.Sprintf("unimplemented; %d", prec))
				}
			}
		}
	}
	return pp
}

func (cell *Cell) make_cells_i(cad *CAD, pf ProjFactor, porg *Poly) ([]*Cell, sign_t) {
	prec := uint(53)

	p := porg.Clone()
	pp := cell.subst_intv_nozero(cad, p)

	sgn := sign_t(pp.Sign())
	cells := make([][]*Cell, 0)

	// 定数項のゼロは取り除く.
	for i, c := range pp.c {
		if !c.IsZero() {
			if i > 0 {
				if i == len(pp.c)-1 {
					// numeric 確定..
					c := NewCell(cad, cell, pf.Index())
					c.de = cell.de
					c.multiplicity[pf.Index()] = mult_t(i)
					c.intv.inf = zero
					c.intv.sup = zero
					return []*Cell{c}, sgn
				}
				qq := NewPoly(pp.lv, len(pp.c)-i)
				copy(qq.c, pp.c[i:])
				pp = qq

				q := NewPoly(p.lv, len(p.c)-i)
				copy(q.c, p.c[i:])
				p = q

				c := NewCell(cad, cell, pf.Index())
				c.de = cell.de
				c.multiplicity[pf.Index()] = mult_t(i)
				c.intv.inf = zero
				c.intv.sup = zero
				cells = append(cells, []*Cell{c})
			}
			break
		}
	}

	// even polynomial.
	even := 1
	for pp.isEven() {
		even *= 2
		qq := NewPoly(pp.lv, pp.deg()/2+1)
		q := NewPoly(p.lv, len(qq.c))
		for i := 0; i < len(qq.c); i++ {
			qq.c[i] = pp.c[2*i]
			q.c[i] = p.c[2*i]
		}

		pp = qq
		p = q
	}

	c, err := cell.root_iso_i(cad, pf, p, pp, prec, 1)
	if err == nil {
		cells = append(cells, c)
	} else if hmf := pf.hasMultiFctr(cad, cell); hmf != 0 {

		var ps []*cadSqfr
		if hmf == PF_EVAL_YES && p.deg() == 2 {
			// 2次のときは微分したら良いよ
			jk := p.diff(p.lv).(*Poly)
			jk, _ = jk.pp()
			ps = []*cadSqfr{newCadSqfr(nil, jk, 2)}
		} else {
			ps = cad.sym_sqfr2(p, cell)
		}

		if hmf == PF_EVAL_YES && len(ps) == 1 && ps[0].r == 1 {
			// 分解されなかった...
			fmt.Printf("p=%v: hmf=%v\n", p, hmf)
			for _, poly_mul := range ps {
				fmt.Printf("p=(%v)^(%d)\n", poly_mul.p, poly_mul.r)
			}
			panic("??")
		}

		// もし分解されていなかったら...
		for _, poly_mul := range ps {
			p := poly_mul.p

			for prec = uint(53); ; prec += 53 {
				pp := cell.subst_intv_nozero(cad, p)
				c, err := cell.root_iso_i(cad, pf, p, pp, prec, poly_mul.r)
				if err == nil {
					cells = append(cells, c)
					break
				}
				cell.improveIsoIntv(poly_mul.p, true)
			}
		}
	} else {
		// 係数の符号が確定している.
		for prec += 60; prec < 500; prec += 53 {
			pp = cell.subst_intv_nozero(cad, p)
			c, err := cell.root_iso_i(cad, pf, porg, pp, prec, 1)
			if err == nil {
				cells = append(cells, c)
				break
			}
			cell.improveIsoIntv(porg, true)
		}
	}

	ccc := cad.cellmerge(cells, false)
	for even > 1 {
		cx := make([]*Cell, 0, len(ccc))
		for i := len(ccc) - 1; i >= 0; i-- {
			c := ccc[i]
			if c.Sign() > 0 {
				c.Square(even) // 2乗する
				cx = append(cx, c.Neg())
			}
		}
		for _, c := range ccc {
			if c.Sign() >= 0 {
				cx = append(cx, c)
			}
		}
		ccc = cx
		break
	}

	return ccc, sgn
}

func (cell *Cell) Square(even int) {
	p := NewPoly(cell.lv, cell.defpoly.deg()*even+1)
	for i := 0; i < len(p.c); i++ {
		p.c[i] = zero
	}
	for i := 0; i < len(cell.defpoly.c); i++ {
		p.c[i*even] = cell.defpoly.c[i]
	}
	cell.defpoly = p
	for e := 1; e < even; e *= 2 {
		cell.nintv.inf.Sqrt(cell.nintv.inf)
		cell.nintv.sup.Sqrt(cell.nintv.sup)
	}
}

func (cell *Cell) Neg() *Cell {
	c := new(Cell)
	*c = *cell
	c.nintv = newInterval(cell.nintv.Prec())
	c.nintv.inf.Neg(cell.nintv.sup)
	c.nintv.sup.Neg(cell.nintv.inf)
	return c
}

func (cell *Cell) Sign() int {
	if cell.intv.inf != nil {
		if cell.intv.inf.Sign() > 0 {
			return 1
		} else if cell.intv.sup.Sign() < 0 {
			return -1
		}
	} else {
		if cell.nintv.inf.Sign() > 0 {
			return 1
		} else if cell.nintv.sup.Sign() < 0 {
			return -1
		}
	}
	return 0
}

func (cell *Cell) root_iso_i(cad *CAD, pf ProjFactor, porg, pp *Poly, prec uint, multiplicity mult_t) ([]*Cell, error) {
	// porg: 区間代入前
	// pp: interval polynomial

	ans, err := pp.iRealRoot(prec, 1000)
	cad.stat.irealroot++
	if err != nil {
		cad.log(3, "irealroot failed: %s\n", err.Error())
		return nil, err
	}
	cad.stat.irealroot_ok++

	sgn := pp.Sign()
	cells := make([]*Cell, len(ans), len(ans)+1)
	for i := len(ans) - 1; i >= 0; i-- {
		sgn *= -1
		c := NewCell(cad, cell, pf.Index())
		c.nintv = ans[i]
		c.sgn_of_left = sign_t(sgn)
		c.de = true
		c.multiplicity[pf.Index()] = multiplicity
		c.defpoly = porg
		cells[i] = c
	}
	return cells, nil
}

func (cell *Cell) improveIsoIntv(p *Poly, parent bool) {
	// 分離区間の改善
	if parent && cell.lv >= 0 {
		cell.parent.improveIsoIntv(p, parent)
	}
	if cell.defpoly == nil {
		return
	}
	if p != nil && p.Deg(cell.lv) <= 0 {
		// 関係ないセルは改善しなくても良い
		return
	}

	switch l := cell.intv.inf.(type) {
	case *BinInt:
		// binint ということは, realroot の出力であり，１変数多項式
		for ii := 0; ii < 30; ii++ {
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
				break
			}
			l = cell.intv.inf.(*BinInt)
		}
		cell.nintv = nil
		return
	}
	if cell.intv.inf != nil {
		cell.nintv = nil
		return
	}

	var prec uint
	for prec = cell.nintv.Prec() + 64; prec < 1000; prec += 60 {
		if cell.nintv_divide(prec) {
			return
		}
		cell.parent.improveIsoIntv(nil, true)
	}

	cell.Print()
	panic(fmt.Sprintf("unimplemented: prec=%d", prec))
}

func (cell *Cell) valid(cad *CAD) error {
	// cell の妥当性評価 for debug
	if cell.lv >= 0 {
		if cell.defpoly == nil {
			if cell.intv.inf != cell.intv.sup {
				return fmt.Errorf("defpoly=nil but inf != sup %v", cell.Index())
			}
		}
	}

	idx := cell.Index()
	if len(idx) > 0 && (cad.stage < 2 || cad.q[cell.lv] == q_free) {
		c := cad.root
		for _, x := range idx {
			c = c.children[x]
		}
		if c != cell {
			return fmt.Errorf("cell=%v... but ...", idx)
		}
	}

	if cell.children != nil {
		nt := 0
		nf := 0
		ntc := 0
		nfc := 0
		children := make([]*Cell, 0, len(cell.children))
		for i, c := range children {
			if c.index != uint(i) {
				return fmt.Errorf("index invalid. expect=%d, actual=%v", i, c.Index())
			}
		}
		for i := 0; i < len(cell.children); i += 2 {
			children = append(children, cell.children[i])
		}
		for i := 1; i < len(cell.children); i += 2 {
			children = append(children, cell.children[i])
		}
		for _, c := range children {
			if err := c.valid(cad); err != nil {
				return err
			}
			if c.truth == t_true {
				nt++
				if c.children != nil {
					ntc++
				}

			} else if c.truth == t_false {
				nf++
				if c.children != nil {
					nfc++
				}
			}
			if c.parent != cell && (cad.stage < 2 || cell.truth < 0) {
				return fmt.Errorf("%v.parent=%v != %v invalid [%p:%p]", c.Index(), c.parent.Index(), cell.Index(), c.parent, cell)
			}
		}

		errmes := ""
		if cad.q[cell.lv+1] == q_forall {
			if nfc > 1 {
				errmes = fmt.Sprintf("forall + many false cells")
			} else if nf > 0 && cell.truth != t_false {
				errmes = fmt.Sprintf("forall + false false false")
			} else if nf == 0 && nt > 0 && cell.truth != t_true {
				errmes = fmt.Sprintf("forall + true true true")
			}
		} else if cad.q[cell.lv+1] == q_exists {
			if ntc > 1 {
				errmes = fmt.Sprintf("exists + many false cells")
			} else if nt > 0 && cell.truth != t_true {
				errmes = fmt.Sprintf("exists + true true true")
			} else if nt == 0 && nf > 0 && cell.truth != t_false {
				errmes = fmt.Sprintf("exists + false false false")
			}
		}
		if errmes != "" {
			cell.Print("cellp")
			cell.Print("signatures")
			return fmt.Errorf("%s: index=%v, truth=%d, true=(%d/%d), false=(%d/%d)", errmes, cell.Index(), cell.truth, ntc, nt, nfc, nf)
		}
	}

	return nil
}
