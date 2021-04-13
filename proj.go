package ganrac

import (
	"fmt"
)

// cylindrical algebraic decomposition

func (cad *CAD) addPolyIrr(q *Poly, isInput bool) *ProjFactor {
	// assume: lc(q) > 0, irreducible
	proj_factors := cad.proj[q.lv]
	for _, pf := range proj_factors.pf {
		if pf.p.Equals(q) {
			if isInput {
				pf.input = true
			}
			return pf
		}
	}
	pf := new(ProjFactor)
	pf.p = q
	pf.input = isInput
	proj_factors.pf = append(proj_factors.pf, pf)
	return pf
}

func (cad *CAD) addProjRObj(q RObj) *ProjLink {
	switch cz := q.(type) {
	case *Poly:
		return cad.addPoly(cz, false)
	case NObj:
		return cad.get_projlink_num(cz.Sign())
	default:
		panic("unknown")
	}
}

func (cad *CAD) addPoly(q *Poly, isInput bool) *ProjLink {
	pl := newProjLink()
	pl.sgn = 1
	fctr := cad.g.ox.Factor(q)
	cc, _ := fctr.Geti(0)
	if cc0, _ := cc.(*List).Geti(0); cc0.(RObj).Sign() < 0 {
		pl.sgn *= -1
	}

	for i := fctr.Len() - 1; i > 0; i-- {
		fctri := fctr.getiList(i)
		poly := fctri.getiPoly(0)
		if poly.Sign() < 0 {
			poly = poly.Neg().(*Poly)
			pl.sgn *= -1
		}
		pf := cad.addPolyIrr(poly, isInput)
		pl.addPoly(pf, uint(fctri.getiInt(1).Int64()))
	}

	return pl
}

func newProjLink() *ProjLink {
	pl := new(ProjLink)
	pl.multiplicity = make([]uint, 0)
	pl.projs.pf = make([]*ProjFactor, 0)
	return pl
}

func (pl *ProjLink) addPoly(p *ProjFactor, r uint) {
	pl.projs.pf = append(pl.projs.pf, p)
	pl.multiplicity = append(pl.multiplicity, r)
}

func (pl *ProjLink) merge(p *ProjLink) {
	pl.sgn *= p.sgn
	for i := 0; i < len(p.multiplicity); i++ {
		pl.addPoly(p.projs.pf[i], p.multiplicity[i])
	}
}

func (cad *CAD) Projection(algo ProjectionAlgo) error {
	fmt.Printf("go proj algo=%d, lv=%d\n", algo, len(cad.proj))
	for lv := len(cad.proj) - 1; lv > 0; lv-- {
		proj_mcallum(cad, Level(lv))
	}

	// sort してインデックスをつける
	// @TODO ソートする.
	for lv := len(cad.proj) - 1; lv >= 0; lv-- {
		pj := cad.proj[lv]
		for i, pf := range pj.pf {
			pf.index = uint(i)
			ss := ' '
			if pf.input {
				ss = 'i'
			}
			fmt.Printf("[%d,%2d,%c,%d] %v\n", lv, i, ss, pf.p.lv, pf.p)
		}
	}

	// 最下層の coeff だけ設定しておく.
	coef := make([]*ProjLink, 1)
	coef[0] = cad.get_projlink_num(1)
	for _, pf := range cad.proj[0].pf {
		pf.coeff = coef
	}

	return nil
}

func proj_mcallum(cad *CAD, lv Level) {
	pj := cad.proj[lv]
	for _, pf := range pj.pf {
		proj_mcallum_coeff(cad, pf)
		proj_mcallum_discrim(cad, pf)
	}

	pj.resultant = make([][]*ProjLink, len(pj.pf))
	for i := 0; i < len(pj.pf); i++ {
		pj.resultant[i] = make([]*ProjLink, i)
		for j := 0; j < i; j++ {
			dd := cad.g.ox.Discrim(pj.pf[i].p, pj.pf[j].p.lv)
			pj.resultant[i][j] = cad.addProjRObj(dd)
		}
	}
}

func (cad *CAD) get_projlink_num(sign int) *ProjLink {
	if sign > 0 {
		return cad.pl4const[1]
	} else if sign < 0 {
		return cad.pl4const[2]
	} else {
		return cad.pl4const[0]
	}
}

func proj_mcallum_coeff(cad *CAD, pf *ProjFactor) {
	pf.coeff = make([]*ProjLink, len(pf.p.c))
	for i := len(pf.p.c) - 1; i >= 0; i-- {
		c := pf.p.c[i]
		if c.IsNumeric() {
			pf.coeff[i] = cad.get_projlink_num(c.Sign())
			if !c.IsZero() {
				return
			}
		} else {
			pf.coeff[i] = cad.addProjRObj(c)
		}
	}
	// GB で vanish チェック
	// gb := cad.g.ox.GB(list, uint(len(cad.proj)))
	// if !gbHasZeros(gb) {
	// 	// 主係数のみ... だったはず. @TODO
	// 	j := len(pf.p.c) - 1
	// 	cz := pf.p.c[j].(*Poly)
	// 	pf.coeff[j] = cad.addPoly(cz, false)
	// }
}

func proj_mcallum_discrim(cad *CAD, pf *ProjFactor) {
	dd := cad.g.ox.Discrim(pf.p, pf.p.lv)
	pf.discrim = cad.addProjRObj(dd)
}

func gbHasZeros(gb *List) bool {
	if gb.Len() != 1 {
		return false
	}

	v, _ := gb.Geti(0)
	p, _ := v.(RObj)
	return p.IsNumeric()
}

func (pfs *ProjFactors) hasCommonRoot(c *Cell, i, j uint) bool {
	// 射影因子の符号で，共通因子を持つか調べる.
	// true なら，もつ可能性がある.
	// @TODO 主係数がおちたときに，共通因子もたなくてもゼロ
	if pfs.resultant == nil {
		return false
	}

	if i < j {
		b, _ := pfs.resultant[j][i].evalSign(c)
		return b == 0
	} else {
		b, _ := pfs.resultant[i][j].evalSign(c)
		return b == 0
	}
}