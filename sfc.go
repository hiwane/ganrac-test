package ganrac

import (
	"fmt"
	"sort"
	"time"
)

// solution formula construction for truth invariant CAD's
// Christopher W. Brown. thesis, 1999

const (
	// pdq() の復帰
	SFC_PROJ_DEFINABLE   = 0
	SFC_PROJ_UNDEFINABLE = 1
	SFC_PROJ_UNDET       = 2
)

type CADSfc struct {
	cad     *CAD
	lt      []*Cell // true leaf cell
	lf      []*Cell // false leaf cell
	t3      []*Cell
	evaltbl [][]bool
	freen   int          // number of free variables
	cpair   [][][4]*Cell // conflicting pair
}

type sfcAtom struct {
	lv    int
	index uint
	op    OP
}

func NewCADSfc(cad *CAD) *CADSfc {
	sfc := new(CADSfc)
	sfc.cad = cad

	for sfc.freen = 0; sfc.freen < len(sfc.cad.q); sfc.freen++ {
		if sfc.cad.q[sfc.freen] >= 0 {
			break
		}
	}

	sfc.evaltbl = make([][]bool, 3)
	// X     LT      EQ     LE    GT     NE     GE
	sfc.evaltbl[0] = []bool{false, true, false, true, false, true, false} // sgn<0
	sfc.evaltbl[1] = []bool{false, false, true, true, false, false, true} // sgn=0
	sfc.evaltbl[2] = []bool{false, false, false, false, true, true, true} // sgn>0

	return sfc
}

func (sfc *CADSfc) add_conflicting_pairs(cells []*Cell, min, max int) {
	nt := make([]*Cell, 0, max-min)
	nf := make([]*Cell, 0, max-min)
	for i := min; i < max; i++ {
		c := cells[i]
		if c.truth == t_true {
			nt = append(nt, c)
		} else if c.truth == t_false {
			nf = append(nf, c)
		}
	}

	for _, ct := range nt {
		for _, cf := range nf {
			sfc.add_conflicting_pair(ct, cf)
		}
	}
}

func (sfc *CADSfc) add_conflicting_pair(ctrue, cfalse *Cell) {
	var a [4]*Cell
	a[2] = ctrue
	a[3] = cfalse

	for ctrue.parent != cfalse.parent {
		ctrue = ctrue.parent
		cfalse = cfalse.parent
	}
	a[0] = ctrue
	a[1] = cfalse

	sfc.cpair[ctrue.lv] = append(sfc.cpair[ctrue.lv], a)
}

/*
 * split K into the lists K L and K N , the leaf nodes and non-leaf nodes, respec-
 tively, in K .
*/
func (sfc *CADSfc) pdqv22_split_leaf(cells []*Cell, min, max int) ([]*Cell, int) {
	t := 0
	ret := make([]*Cell, 0)
	var ct, cf *Cell
	for i := min; i < max; i++ {
		c := cells[i]
		if c.children != nil && int(c.lv) < sfc.freen-1 {
			// leaf ではない
			ret = append(ret, c.children...)
		} else if c.truth == 0 {
			t |= 0x02
			cf = c
		} else if c.truth == 1 {
			t |= 0x01
			ct = c
		} else {
			c.Print()
			c.parent.Print("signatures")
			panic("internal error")
		}
	}

	if ct != nil {
		sfc.lt = append(sfc.lt, ct)
	}
	if cf != nil {
		sfc.lf = append(sfc.lf, cf)
	}

	if t == (0x1 | 0x2) {
		sfc.add_conflicting_pairs(cells, min, max)
	}

	return ret, t
}

func (sfc *CADSfc) cmp_signature(c, d *Cell) int {
	for k := 0; k < len(c.signature); k++ {
		if c.signature[k] < d.signature[k] {
			return -1
		} else if c.signature[k] > d.signature[k] {
			return +1
		}
	}
	return 0
}

func (sfc *CADSfc) pdqv22(lv int, cells []*Cell, min, max int) int {
	// cells[min]..cells[max]が同じシグネチャ
	// returns (内部ノード, T値)

	cs, t := sfc.pdqv22_split_leaf(cells, min, max)
	if t == 0x3 {
		if lv == sfc.freen {
			t |= 0x8
		} else {
			t |= 0x4
		}
	}
	if len(cs) == 0 {
		return t
	}

	// 子供をソートして....
	sort.Slice(cs, func(i, j int) bool {
		return sfc.cmp_signature(cs[i], cs[j]) < 0
	})

	j := 0
	tl := t
	for i := j + 1; i < len(cs); i++ {
		if sfc.cmp_signature(cs[j], cs[i]) != 0 {
			t |= sfc.pdqv22(lv+1, cs, j, i)
			if (t & 0x4) != 0 { // projection undefinable が確定
				return t
			}
			j = i
		}
	}
	t |= sfc.pdqv22(lv+1, cs, j, len(cs))
	if tl != 0 && t == 0x3 {
		/*
		 * level  = $lv で leaf なものがいた.
		 * level >= $lv で conflict
		 *
		 * e.g. example(easy7); opt(nproj,y); opt(nlift,y);
		 *      truth
		 *      (4)    0     (-,+,0)
		 *      (6,6)  1     (-,+,0,-,+,0,0)
		 */
		t |= 0x3
		if (tl&0x4) == 0 && (t&0x08) == 0 {
			/* min, max 間で leaf なものを t3cell に追加 */
			for i := min; i < max; i++ {
				if cs[i].children == nil {
					sfc.t3 = append(sfc.t3, cs[i])
				}
			}
		}
	}

	return t
}

/*
 * projection definable query
 *
 * return:
 * - SFC_PROJ_DEFINABLE if projection definable
 * - SFC_PROJ_UNDEFINABLE if projection undefinable
 * - SFC_PROJ_UNDET     otherwise / undetermined
 */
func (sfc *CADSfc) pdq() int {
	sfc.lt = make([]*Cell, 0)
	sfc.lf = make([]*Cell, 0)
	sfc.t3 = make([]*Cell, 0)

	sfc.cpair = make([][][4]*Cell, sfc.freen)
	for i := 0; i < sfc.freen; i++ {
		sfc.cpair[i] = make([][4]*Cell, 0)
	}

	cells := []*Cell{sfc.cad.root}

	sfc.cad.log(2, "pdq() start!>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
	t := sfc.pdqv22(0, cells, 0, 1)
	sfc.cad.log(2, "  pdq() t=%#x, [%#x %#x] [t=%d,f=%d]\n", t, t&0xc, t&0x8,
		len(sfc.lt), len(sfc.lf))
	if (t & (0x4 | 0x8)) == 0 {
		return SFC_PROJ_DEFINABLE
	} else if (t & 0x8) != 0 {
		return SFC_PROJ_UNDEFINABLE
	} else {
		return SFC_PROJ_UNDET
	}
}

func (sfc *CADSfc) gen_atoms() []*sfcAtom {
	a := make([]*sfcAtom, 0) // @TODO
	for i := 0; i < sfc.freen; i++ {
		for j, pf := range sfc.cad.proj[i].gets() {
			if pf.Sign() != 0 {
				continue
			}
			for _, op := range []OP{EQ, LE, GE, NE, LT, GT} {
				if pf.Index() != uint(j) {
					panic(fmt.Sprintf("hey; j=%d, pf=%d: %v", j, pf.Index(), pf.P()))
				}
				a = append(a, &sfcAtom{i, uint(j), op})
			}
		}
	}
	sfc.cad.log(2, "  genatoms=%d\n", len(a))
	return a
}

func (sfc *CADSfc) eval(ctable []*Cell, ta *sfcAtom) bool {
	if len(ctable[ta.lv].signature) <= int(ta.index) {
		sfc.cad.log(2, "sfc.eval() %v: %d\n", ta, len(ctable[ta.lv].signature))
		sfc.cad.log(2, "root=%p:%p, %p:%p\n",
			sfc.cad.root, sfc.cad.root.children[2],
			ctable[ta.lv].parent, ctable[ta.lv])
		ctable[ta.lv].Print()
	}
	return sfc.evaltbl[ctable[ta.lv].signature[ta.index]+1][ta.op]
}

func (sfc *CADSfc) captured(ctable []*Cell, la []*sfcAtom, impls [][]int) bool {
	for _, impl := range impls {

		for _, im := range impl {
			ta := la[im]
			if len(ctable) <= ta.lv {
				goto _NEXT
			}

			if !sfc.eval(ctable, ta) {
				goto _NEXT
			}
		}
		return true

	_NEXT:
	}
	return false
}

func (sfc *CADSfc) isIncluded(s, h []int) bool {
	// s はソート済みで， s の要素が h に含まれていたら true を返す
	for _, hv := range h {
		for _, sv := range s {
			if hv == sv {
				return true
			}
		}
	}
	return false
}

func (sfc *CADSfc) _hitting_set(s [][]int, h []int, idx, maxn int) []int {
	// s[idx] 以降をたどる
	// maxn これまで見つかっている最適解の長さ

	var si []int

	for ; idx < len(s); idx++ {
		// 前処理でもう 含まれているものは飛ばす
		si = s[idx]
		if !sfc.isIncluded(si, h) {
			break
		}
	}
	// vars(a,b,c,x);F=ex([x],a*x^2+b*x+c==0); C=cadinit(F); cadproj(C); cadlift(C); print(C, "signatures"); cadsfc(C);
	if idx == len(s) {
		return h
	} else if len(h)+1 >= maxn {
		// もう今より良い解は見つからない
		return nil
	} else {
		// si の要素を追加して試す.
		hlen := len(h) + 1
		h = append(h, 0)
		var hmin []int
		for _, sv := range si {
			h[hlen-1] = sv
			if len(h) != hlen {
				panic("boo")
			}
			t := sfc._hitting_set(s, h, idx+1, maxn)
			if t != nil {
				hmin = make([]int, len(t))
				copy(hmin, t)
				if len(hmin) == hlen {
					return hmin
				}
				maxn = len(hmin)
			}
		}
		return hmin
	}
}

func (sfc *CADSfc) hitting_set(s [][]int) []int {
	if len(s) == 0 {
		return []int{}
	}

	sort.Slice(s, func(i, j int) bool {
		return len(s[i]) < len(s[j])
	})

	return sfc._hitting_set(s, []int{}, 0, len(s[len(s)-1])+10)
}

func (sfc *CADSfc) implcons(ctable []*Cell, la []*sfcAtom) []int {

	// la のうち, c が true となる cell をすべて抽出
	ai := make([]int, 0, len(la)/2)
	for i, ta := range la {
		if len(ctable) <= ta.lv {
			continue
		}
		if sfc.eval(ctable, ta) {
			ai = append(ai, i)
		}
	}
	// ai のうち, false cell が false になるものたちを抽出
	s := make([][]int, 0, len(sfc.lf))
	for _, c := range sfc.lf {
		ctable = sfc.set_ctable(c, ctable)

		impls := make([]int, 0, len(ai)/2)
		for _, aa := range ai {
			ta := la[aa]
			if len(ctable) <= ta.lv {
				continue
			}
			if !sfc.eval(ctable, ta) {
				impls = append(impls, aa)
			}
		}
		if len(impls) > 0 {
			s = append(s, impls)
		} else {
			panic("e")
		}
	}

	return sfc.hitting_set(s)
}

func (sfc *CADSfc) set_ctable(cell *Cell, ctable []*Cell) []*Cell {
	for len(ctable) <= int(cell.lv) {
		ctable = append(ctable, cell)
	}
	for ci := cell; ci.lv >= 0; ci = ci.parent {
		ctable[ci.lv] = ci
	}
	return ctable[:cell.lv+1]
}

func (sfc *CADSfc) another_hp(impls [][]int, la []*sfcAtom, ctable []*Cell) [][]int {

	s := make([][]int, 0)
	for _, c := range sfc.lt {
		ctable = sfc.set_ctable(c, ctable)
		sj := make([]int, 0)
		for j, impl := range impls {
			b := true
			for _, aa := range impl {
				ta := la[aa]
				if len(ctable) <= ta.lv {
					b = false
					break
				}
				if !sfc.eval(ctable, ta) {
					b = false
					break
				}

			}
			if b {
				sj = append(sj, j)
			}
		}
		s = append(s, sj)
	}

	hs := sfc.hitting_set(s)
	ret := make([][]int, 0, len(hs))
	for _, h := range hs {
		ret = append(ret, impls[h])
	}
	return ret

}

func (sfc *CADSfc) simplesf(la []*sfcAtom) Fof {

	impls := make([][]int, 0)
	ctable := make([]*Cell, sfc.freen)
	for _, c := range sfc.lt {
		ctable = sfc.set_ctable(c, ctable)

		// すでにある implicant に捕獲されているか
		if sfc.captured(ctable, la, impls) {
			continue
		}

		// c を捕獲し, すべての false cell を含まない implicant を求める
		ai := sfc.implcons(ctable, la)
		impls = append(impls, ai)
	}

	// another hitting problem
	if len(impls) > 1 {
		impls = sfc.another_hp(impls, la, ctable)
	}

	var fml Fof = falseObj
	for _, impl := range impls {
		var ff Fof = trueObj
		for _, atm := range impl {
			ta := la[atm]
			ff = NewFmlAnd(ff, NewAtom(sfc.cad.proj[ta.lv].get(ta.index).P(), ta.op))
		}
		fml = NewFmlOr(fml, ff)
	}
	return fml
}

func (cad *CAD) Sfc() (Fof, error) {
	if cad.root.truth == t_false {
		return falseObj, nil
	} else if cad.root.truth == t_true {
		return trueObj, nil
	} else if cad.stage != 2 {
		return nil, fmt.Errorf("invalid stage")
	}
	tm_start := time.Now()

	sfc := NewCADSfc(cad)
	t := sfc.pdq()
	if len(sfc.lt) == 0 {
		return falseObj, nil
	}
	if len(sfc.lf) == 0 {
		return trueObj, nil
	}
	for ccc := 0; t != SFC_PROJ_DEFINABLE; ccc++ {
		if t == SFC_PROJ_UNDEFINABLE && ccc != 0 {
			return sfc.make_pdf()
		} else {
			// partial CAD な部分でひかかったので，
			// fullcad なら projection definable かもしれない

			// 論文上 undef が復帰することになっているけど
			// 自由変数でみたときの葉でないセルで真偽値が決まる =>
			// 真偽値は同じになりそうなものだけど...
			err := sfc.lift_conlicts(cad)
			if err != nil {
				return nil, err
			}
			sfc = NewCADSfc(cad)
			t = sfc.pdq()
		}
		// t = sfc.pdq()
		// if ccc > 5 {
		// 	panic("ho")
		// }
	}

	la := sfc.gen_atoms()
	fof := sfc.simplesf(la)
	cad.stat.tm[2] = time.Since(tm_start)
	return fof, nil
}
