package ganrac

// H. Hong
// An improvement of the projection operator in cylindrical algebraic decomposition
// ISSAC 1990
import (
	"fmt"
	"io"
)

type ProjFactorHH struct {
	ProjFactorBase
	coeff []*ProjLink
	psc   [][]*ProjLink // psc[i][j] = psc_j(red^i(p), red^i(p'))
}

type ProjFactorsHH struct {
	pf []ProjFactor

	// psc[p][q][k][j] = psc_j(red^k(p),q), p > q
	pscs [][][][]*ProjLink
}

func newProjFactorsHH() *ProjFactorsHH {
	pfs := new(ProjFactorsHH)
	pfs.pf = make([]ProjFactor, 0)
	pfs.pscs = make([][][][]*ProjLink, 0)
	return pfs
}

func (pfs *ProjFactorsHH) addPoly(p *Poly, isInput bool) ProjFactor {
	pf := new(ProjFactorHH)
	pf.p = p
	pf.input = isInput
	pfs.pf = append(pfs.pf, pf)
	return pf
}

func (pfs *ProjFactorsHH) gets() []ProjFactor {
	return pfs.pf
}

func (pfs *ProjFactorsHH) get(index uint) ProjFactor {
	return pfs.pf[index]
}

func (pfs *ProjFactorsHH) Len() int {
	return len(pfs.pf)
}

func (pfs *ProjFactorsHH) doProj(cad *CAD, i int) {
	pf := pfs.pf[i].(*ProjFactorHH)
	if pf.Sign() == 0 {
		pf.proj_coeff(cad)
		pf.proj_psc(cad)
	}
	// @TODO if pf.Sign() != 0 || pf.Sign() != 0 {

	r := make([][][]*ProjLink, i)
	pfs.pscs = append(pfs.pscs, r)
	for j := 0; j < i; j++ {
		pfj := pfs.get(uint(j))
		pj2 := pfj.P() // 符号が決まっている...
		pj := NewPoly(pj2.lv, len(pj2.c))
		copy(pj.c, pj2.c)

		pfs.pscs[i][j] = make([][]*ProjLink, pj.deg()+1)

		for k := pj.deg(); k > 0; k-- {
			pj.c = pj.c[:k+1]
			if !pj.c[k].IsZero() {
				pfs.pscs[i][j][k] = pf.pscs(cad, pj, pf.P())
			}
		}
	}
}

func (pf *ProjFactorHH) evalSign(cell *Cell) OP {
	d := pf.Deg()
	if d < 2 {
		return OP_TRUE
	}
	for d > 2 {
		if (pf.coeff[d].evalSign(cell) & EQ) == 0 {
			return OP_TRUE
		}
		d--
	}
	cs := pf.coeff[d].evalSign(cell)
	if cs&EQ != 0 {
		return OP_TRUE
	}

	// discriminant の -1倍
	ds := pf.psc[2][0].evalSign(cell)
	if ds == GT {
		return cs
	} else if ds == GE {
		return cs | EQ
	}
	return OP_TRUE
}

func (pf *ProjFactorHH) proj_coeff(cad *CAD) {
	pf.coeff = make([]*ProjLink, len(pf.p.c))
	gb := NewList()
	vars := NewList()
	bl := make([]bool, pf.p.lv)

	for i := len(pf.p.c) - 1; i >= 0; i-- {
		c := pf.p.c[i]
		if c.IsNumeric() {
			pf.coeff[i] = cad.get_projlink_num(c.Sign())
			if !c.IsZero() {
				return
			}
		} else {
			if gb.Len() > 0 {
				for lv, b := range bl {
					if !b && c.(*Poly).hasVar(Level(lv)) {
						bl[lv] = true
						vars.Append(NewPolyVar(Level(lv)))
					}
				}
				r, neg := cad.g.ox.Reduce(c.(*Poly), gb, vars, 0)
				if neg {
					c = r.Neg()
				} else {
					c = r
				}
				if c.IsNumeric() {
					if !c.IsZero() {
						return
					}
					continue
				}
			}
			pf.coeff[i] = cad.addProjRObj(c)
			for lv, b := range bl {
				if !b && c.(*Poly).hasVar(Level(lv)) {
					bl[lv] = true
					vars.Append(NewPolyVar(Level(lv)))
				}
			}
			gb.Append(c)
			gb = cad.g.ox.GB(gb, vars, 0)
		}
	}
}

func (pf *ProjFactorHH) pscs(cad *CAD, p, q *Poly) []*ProjLink {
	// return [psc_j(p,q) for j=0, ..., min(deg(p),deg(q)]
	d := p.deg()
	if dq := q.deg(); dq < d {
		d = dq
	}
	ret := make([]*ProjLink, d)
	for j := 0; j < d; j++ {
		psc := cad.g.ox.Psc(p, q, p.lv, int32(j))
		cad.stat.psc++
		ret[j] = cad.addProjRObj(psc)
		for _, pj := range ret[j].projs {
			if err := pj.P().valid(); err != nil {
				panic(fmt.Sprintf("P=%v, err=%v\n", pj.P(), err))
			}
		}
	}

	return ret
}

func (pf *ProjFactorHH) proj_psc(cad *CAD) {
	deg := len(pf.p.c) - 1
	if deg == 1 { // linear
		return
	}
	pf.psc = make([][]*ProjLink, deg+1)
	p := NewPoly(pf.p.lv, deg+1)
	copy(p.c, pf.p.c)
	pd := p.diff(p.lv).(*Poly)
	for i := deg; i > 1; i-- {
		p.c = p.c[:i+1]
		pd.c = pd.c[:i]
		if p.c[i].IsZero() {
			continue
		}

		pf.psc[i] = pf.pscs(cad, p, pd)

		c := pf.p.c[i]
		if c.IsNumeric() {
			pf.coeff[i] = cad.get_projlink_num(c.Sign())
			if !c.IsZero() {
				return
			}
		}
	}
}

func (pf *ProjFactorHH) evalCoeff(cad *CAD, cell *Cell, deg int) OP {
	if pf.coeff == nil || pf.coeff[deg] == nil {
		return OP_TRUE
	}
	return pf.coeff[deg].evalSign(cell)
}

func (pf *ProjFactorHH) hasMultiFctr(cad *CAD, cell *Cell) int {
	// return -1 重複根をもつかも (unknown)
	//         0 重複根をもたない (false)
	//         1 重複根を必ずもつ (true)

	for d := pf.P().deg(); d > 1; d-- {
		if pf.coeff[d].evalSign(cell) != 0 {
			switch pf.psc[d][0].evalSign(cell) {
			case EQ:
				return PF_EVAL_YES
			case NE, GT, LT:
				return PF_EVAL_NO
			default:
				return PF_EVAL_UNKNOWN
			}
		}
	}
	return PF_EVAL_NO
}

func (pfs *ProjFactorsHH) hasCommonRoot(cad *CAD, c *Cell, i, j uint) int {
	// return -1 重複根をもつかも (unknown)
	//         0 重複根をもたない (false)
	//         1 重複根を必ずもつ (true)

	// 射影因子の符号で，共通因子を持つか調べる.
	// true なら，もつ可能性がある.
	if i < j {
		i, j = j, i
	}

	pi := pfs.pf[i]
	if (pi.evalCoeff(cad, c, pi.Deg()) & EQ) != 0 {
		return PF_EVAL_UNKNOWN
	}

	pj := pfs.pf[j]
	d := pj.Deg()
	for d > 0 {
		if (pj.evalCoeff(cad, c, d) & EQ) == 0 {
			break
		}
		d--
	}

	if (pfs.pscs[i][j][d][0].evalSign(c) & EQ) != 0 {
		return PF_EVAL_YES
	} else {
		return PF_EVAL_NO
	}
}

func (pf *ProjFactorHH) FprintProjFactor(b io.Writer, cad *CAD) {
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
	for k := len(pf.psc) - 1; k >= 0; k-- {
		for i := len(pf.psc[k]) - 1; i >= 0; i-- {
			fmt.Fprintf(b, "psc[%d,%d]=", k, i)
			pf.psc[k][i].Fprint(b)
		}
	}
}

func (pf *ProjFactorHH) vanishChk(cad *CAD, cell *Cell) bool {
	return true
}
