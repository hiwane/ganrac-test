package ganrac

import (
	"fmt"
	"io"
	"os"
	"sort"
)

type ProjFactorBase struct {
	p     *Poly
	index uint
	input bool // 入力の論理式に含まれるか.
}

type ProjFactorMC struct {
	ProjFactorBase
	coeff   []*ProjLink
	discrim *ProjLink
}

type ProjFactorsMC struct {
	pf        []ProjFactor
	resultant [][]*ProjLink
}

func (pfb *ProjFactorBase) P() *Poly {
	return pfb.p
}

func (pfb *ProjFactorBase) Index() uint {
	return pfb.index
}

func (pfb *ProjFactorBase) SetIndex(i uint) {
	pfb.index = i
}

func (pfb *ProjFactorBase) Input() bool {
	return pfb.input
}

func (pfb *ProjFactorBase) SetInputT(b bool) {
	pfb.input = (pfb.input || b)
}

func (pfb *ProjFactorBase) Lv() Level {
	return pfb.p.lv
}

func (pfb *ProjFactorBase) Deg() int {
	return len(pfb.p.c) - 1
}

func (pfs *ProjFactorsMC) gets() []ProjFactor {
	return pfs.pf
}

func (pfs *ProjFactorsMC) get(index uint) ProjFactor {
	return pfs.pf[index]
}

func (pfs *ProjFactorsMC) Len() int {
	return len(pfs.pf)
}

func newProjFactorsMC() *ProjFactorsMC {
	pfs := new(ProjFactorsMC)
	pfs.pf = make([]ProjFactor, 0)
	return pfs
}

// cylindrical algebraic decomposition

func (cad *CAD) addPolyIrr(q *Poly, isInput bool) ProjFactor {
	// assume: lc(q) > 0, irreducible
	if q.Sign() < 0 {
		panic("invalid")
	}
	proj_factors := cad.proj[q.lv]
	for _, pf := range proj_factors.gets() {
		if pf.P().Equals(q) {
			pf.SetInputT(isInput)
			return pf
		}
	}
	return proj_factors.addPoly(q, isInput)
}

func (cad *CAD) addProjRObj(q RObj) *ProjLink {
	switch cz := q.(type) {
	case *Poly:
		return cad.addPoly(cz, false)
	case NObj:
		return cad.get_projlink_num(cz.Sign())
	default:
		fmt.Printf("cz=%v\n", cz)
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
		multi := uint(fctri.getiInt(1).Int64())
		if poly.Sign() < 0 {
			poly = poly.Neg().(*Poly)
			if multi%2 != 0 {
				pl.sgn *= -1
			}
		}
		pf := cad.addPolyIrr(poly, isInput)
		pl.addPoly(pf, multi)
	}

	return pl
}

func newProjLink() *ProjLink {
	pl := new(ProjLink)
	pl.multiplicity = make([]uint, 0)
	pl.projs = make([]ProjFactor, 0)
	return pl
}

func (pl *ProjLink) addPoly(p ProjFactor, r uint) {
	pl.projs = append(pl.projs, p)
	pl.multiplicity = append(pl.multiplicity, r)
}

func (pl *ProjLink) merge(p *ProjLink) {
	pl.sgn *= p.sgn
	for i := 0; i < len(p.multiplicity); i++ {
		pl.addPoly(p.projs[i], p.multiplicity[i])
	}
}

func (cad *CAD) Projection(algo ProjectionAlgo) error {
	fmt.Printf("go proj algo=%d, lv=%d\n", algo, len(cad.proj))
	for lv := len(cad.proj) - 1; lv > 0; lv-- {

		sort.Slice(cad.proj[lv].gets(), func(i, j int) bool {
			return cad.proj[lv].get(uint(i)).P().Cmp(cad.proj[lv].get(uint(j)).P()) < 0
		})
		proj_mcallum(cad, Level(lv))
	}
	{
		lv := 0
		sort.Slice(cad.proj[lv].gets(), func(i, j int) bool {
			return cad.proj[lv].get(uint(i)).P().Cmp(cad.proj[lv].get(uint(j)).P()) < 0
		})
	}

	// インデックスをつける
	for lv := len(cad.proj) - 1; lv >= 0; lv-- {
		pj := cad.proj[lv]
		for i, pf := range pj.gets() {
			pf.SetIndex(uint(i))
		}
	}

	// 最下層の coeff だけ設定しておく. @AAA
	// coef := make([]*ProjLink, 1)
	// coef[0] = cad.get_projlink_num(1)
	// for _, pf := range cad.proj[0].gets() {
	// 	pf.coeff = coef
	// }
	cad.stage = 1

	cad.PrintProj()

	return nil
}

func proj_mcallum(cad *CAD, lv Level) {
	pj := cad.proj[lv].(*ProjFactorsMC)
	for _, _pf := range pj.gets() {
		pf := _pf.(*ProjFactorMC)
		pf.proj_coeff(cad)
		pf.proj_discrim(cad)
	}

	pj.resultant = make([][]*ProjLink, pj.Len())
	for i := 0; i < len(pj.pf); i++ {
		pj.resultant[i] = make([]*ProjLink, i)
		for j := 0; j < i; j++ {
			dd := cad.g.ox.Resultant(pj.get(uint(i)).P(), pj.get(uint(j)).P(), lv)
			cad.stat.resultant++
			pj.resultant[i][j] = cad.addProjRObj(dd)
		}
	}
}

func (cad *CAD) get_projlink_num(sign int) *ProjLink {
	// 定数なときの，符号を指定して，対応する pl を返す
	if sign > 0 {
		return cad.pl4const[1]
	} else if sign < 0 {
		return cad.pl4const[2]
	} else {
		return cad.pl4const[0]
	}
}

func (pf *ProjFactorMC) evalSign(cell *Cell) (sign_t, bool) {
	if pf.Deg() != 2 {
		return 0, false
	}
	ds, dok := pf.discrim.evalSign(cell)
	cs, cok := pf.coeff[2].evalSign(cell)
	if dok && cok && ds < 0 && cs != 0 {
		return cs, true
	}
	return 0, false
}

func (pf *ProjFactorMC) proj_coeff(cad *CAD) {
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

func (pf *ProjFactorMC) proj_discrim(cad *CAD) {
	dd := cad.g.ox.Discrim(pf.p, pf.p.lv)
	cad.stat.discriminant++
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

func (pf *ProjFactorMC) evalCoeff(cad *CAD, cell *Cell, deg int) (sign_t, bool) {
	if pf.coeff[deg] == nil {
		return 0, false
	}
	return pf.coeff[deg].evalSign(cell)
}

func (pf *ProjFactorMC) hasMultiFctr(cad *CAD, cell *Cell) int {
	// return -1 重複根をもつかも (unknown)
	//         0 重複根をもたない (false)
	//         1 重複根を必ずもつ (true)
	if pf.discrim == nil {
		pf.FprintProjFactor(os.Stdout, cad)
	}
	s, d := pf.discrim.evalSign(cell)
	if !d {
		return -1
	} else if s == 0 {
		return 1
	} else {
		return 0
	}
}

func (pfs *ProjFactorsMC) addPoly(p *Poly, isInput bool) ProjFactor {
	pf := new(ProjFactorMC)
	pf.p = p
	pf.input = isInput
	pfs.pf = append(pfs.pf, pf)
	return pf
}

func (pfs *ProjFactorsMC) hasCommonRoot(cad *CAD, c *Cell, i, j uint) int {
	// return -1 重複根をもつかも (unknown)
	//         0 重複根をもたない (false)
	//         1 重複根を必ずもつ (true)

	// 射影因子の符号で，共通因子を持つか調べる.
	// true なら，もつ可能性がある.
	for _, pf := range []ProjFactor{pfs.pf[i], pfs.pf[j]} {
		// 次数が落ちていると，共通根を持たなくても終結式が 0 になる
		s, d := pf.evalCoeff(cad, c, pf.Deg())
		if !d || s == 0 {
			return -1
		}
	}

	var pl *ProjLink
	if i < j {
		pl = pfs.resultant[j][i]
	} else {
		pl = pfs.resultant[i][j]
	}
	s, d := pl.evalSign(c)
	if !d {
		return -1
	} else if s == 0 {
		return 1
	} else {
		return 0
	}
}

func (cad *CAD) PrintProj(args ...interface{}) {
	cad.FprintProj(os.Stdout, args...)
}

func (cad *CAD) FprintProjs(b io.Writer, lv Level) {
	pj := cad.proj[lv]
	for i, pf := range pj.gets() {
		ss := ' '
		if pf.Input() {
			ss = 'i'
		}
		fmt.Fprintf(b, "[%d,%2d,%c,%2d] %v\n", lv, i, ss, pf.Deg(), pf.P())
	}
}

func (pl *ProjLink) Fprint(b io.Writer) {
	if pl.sgn > 0 {
		fmt.Fprintf(b, "+")
	} else if pl.sgn < 0 {
		fmt.Fprintf(b, "-")
	} else if pl.sgn == 0 {
		fmt.Fprintf(b, "0")
	}
	for i, pf := range pl.projs {
		fmt.Fprintf(b, " <%d,%3d>^%d", pf.Lv(), pf.Index(), pl.multiplicity[i])
	}
	fmt.Fprintf(b, "\n")
}

func (pf *ProjFactorMC) FprintProjFactor(b io.Writer, cad *CAD) {
	ss := ' '
	if pf.input {
		ss = 'i'
	}
	fmt.Fprintf(b, "[%d,%2d,%c,%2d] %v\n", pf.Lv(), pf.Index(), ss, pf.Deg(), pf.P())
	for i := len(pf.coeff) - 1; i >= 0; i-- {
		if pf.coeff[i] != nil {
			fmt.Fprintf(b, "coef[%d]=", i)
			pf.coeff[i].Fprint(b)
		}
	}
	fmt.Fprintf(b, "discrim=")
	pf.discrim.Fprint(b)
}

func (cad *CAD) FprintInput(b io.Writer, args ...interface{}) error {
	fmt.Fprintf(b, "input formula\n")
	cad.fprintInput(b, cad.fml)
	return nil
}

func (cad *CAD) fprintInput(b io.Writer, fml Fof) {
	switch ff := fml.(type) {
	case *AtomProj:
		fmt.Fprintf(b, "\n%v\n ...   ", ff)
		ff.pl.Fprint(b)
	case *FmlAnd:
		for _, f := range ff.fml {
			cad.fprintInput(b, f)
		}
	case *FmlOr:
		for _, f := range ff.fml {
			cad.fprintInput(b, f)
		}
	}
}

func (cad *CAD) FprintProj(b io.Writer, args ...interface{}) error {
	idx := 0
	lv := Level(-1)
	if len(args) > idx {
		if ii, ok := args[idx].(*Int); ok {
			lvv := ii.Int64()
			idx++
			if lvv >= int64(len(cad.proj)) {
				return fmt.Errorf("invalid argument [level=%d]", lvv)
			}
			lv = Level(lvv)
		}
	}
	index := -1
	if len(args) > idx {
		if ii, ok := args[idx].(*Int); ok {
			if ii.Int64() >= int64(cad.proj[lv].Len()) {
				return fmt.Errorf("invalid argument [index=%d]", index)
			}
			index = int(ii.Int64())
			idx++
		}
	}

	if lv >= 0 {
		if index < 0 {
			cad.FprintProjs(b, lv)
		} else {
			cad.proj[lv].get(uint(index)).FprintProjFactor(b, cad)
		}
	} else {
		for lv = Level(len(cad.proj) - 1); lv >= 0; lv-- {
			cad.FprintProjs(b, lv)
		}
	}
	return nil
}
