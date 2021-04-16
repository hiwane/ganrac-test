package ganrac

import (
	"fmt"
	"os"
	"sort"
)

// solution formula construction for truth invariant CAD's
// Christopher W. Brown. thesis, 1999
type CADSfc struct {
	cad        *CAD
	lt, lf, t3 []*Cell
	evaltbl    [][]bool
	freen      int
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

func (sfc *CADSfc) pdqv22_split_leaf(cells []*Cell, min, max int) ([]*Cell, int) {
	t := 0
	ret := make([]*Cell, 0)
	var ct, cf *Cell
	for i := min; i < max; i++ {
		c := cells[i]
		if c.children != nil && int(c.lv) < sfc.freen-1 {
			ret = append(ret, c.children...)
		} else if c.truth == 0 {
			t |= 0x02
			cf = c
		} else if c.truth == 1 {
			t |= 0x01
			ct = c
		} else {
			c.Print(os.Stdout)
			panic("s")
		}
	}

	if ct != nil {
		sfc.lt = append(sfc.lt, ct)
	}
	if cf != nil {
		sfc.lf = append(sfc.lf, cf)
	}

	// @TODO

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
		if lv == len(sfc.cad.q) {
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
			// for i = min; i < max; i++ {
			// 	if (k->v[i]->children == NULL) {
			// 		synstack_push(sinf->t3cell, k->v[i]);
			// 	}
			// }
		}
	}

	return t
}

func (sfc *CADSfc) pdq(root *Cell) int {
	// return 0 if projection definable
	// return 1 if projection undefinable
	// return 2 undetermined
	cells := []*Cell{root}
	t := sfc.pdqv22(0, cells, 0, 1)
	fmt.Printf("pdq() t=%x, [%x %x, %x]\n", t, t&0xc, t&0x8, 0x4|0x8)
	if (t & (0x4 | 0x8)) == 0 {
		return 0
	} else if (t % 0x8) != 0 {
		return 1
	} else {
		return 2
	}
}

func (sfc *CADSfc) gen_atoms() []*sfcAtom {
	a := make([]*sfcAtom, 0) // @TODO
	for i := 0; i < sfc.freen; i++ {
		for j := 0; j < len(sfc.cad.proj[i].pf); j++ {
			for _, op := range []OP{LE, GE, NE, EQ, LT, GT} {
				a = append(a, &sfcAtom{i, sfc.cad.proj[i].pf[j].index, op})
			}
		}
	}
	fmt.Printf("genatoms=%d\n", len(a))
	return a
}

func (sfc *CADSfc) eval(ctable []*Cell, ta *sfcAtom) bool {
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
	// init(a,b,c,x);F=ex([x],a*x^2+b*x+c==0); C=cadinit(F); cadproj(C); cadlift(C); print(C, "cells"); cadsfc(C);
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
			fmt.Printf("t=%v, lenh=%v, maxn=%d\n", t, len(h), maxn)
			if t != nil {
				hmin = make([]int, len(t))
				copy(hmin, t)
				if len(hmin) == hlen {
					return hmin
				}
				maxn = len(hmin)
			}
		}
		if hmin == nil {
			panic("naze?")
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

	for _, ss := range s {
		fmt.Printf(" hs() %v\n", ss)
	}

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
	fmt.Printf("ai=%v, %v [%d]\n", ai, ctable[0].Index(), len(sfc.lf))
	fmt.Printf("la[ 6]=%v\n", la[6])
	fmt.Printf("la[10]=%v\n", la[10])

	// ai のうち, false cell が false になるものたちを抽出
	s := make([][]int, 0, len(sfc.lf))
	for _, c := range sfc.lf {
		for u := c; u.lv >= 0; u = u.parent {
			ctable[u.lv] = u
		}
		ctable = ctable[:c.lv+1]

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
		fmt.Printf(".. impl=%v\n", impls)
		if len(impls) > 0 {
			s = append(s, impls)
		} else {
			panic("e")
		}
	}

	return sfc.hitting_set(s)
}

func (sfc *CADSfc) simplesf(la []*sfcAtom) Fof {

	impls := make([][]int, 0)
	ctable := make([]*Cell, sfc.freen)
	for _, c := range sfc.lt {

		for ci := c; ci.lv >= 0; ci = ci.parent {
			ctable[ci.lv] = ci
		}
		ctable = ctable[:c.lv+1]

		// すでにある implicant に捕獲されているか
		if sfc.captured(ctable, la, impls) {
			fmt.Printf("c=%v captured %v\n", c.Index(), impls)
			continue
		}

		// c を捕獲し, すべての false cell を含まない implicant を求める
		ai := sfc.implcons(ctable, la)
		impls = append(impls, ai)
		fmt.Printf("c=%v, new impl %v\n", c.Index(), ai)
	}

	// another hitting problem

	var fml Fof = falseObj
	for _, impl := range impls {
		var ff Fof = trueObj
		for _, atm := range impl {
			ta := la[atm]
			ff = NewFmlAnd(ff, NewAtom(sfc.cad.proj[ta.lv].pf[ta.index].p, ta.op))
		}
		fml = NewFmlOr(fml, ff)
	}
	return fml
}

func (cad *CAD) Sfc() (Fof, error) {
	if cad.root.truth == 0 {
		return falseObj, nil
	} else if cad.root.truth == 1 {
		return trueObj, nil
	}

	sfc := NewCADSfc(cad)
	t := sfc.pdq(cad.root)
	if len(sfc.lt) == 0 {
		return falseObj, nil
	}
	if len(sfc.lf) == 0 {
		return trueObj, nil
	}
	if t >= 1 {
		panic(fmt.Sprintf("unsupported pdq=%d", t))
	}

	la := sfc.gen_atoms()
	return sfc.simplesf(la), nil
}
