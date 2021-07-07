package ganrac

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

type ProjFactor interface {
	P() *Poly
	Lv() Level
	Deg() int

	Index() uint
	SetIndex(i uint)

	// 入力の論理式に含まれるなら true
	Input() bool
	SetInputT(b bool)

	// 数値手法による評価
	Sign() sign_t
	SetSign(sgn sign_t)

	// 係数の符号を返す
	evalCoeff(cad *CAD, cell *Cell, deg int) OP

	// return -1 不明 (unknown)
	//         0 sqfree でない (false)
	//         1 sqrfree である (true)
	hasMultiFctr(cad *CAD, parent *Cell) int

	// cell.lv < pf.lv に呼び出される.
	// 2次なら符号確定できるとか, 数値計算でとか.
	evalSign(cell *Cell) OP

	// McCallum Proj. 用.  well-oriented チェック
	vanishChk(cad *CAD, cell *Cell) bool

	// 表示用
	FprintProjFactor(b io.Writer, cad *CAD)

	numEval(cad *CAD) // 数値評価
}

type ProjFactors interface {
	Len() int
	gets() []ProjFactor
	get(index uint) ProjFactor

	// return -1 不明 (unknown)
	//         0 重複根をもたない (false)
	//         1 重複根を必ずもつ (true)
	hasCommonRoot(cad *CAD, parent *Cell, i, j uint) int

	// 新しい proj.factor を追加する
	//   isInput: 入力の論理式由来か
	addPoly(p *Poly, isInput bool) ProjFactor

	doProj(cad *CAD, idx int)
}

type ProjLink struct {
	op           OP // LT or GT. 符号を表す
	multiplicity []uint
	projs        []ProjFactor
}

type ProjFactorBase struct {
	p     *Poly
	index uint
	input bool // 入力の論理式に含まれるか.
	sgn   sign_t
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

func (pfb *ProjFactorBase) Sign() sign_t {
	return pfb.sgn
}

func (pfb *ProjFactorBase) SetSign(b sign_t) {
	if !pfb.input { // 入力は消せないでしょう
		pfb.sgn = b
	}
}

func (pfb *ProjFactorBase) numEval(cad *CAD) {
	if pfb.input || len(cad.u) == 0 {
		return
	}

	prec := uint(53)
	p := pfb.P().toIntv(prec).(*Poly)
	for lv := p.lv; lv >= 0; lv-- {
		qq := p.SubstIntv(cad.u[lv], lv, prec)
		if q, ok := qq.(*Interval); ok {
			if q.inf.Sign() > 0 {
				pfb.SetSign(1)
				return
			} else if q.sup.Sign() < 0 {
				pfb.SetSign(-1)
				return
			}
			return
		}
		p = qq.(*Poly)
	}
}

func (pfb *ProjFactorBase) Lv() Level {
	return pfb.p.lv
}

func (pfb *ProjFactorBase) Deg() int {
	return len(pfb.p.c) - 1
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
	sgn := 1
	fctr := cad.g.ox.Factor(q)
	cc, _ := fctr.Geti(0)
	if cc0, _ := cc.(*List).Geti(0); cc0.(RObj).Sign() < 0 {
		sgn *= -1
	}

	for i := fctr.Len() - 1; i > 0; i-- {
		fctri := fctr.getiList(i)
		poly := fctri.getiPoly(0)
		multi := uint(fctri.getiInt(1).Int64())
		if poly.Sign() < 0 {
			poly = poly.Neg().(*Poly)
			if multi%2 != 0 {
				sgn *= -1
			}
		}
		pf := cad.addPolyIrr(poly, isInput)
		pl.addPoly(pf, multi)
	}
	if sgn > 0 {
		pl.op = GT
	} else if sgn < 0 {
		pl.op = LT
	} else {
		panic("?")
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
	if p.op == LT {
		pl.op = pl.op.neg()
	}
	for i, m := range p.multiplicity {
		pl.addPoly(p.projs[i], m)
	}
}

func (cad *CAD) getU() []*Interval {
	fml := cad.fml.simplFctr(cad.g)
	_, t, f := fml.simplNum(cad.g, nil, nil)
	cad.u = make([]*Interval, len(cad.q))
	for lv := 0; lv < len(cad.q); lv++ {
		us := t.getU(f, Level(lv))
		u := newInterval(53)
		u.sup = us[len(us)-1].sup
		u.inf = us[0].inf
		cad.u[lv] = u
	}
	return cad.u
}

func (cad *CAD) Projection(algo ProjectionAlgo) (*List, error) {
	if cad.stage >= CAD_STAGE_PROJED {
		return nil, fmt.Errorf("already projected")
	}
	cad.palgo = algo
	cad.log(1, "go proj algo=%d, lv=%d\n", algo, len(cad.proj))
	tm_start := time.Now()

	// projection の準備
	cad.initProj(algo)
	cad.getU()
	for _, p := range cad.apppoly {
		cad.addPoly(p, false)
	}

	for lv := len(cad.proj) - 1; lv > 0; lv-- {

		// 数値評価.
		for i := cad.proj[lv].Len() - 1; i >= 0; i-- {
			cad.proj[lv].get(uint(i)).numEval(cad)
		}

		// sfc のために，「簡単」な論理式を前に置く
		sort.Slice(cad.proj[lv].gets(), func(i, j int) bool {
			return cad.proj[lv].get(uint(i)).P().Cmp(cad.proj[lv].get(uint(j)).P()) < 0
		})

		for i := 0; i < cad.proj[lv].Len(); i++ {
			cad.proj[lv].doProj(cad, i)
		}
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
	if 2 <= cad.g.verbose_cad {
		cad.PrintProj()
	}

	projs := NewList()
	for lv := 0; lv < len(cad.proj); lv++ {
		pp := NewList()
		projs.Append(pp)
		for i := 0; i < cad.proj[lv].Len(); i++ {
			pp.Append(cad.proj[lv].get(uint(i)).P())
		}
	}

	cad.stage = CAD_STAGE_PROJED
	cad.stat.tm[0] = time.Since(tm_start)
	return projs, nil
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

func gbHasZeros(gb *List) bool {
	if gb.Len() != 1 {
		return false
	}

	v, _ := gb.Geti(0)
	p, _ := v.(RObj)
	return p.IsNumeric()
}

func (cad *CAD) PrintProj(args ...interface{}) {
	cad.FprintProj(os.Stdout, args...)
}

func (cad *CAD) FprintProjs(b io.Writer, lv Level) {
	pj := cad.proj[lv]
	for _, pf := range pj.gets() {
		ss := ' '
		if pf.Input() {
			ss = 'i'
		} else if pf.Sign() > 0 {
			ss = '+'
		} else if pf.Sign() < 0 {
			ss = '-'
		}
		fmt.Fprintf(b, "[%d,%2d,%c,%2d] %v\n", lv, pf.Index(), ss, pf.Deg(), pf.P())
	}
}

func (pl *ProjLink) Fprint(b io.Writer) {
	switch pl.op {
	case GT:
		fmt.Fprintf(b, "+")
	case LT:
		fmt.Fprintf(b, "-")
	case EQ:
		fmt.Fprintf(b, "0")
	default:
		panic("!")
	}
	for i, pf := range pl.projs {
		fmt.Fprintf(b, " P(%d,%3d)^%d", pf.Lv(), pf.Index(), pl.multiplicity[i])
	}
	fmt.Fprintf(b, "\n")
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

func (pl *ProjLink) sgn2op(s sign_t) OP {
	if s < 0 {
		return LT
	} else if s > 0 {
		return GT
	} else {
		return EQ
	}
}

func (pl *ProjLink) evalSign(cell *Cell) OP {
	// returns pが cell 上で取りうる符号を返す
	op := pl.op

	for i := 0; i < len(pl.multiplicity); i++ {
		pf := pl.projs[i]
		if cell.lv < pf.Lv() {
			if pf.Sign() > 0 {
				continue
			} else if pf.Sign() < 0 {
				op = op.neg()
				continue
			}

			switch pf.evalSign(cell) {
			case OP_TRUE:
				return OP_TRUE
			case EQ:
				return EQ
			case GT:
				break
			case LT:
				op = op.neg()
			case GE:
				op |= EQ
			case LE:
				op = op.neg()
				op |= EQ
			case NE:
				op |= NE
			}
			continue
		}

		c := cell
		for c.lv != pf.Lv() {
			c = c.parent
		}
		s := c.signature[pf.Index()]
		if s == 0 {
			return EQ
		} else if s < 0 && pl.multiplicity[i]%2 == 1 {
			op = op.neg()
		}
	}
	return op
}
