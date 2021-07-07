package ganrac

// McCallum Projection
// S. McCallum.
// An improved projection operator for cylindrical algebraic decomposition
// In Quantier Elimination and Cylindrical Algebraic Decomposition (1998)
import (
	"fmt"
	"io"
	"os"
)

type ProjFactorMC struct {
	ProjFactorBase
	coeff   []*ProjLink
	discrim *ProjLink
}

type ProjFactorsMC struct {
	pf []ProjFactor

	// resultant[i][j] = res(pf[i], pf[j]) where i > j
	resultant [][]*ProjLink
}

func newProjFactorsMC() *ProjFactorsMC {
	pfs := new(ProjFactorsMC)
	pfs.pf = make([]ProjFactor, 0)
	pfs.resultant = make([][]*ProjLink, 0)
	return pfs
}

func (pfs *ProjFactorsMC) addPoly(p *Poly, isInput bool) ProjFactor {
	pf := new(ProjFactorMC)
	pf.p = p
	pf.input = isInput
	pfs.pf = append(pfs.pf, pf)
	return pf
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

func (pfs *ProjFactorsMC) doProj(cad *CAD, i int) {
	pf := pfs.pf[i].(*ProjFactorMC)
	if pf.Sign() == 0 {
		pf.proj_coeff(cad)
		pf.proj_discrim(cad)
	}

	r := make([]*ProjLink, i)
	pfs.resultant = append(pfs.resultant, r)
	for j := 0; j < i; j++ {
		pg := pfs.get(uint(j))
		if pf.Sign() != 0 || pg.Sign() != 0 {
			// 交わりません.
			pfs.resultant[i][j] = cad.pl4const[1]
			continue
		}

		pj := pg.P()
		dd := cad.g.ox.Resultant(pf.p, pj, pf.p.lv)
		cad.stat.resultant++
		pfs.resultant[i][j] = cad.addProjRObj(dd)
	}
}

func (pf *ProjFactorMC) evalSign(cell *Cell) OP {
	if pf.Deg() != 2 {
		return OP_TRUE
	}
	cs := pf.coeff[2].evalSign(cell)
	if (cs & EQ) == 0 {
		ds := pf.discrim.evalSign(cell)
		if ds == LT {
			return cs
		} else if ds == LE {
			return cs | EQ
		}
	}
	return OP_TRUE
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

func (pf *ProjFactorMC) evalCoeff(cad *CAD, cell *Cell, deg int) OP {
	// fmt.Printf("deg=%d, coef.len=%d, %d\n", deg, len(pf.coeff), pf.Sign())
	// pf.FprintProjFactor(os.Stdout, cad)
	if pf.coeff[deg] == nil {
		return OP_TRUE
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

	if (pf.evalCoeff(cad, cell, pf.Deg()) & EQ) != 0 {
		return PF_EVAL_UNKNOWN
	}
	switch pf.discrim.evalSign(cell) {
	case EQ:
		return PF_EVAL_YES
	case NE, GT, LT:
		return PF_EVAL_NO
	default:
		return PF_EVAL_UNKNOWN
	}
}

func (pfs *ProjFactorsMC) hasCommonRoot(cad *CAD, c *Cell, i, j uint) int {
	// return -1 重複根をもつかも (unknown)
	//         0 重複根をもたない (false)
	//         1 重複根を必ずもつ (true)

	// 射影因子の符号で，共通因子を持つか調べる.
	// true なら，もつ可能性がある.
	n := 0
	for _, pf := range []ProjFactor{pfs.pf[i], pfs.pf[j]} {
		// 次数が落ちていると，共通根を持たなくても終結式が 0 になる
		if pf.(*ProjFactorMC).coeff == nil {
			// rlift 時には proj が構成されていない場合がある
			return PF_EVAL_UNKNOWN
		}
		if (pf.evalCoeff(cad, c, pf.Deg()) & EQ) != 0 {
			n++
		}
	}
	if n == 2 {
		return PF_EVAL_UNKNOWN
	}

	var pl *ProjLink
	if i < j {
		pl = pfs.resultant[j][i]
	} else {
		pl = pfs.resultant[i][j]
	}
	switch pl.evalSign(c) {
	case EQ:
		return PF_EVAL_YES
	case NE, GT, LT:
		return PF_EVAL_NO
	default:
		return PF_EVAL_UNKNOWN
	}
}

func (pf *ProjFactorMC) FprintProjFactor(b io.Writer, cad *CAD) {
	ss := ' '
	if pf.input {
		ss = 'i'
	}
	lv := pf.Lv()
	idx := pf.Index()
	fmt.Fprintf(b, "[%d,%2d,%c,%2d] %v\n", lv, idx, ss, pf.Deg(), pf.P())
	for i := len(pf.coeff) - 1; i >= 0; i-- {
		if pf.coeff[i] != nil {
			fmt.Fprintf(b, "coef[%d]=", i)
			pf.coeff[i].Fprint(b)
		}
	}
	if pf.discrim != nil {
		fmt.Fprintf(b, "discrim=")
		pf.discrim.Fprint(b)
	}

	if lv > 0 {
		pfs := cad.proj[lv].(*ProjFactorsMC)
		for i := uint(0); i < idx; i++ {
			fmt.Fprintf(b, "res[%2d]=", i)
			pfs.resultant[idx][i].Fprint(b)
		}
		for i := int(idx) + 1; i < len(pfs.resultant); i++ {
			fmt.Fprintf(b, "res[%2d]=", i)
			pfs.resultant[i][idx].Fprint(b)
		}
	}
}
