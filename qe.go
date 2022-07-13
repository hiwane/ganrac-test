package ganrac

// Automated Natural Language Geometry Math Problem Solving by Real Quantier Elimination
// Hidenao Iwane, Takuya Matsuzaki, Noriko Arai, Hirokazu Anai
// ADG2014

import (
	"fmt"
	"sort"
)

type algo_t int64

const (
	QEALGO_VSLIN  = 0x0001
	QEALGO_VSQUAD = 0x0002

	QEALGO_EQLIN  = 0x0010
	QEALGO_EQQUAD = 0x0020

	QEALGO_NEQ = 0x0100 // 非等式制約QE

	QEALGO_SMPL_EVEN = 0x100000000
	QEALGO_SMPL_HOMO = 0x200000000
	QEALGO_SMPL_TRAN = 0x400000000
	QEALGO_SMPL_ROTA = 0x800000000
)

type QEopt struct {
	varn      Level
	log_level int
	Algo      algo_t
	g         *Ganrac
	seqno     int
	assert    bool
}

type qeCond struct {
	neccon, sufcon Fof
	dnf            bool
	depth          int
}

func NewQEopt() *QEopt {
	o := new(QEopt)
	o.Algo = -1
	o.assert = true
	return o
}

func getQEoptStr(algo int64) string {
	switch algo {
	case QEALGO_EQQUAD:
		return "eqquad"
	case QEALGO_EQLIN:
		return "eqlin"
	case QEALGO_VSQUAD:
		return "vsquad"
	case QEALGO_VSLIN:
		return "vslin"
	case QEALGO_NEQ:
		return "neq"
	case QEALGO_SMPL_EVEN:
		return "smpleven"
	case QEALGO_SMPL_HOMO:
		return "smplhomo"
	case QEALGO_SMPL_TRAN:
		return "smpltran"
	case QEALGO_SMPL_ROTA:
		return "smplrot"
	}
	return ""
}

func (qeopt *QEopt) SetAlgo(algo algo_t, v bool) {
	if v {
		qeopt.Algo |= algo
	} else {
		qeopt.Algo &= ^algo
	}
}

func (qeopt *QEopt) log(cond qeCond, level int, label, fmt string, args ...interface{}) {
	if qeopt.log_level < level {
		return
	}

	v := make([]interface{}, len(args)+3)
	v[0] = label
	v[1] = qeopt.seqno
	v[2] = cond.depth
	copy(v[3:], args)
	qeopt.g.log(level, "%5s[%3d,%3d] "+fmt, v...)
}

func (qeopt *QEopt) num_var(f Fof) int {
	b := make([]bool, qeopt.varn)
	f.Indets(b)
	m := 0
	for _, b := range b {
		if b {
			m++
		}
	}
	return m
}

func (qeopt *QEopt) new_var() Level {
	v := qeopt.varn
	qeopt.varn += 1
	return v
}

func (qeopt *QEopt) fmlcmp(f1, f2 Fof) bool {
	switch g1 := f1.(type) {
	case FofQ:
		switch g2 := f2.(type) {
		case FofQ:
			m1 := qeopt.num_var(g1)
			m2 := qeopt.num_var(g2)
			return m1 < m2
		default:
			return false
		}
	case FofAO:
		switch g2 := f2.(type) {
		case FofAO:
			if g1.IsQff() && !g2.IsQff() {
				return true
			} else if !g1.IsQff() && g2.IsQff() {
				return false
			}

			m1 := qeopt.num_var(g1)
			m2 := qeopt.num_var(g2)
			if m1 != m2 {
				return m1 < m2
			}

			m1 = g1.numAtom()
			m2 = g2.numAtom()
			if m1 != m2 {
				return m1 < m2
			}

			m1 = g1.sotd()
			m2 = g1.sotd()
			return m1 <= m2

		default:
			return false
		}
	default: // atom
		switch g2 := f2.(type) {
		case FofQ:
			return true
		case FofAO:
			return true
		default:
			m1 := qeopt.num_var(g1)
			m2 := qeopt.num_var(g2)
			if m1 != m2 {
				return m1 < m2
			}

			m1 = g1.sotd()
			m2 = g1.sotd()
			return m1 <= m2
		}
	}
}

func (qeopt *QEopt) qe_init(g *Ganrac, fof Fof) {
	qeopt.varn = fof.maxVar() + 1
	qeopt.g = g
	if qeopt.Algo == 0 {
		qeopt.Algo = -1
	}
}

func (cond *qeCond) qecond_init() {
	cond.neccon = trueObj
	cond.sufcon = falseObj
}

func (g *Ganrac) QE(fof Fof, qeopt *QEopt) Fof {
	var cond qeCond
	qeopt.qe_init(g, fof)
	cond.qecond_init()
	return qeopt.qe(fof, cond)
}

func (qeopt QEopt) qe(fof Fof, cond qeCond) Fof {
	cond.depth++
	qeopt.seqno++
	qeopt.log(cond, 2, "qe", "%v\n", fof)
	for {
		fof = fof.nonPrenex()
		qeopt.log(cond, 2, "qe", "1:%v\n", fof)
		switch fq := fof.(type) {
		case FofQ:
			if fof.isPrenex() {
				fof = qeopt.qe_prenex(fq, cond)
			} else {
				fof = qeopt.qe_nonpreq(fq, cond)
			}
		case FofAO:
			if fof.IsQff() {
				return qeopt.simplify(fof, cond)
			}
			fof = qeopt.qe_andor(fq, cond)
		default: // atom
			return fof
		}
		if fof.IsQff() {
			return fof
		}
	}
}

func (qeopt QEopt) simplify(qff Fof, cond qeCond) Fof {
	return qeopt.g.simplFof(qff, cond.neccon, cond.sufcon)
}

/*
 * fof: prenex formula
 */
func (qeopt QEopt) qe_prenex(fof FofQ, cond qeCond) Fof {
	qeopt.log(cond, 2, "qepr", "%v\n", fof)
	// exists-or, forall-and は分解できる.
	fofq := fof
	if err := fofq.valid(); err != nil || !fofq.isPrenex() {
		panic(fmt.Sprintf("err=%v, prenex=%v", err, fofq.isPrenex()))
	}

	fs := make([]FofQ, 1)
	fs[0] = fofq
	for {
		fml := fofq.Fml()
		if fml.IsQuantifier() {
			fofq = fml.(FofQ)
			fs = append(fs, fofq)
		} else {
			if ao, ok := fml.(FofAO); ok {
				if fofq.isForAll() == ao.isAnd() {
					// 分解できる.
					//      ex([x], f1 or f2 or f3)
					// <==> ex([x], f1) or ex([x], f2) or ex([x], f3)
					var cond2 qeCond = cond
					cond2.dnf = !ao.isAnd()
					cond2.depth = cond.depth + 1

					ret := make([]Fof, len(ao.Fmls()))
					for i, f := range ao.Fmls() {
						ret[i] = fofq.gen(fofq.Qs(), f)
					}

					ao = ao.gen(ret).(FofAO)
					fmlq := qeopt.qe_andor(ao, cond2)
					for i := len(fs) - 2; i >= 0; i-- {
						fmlq = fs[i].gen(fs[i].Qs(), fmlq)
					}

					return qeopt.qe(fmlq, cond)
				} // else
			} // else Atom
			break
		}
	}

	// もうがんばるしかない状態.
	return qeopt.qe_prenex_main(fofq, cond)
}

func (qeopt QEopt) reconstruct(fqs []FofQ, ff Fof, cond qeCond) Fof {
	for i := len(fqs) - 1; i >= 0; i-- {
		ff = fqs[i].gen(fqs[i].Qs(), ff)
	}
	return qeopt.qe(ff, cond)

}

func (qeopt QEopt) qe_simpl(fof FofQ, cond qeCond) Fof {

	// 偶論理式
	if (qeopt.Algo & QEALGO_SMPL_EVEN) != 0 {
		if ff := qeopt.qe_evenq(fof, cond); ff != nil {
			return ff
		}
	}

	// 斉次論理式: homogeneous formula
	if (qeopt.Algo & QEALGO_SMPL_HOMO) != 0 {
		if ff := qeopt.qe_homo(fof, cond); ff != nil {
			return ff
		}
	}

	return nil
}

func (qeopt QEopt) qe_prenex_main(prenex_formula FofQ, cond qeCond) Fof {
	fof := prenex_formula

	// quantifier の一番外側を処理する.
	fof = prenex_formula
	fqs := make([]FofQ, 1, 10)
	fqs[0] = fof
	for {
		if fq, ok := fof.Fml().(FofQ); ok {
			fqs = append(fqs, fq)
			fof = fq
		} else {
			break
		}
	}

	////////////////////////////////
	// 複数等式制約の GB による簡単化
	// @see speeding up CAD by GB.
	////////////////////////////////

	////////////////////////////////
	// SDC
	// 分解後に All->DNF/Ex->CNF になるので,
	// quantifier がひとつの場合のみに限定してみる
	////////////////////////////////

	////////////////////////////////
	// Hong93
	// 線形か2次の等式制約が含まれる場合.
	////////////////////////////////
	if qeopt.Algo&(QEALGO_EQLIN|QEALGO_EQQUAD) != 0 {
		if ff := qeopt.qe_quadeq(fof, cond); ff != nil {
			ff = qeopt.reconstruct(fqs, ff, cond)
			ff = qeopt.simplify(ff, cond)
			qeopt.log(cond, 2, "eqret", "%v\n", fof)
			return ff
		}
	}

	////////////////////////////////
	// VS を適用できるか.
	////////////////////////////////
	if (qeopt.Algo & QEALGO_VSLIN) != 0 {
		if ff := qeopt.qe_vslin(fof, cond); ff != nil {
			ff = qeopt.reconstruct(fqs, ff, cond)
			ff = qeopt.simplify(ff, cond)
			qeopt.log(cond, 2, "vsret", "%v\n", fof)
			return ff
		}
	}

	////////////////////////////////
	// 非等式 QE
	////////////////////////////////
	if (qeopt.Algo & QEALGO_NEQ) != 0 {
		if ff := qeopt.qe_neq(fof, cond); ff != nil {
			ff = qeopt.reconstruct(fqs, ff, cond)
			ff = qeopt.simplify(ff, cond)
			qeopt.log(cond, 2, "neq", "%v\n", fof)
			return ff
		}
	}

	////////////////////////////////
	// CAD ではどうしようもないが, VS 2 次が使えるかも?
	////////////////////////////////

	if ff := qeopt.qe_simpl(fof, cond); ff != nil {
		return ff
	}

	////////////////////////////////
	// CAD
	// @TODO 前調査で多項式がおおかったら分配する、のも手ではないか.
	////////////////////////////////
	return qeopt.qe_cad(fof, cond)
}

func (qeopt QEopt) is_easy_cond(fof Fof, cond Fof) bool {
	switch c := cond.(type) {
	case *AtomT, *AtomF:
		return false // 追加する必要がないということ
	case *Atom:
		if !c.isUnivariate() {
			return false
		}
		return fof.hasVar(c.p[0].lv)
	case FofAO:
		for _, f := range c.Fmls() {
			if !qeopt.is_easy_cond(fof, f) {
				return false
			}
		}
	}
	return true
}

func (qeopt QEopt) appendNecSuf(qff Fof, cond qeCond) Fof {
	switch nec := cond.neccon.(type) {
	case *FmlAnd:
		for _, f := range nec.Fmls() {
			if qeopt.is_easy_cond(qff, f) {
				qff = NewFmlAnd(qff, f)
			}
		}
	default:
		if qeopt.is_easy_cond(nec, qff) {
			qff = NewFmlAnd(qff, nec)
		}
	}

	switch suf := cond.sufcon.(type) {
	case *FmlOr:
		for _, f := range suf.Fmls() {
			if qeopt.is_easy_cond(qff, f) {
				qff = NewFmlOr(qff, f)
			}
		}
	default:
		if qeopt.is_easy_cond(suf, qff) {
			qff = NewFmlAnd(qff, suf)
		}
	}

	return qff
}

func (qeopt QEopt) qe_cad(fof FofQ, cond qeCond) Fof {
	qeopt.log(cond, 2, "qecad", "%v\n", fof)
	qeopt.log(cond, 3, "qecad", "nec=%v\n", cond.neccon)
	qeopt.log(cond, 3, "qecad", "suf=%v\n", cond.sufcon)
	// 変数順序を入れ替える. :: 自由変数 -> 束縛変数
	maxvar := qeopt.varn

	b := make([]bool, maxvar)
	fof.Indets(b)
	numvar := 0
	for _, b := range b {
		if b {
			numvar++
		}
	}

	// 自由変数を探す
	fq := fof
	fqs := make([]FofQ, 1, maxvar)
	fqs[0] = fof
	var qff Fof
	for {
		qs := fq.Qs()
		for _, q := range qs {
			b[q] = false
		}
		if ff, ok := fq.Fml().(FofQ); ok {
			fq = ff
			fqs = append(fqs, fq)
		} else {
			qff = fq.Fml()
			break
		}
	}

	// index の下位が自由変数
	m := Level(0)
	o1 := make([]Level, len(b))
	o2 := make([]Level, 0, len(b))
	for i := range o1 {
		o1[i] = -1
	}

	for j, bi := range b {
		if bi {
			o1[j] = m
			o2 = append(o2, Level(j))
			m++
		}
	}

	if m > 1 {
		qff = qeopt.appendNecSuf(qff, cond)
		// 必要条件と十分条件をつけた論理式を再構築
		for i := len(fqs) - 1; i >= 0; i-- {
			qff = fqs[i].gen(fqs[i].Qs(), qff)
		}
		fof = qff.(FofQ)
	}
	qff = nil

	// 外側の限量子から追加
	fq = fof
	for {
		qs := fq.Qs()
		for _, q := range qs {
			o1[q] = m
			o2 = append(o2, q)
			m++
		}
		if ff, ok := fq.Fml().(FofQ); ok {
			fq = ff
		} else {
			break
		}
	}

	// 変数変換 (CAD用に
	fof2 := fof.varShift(+maxvar)
	lvs := make([]Level, 0, len(o2))
	vas := make([]RObj, 0, len(o2))
	for j := len(o1) - 1; j >= 0; j-- {
		if o1[j] >= 0 {
			lvs = append(lvs, Level(j)+maxvar)
			vas = append(vas, NewPolyVar(o1[j]))
		}
	}
	fof2 = fof2.replaceVar(vas, lvs)
	qeopt.log(cond, 2, "cad", "%v\n", fof2)
	cad, err := NewCAD(fof2, qeopt.g)
	if err != nil {
		panic(fmt.Sprintf("cad.lift() input=%v\nerr=%v", fof2, err))
	}
	cad.Projection(PROJ_McCallum)
	err = cad.Lift()
	for err != nil {
		if err != CAD_NO_WO {
			panic(fmt.Sprintf("cad.lift() input=%v\nerr=%v", fof, err))
		}
		qeopt.log(cond, 1, "cad", "not well-oriented %v\n", fof2)

		// NOT well-oriented で Hong-proj へ
		cad, _ = NewCAD(fof2, qeopt.g)
		cad.Projection(PROJ_HONG)
		err = cad.Lift()
	}
	fof3, err := cad.Sfc()
	if err != nil {
		panic(fmt.Sprintf("cad.sfc() input=%v\nerr=%v", fof, err))
	}

	lvs = make([]Level, 0, len(o2))
	vas = make([]RObj, 0, len(o2))
	for j := len(o2) - 1; j >= 0; j-- {
		lvs = append(lvs, Level(j))
		vas = append(vas, NewPolyVar(o2[Level(j)]+maxvar))
	}
	fof3 = fof3.replaceVar(vas, lvs)
	fof3 = fof3.varShift(-maxvar)
	return fof3
}

func (qeopt QEopt) qe_nonpreq(fofq FofQ, cond qeCond) Fof {
	qeopt.log(cond, 2, "qenpr", "%v\n", fofq)
	fs := make([]FofQ, 1)
	fs[0] = fofq
	for {
		fml := fofq.Fml()
		if fml.IsQuantifier() {
			fs = append(fs, fml.(FofQ))
		} else if fmlao, ok := fml.(FofAO); ok {
			fml = qeopt.qe_andor(fmlao, cond)

			// quantifier の再構築
			for i := len(fs) - 1; i >= 0; i-- {
				fml = fs[i].gen(fs[i].Qs(), fml)
			}
			fml = qeopt.qe_prenex(fml.(FofQ), cond)
			if fml.IsQff() {
				return fml
			}
		} else {
			panic("?")
		}
	}
}

func (qeopt QEopt) qe_andor(fof FofAO, cond qeCond) Fof {
	// fof: non-prenex-formula
	qeopt.log(cond, 2, "qeao", "%v\n", fof)
	fmls := fof.Fmls()
	sort.Slice(fmls, func(i, j int) bool {
		return qeopt.fmlcmp(fmls[i], fmls[j])
	})

	for i, f := range fmls {
		var cond2 qeCond

		// cond の構築 @TODO
		cond2 = cond
		cond2.depth = cond.depth + 1
		foth := make([]Fof, 0, len(fmls))
		// とりま atom だけでいいかな...
		for j, g := range fmls {
			if a, ok := g.(*Atom); ok && i != j {
				foth = append(foth, a)
			}
		}
		if len(foth) > 0 {
			necsuf := fof.gen(foth)
			if fof.isAnd() {
				// i 以外は必要条件でしょう.
				cond2.neccon = NewFmlAnd(cond2.neccon, necsuf)
			} else {
				cond2.sufcon = NewFmlOr(cond2.sufcon, necsuf)
			}
		}

		qeopt.log(cond, 2, "qeao", "<%d,i> %v\n", i, f)
		f = qeopt.simplify(f, cond2)
		f = qeopt.qe(f, cond2)
		fmls[i] = qeopt.simplify(f, cond2)
		qeopt.log(cond, 2, "qeao", "<%d,o> %v\n", i, fmls[i])
		switch fmls[i].(type) {
		case *AtomT:
			if !fof.isAnd() {
				return fmls[i]
			}
		case *AtomF:
			if fof.isAnd() {
				return fmls[i]
			}
		}
	}

	return fof.gen(fmls)
}
