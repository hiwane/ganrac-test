package ganrac

// solution formula construction for truth invariant CAD's
// Christopher W. Brown. thesis, 1999

import (
	"fmt"
)

func (sfc *CADSfc) add_ds_proj(lv Level, pf ProjFactor) {
	p := pf.P()

	for p.deg() > 1 {
		p = p.diff(lv).(*Poly)
		sfc.cad.log(3, "    add_ds_proj: %v\n", p)
		sfc.cad.addPoly(p, false)
	}
}

func (sfc *CADSfc) construct_lab(lv Level) []int {
	// S4.4.1 Adding i-level polys to "remove" lv-level conflicting pairs
	if len(sfc.cpair[lv]) == 0 {
		return nil
	}

	n := sfc.cad.proj[lv].Len()

	s := make([][]int, 0, len(sfc.cpair[lv]))
	for _, cs := range sfc.cpair[lv] {
		if cs[0].parent != cs[1].parent || cs[0].lv != lv || cs[2].truth != t_true || cs[3].truth != t_false {
			sfc.cad.log(0, "    construct_lab() %v %d/%d, %d %d\n", cs[0].parent == cs[1].parent, lv, cs[0].lv, cs[2].truth, cs[3].truth)
			panic("@1")
		}

		// @TODO 2回め以降の make_pdf() の実行では，すでに解消されている可能性がある?

		var d int
		if cs[0].index < cs[1].index {
			d = 1
		} else {
			d = -1
		}

		parent := cs[0].parent
		si := make([]int, 0, n/2)
		for j := 0; j < n; j++ {
			sgn := cs[0].signature[j]
			for k := int(cs[0].index) + d; k != int(cs[1].index); k += d {
				cc := parent.children[k]
				if cc.signature[j] != sgn {
					si = append(si, j)
					break
				}
			}
		}
		s = append(s, si)
	}

	return sfc.hitting_set(s)
}

func (sfc *CADSfc) make_pdf() (Fof, error) {
	// S4.4.2 Removing all conflicting pairs

	proj_num := make([]int, sfc.freen)
	for i := Level(sfc.freen - 1); i >= 0; i-- {
		proj_num[i] = sfc.cad.proj[i].Len()
	}

	for lv := Level(sfc.freen - 1); lv >= 0; lv-- {
		sfc.cad.log(2, "==================================================\n")
		sfc.cad.log(2, "== makepdf lv=%d <%v>\n", lv, proj_num)
		sfc.cad.log(2, "==================================================\n")
		// construct a hitting set
		//    for {l(a,b) | (a,b) in the set of all lv-level conflicting pairs}
		lab := sfc.construct_lab(lv)
		if len(lab) == 0 {
			goto _SET_PROJ
		}
		sfc.cad.log(3, "    lab[%d]=%v\n", lv, lab)

		// new projection
		// set \bar{P} to the closure under the projection operator of
		// \bar{P} \cup DS(S)
		for _, j := range lab {
			pf := sfc.cad.proj[lv].get(uint(j))
			sfc.cad.log(3, "  pf[%d,%d]=%v\n", lv, j, pf.P())
			sfc.add_ds_proj(lv, pf)
		}

		// rlift がバグっているので，諦める.
		if proj_num[lv] != sfc.cad.proj[lv].Len() {
			cad2, _ := NewCAD(sfc.cad.qfml, sfc.cad.g)
			for j := proj_num[lv]; j < sfc.cad.proj[lv].Len(); j++ {
				cad2.apppoly = append(cad2.apppoly, sfc.cad.proj[lv].get(uint(j)).P())
			}

			cad2.Projection(sfc.cad.palgo)
			cad2.Lift()
			return cad2.Sfc()
		}

		// rebuild CAD
		if lv > 0 {
			for j := proj_num[lv]; j < sfc.cad.proj[lv].Len(); j++ {
				sfc.cad.log(3, "  lv=%d, j=%d/%d/%d: %v\n", lv, j, proj_num[lv], sfc.cad.proj[lv].Len(), sfc.cad.proj[lv].get(uint(j)).P())
				sfc.cad.proj[lv].doProj(sfc.cad, j)
			}
		}

	_SET_PROJ:
		sfc.cad.log(3, "  @@ LV=%d: %d => %d\n", lv, proj_num[lv], sfc.cad.proj[lv].Len())
		for i, pf := range sfc.cad.proj[lv].gets() {
			pf.SetIndex(uint(i))
		}

		sfc.cad.root.rlift(sfc.cad, lv, proj_num)
		sfc.cad.Lift()
		if err := sfc.cad.root.valid(sfc.cad); err != nil {
			panic(err.Error())
		}
	}

	return nil, fmt.Errorf("??")
}

// no test...
func (sfc *CADSfc) lift_conlicts(cad *CAD) error {
	for i := 0; i < sfc.freen-1; i++ {
		for _, cpair := range sfc.cpair[i] {
			fmt.Printf("go cpair!")
			for j := 2; j < 4; j++ {
				c := cpair[j]
				if int(c.lv) < sfc.freen-1 && c.children == nil {
					// まだ持ち上げてなかったら，持ち上げる
					c.truth = t_undef
					err := c.lift(cad)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
