package ganrac

// Automated Natural Language Geometry Math Problem Solving by Real Quantier Elimination
// Hidenao Iwane, Takuya Matsuzaki, Noriko Arai, Hirokazu Anai
// ADG2014

import (
	"fmt"
	"sort"
)

type QEopt struct {
	varn Level
}

type qeCond struct {
	neccon, sufcon Fof
	dnf            bool
	depth          int
}

func (opt *QEopt) num_var(f Fof) int {
	b := make([]bool, opt.varn)
	f.Indets(b)
	m := 0
	for _, b := range b {
		if b {
			m++
		}
	}
	return m
}

func (opt *QEopt) fmlcmp(f1, f2 Fof) bool {
	switch g1 := f1.(type) {
	case FofQ:
		switch g2 := f2.(type) {
		case FofQ:
			m1 := opt.num_var(g1)
			m2 := opt.num_var(g2)
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

			m1 := opt.num_var(g1)
			m2 := opt.num_var(g2)
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
			m1 := opt.num_var(g1)
			m2 := opt.num_var(g2)
			if m1 != m2 {
				return m1 < m2
			}

			m1 = g1.sotd()
			m2 = g1.sotd()
			return m1 <= m2
		}
	}
}

func (g *Ganrac) QE(fof Fof, opt QEopt) Fof {
	opt.varn = fof.maxVar() + 1

	var cond qeCond
	cond.neccon = trueObj
	cond.sufcon = falseObj

	return g.qe(fof, opt, cond)
}

func (g *Ganrac) qe(fof Fof, opt QEopt, cond qeCond) Fof {
	fmt.Printf("qe   [%4d] %v\n", cond.depth, fof)
	fof = fof.nonPrenex()
	switch fq := fof.(type) {
	case FofQ:
		if fof.isPrenex() {
			return g.qe_prenex(fq, opt, cond)
		}
		return g.qe_nonpreq(fq, opt, cond)
	case FofAO:
		if fof.IsQff() {
			return g.simplify(fof, opt, cond)
		}
		return g.qe_andor(fq, opt, cond)
	default:
		return fof
	}
}

func (g *Ganrac) simplify(qff Fof, opt QEopt, cond qeCond) Fof {
	qff = qff.simplFctr(g)
	qff = qff.simplBasic(cond.neccon, cond.sufcon)
	fmt.Printf("simpl.b! %v\n", qff)
	qff, _, _ = qff.simplNum(g, nil, nil)
	fmt.Printf("simpl.n! %v\n", qff)
	return qff
}

func (g *Ganrac) qe_prenex(fof FofQ, opt QEopt, cond qeCond) Fof {
	fmt.Printf("qepr [%4d] %v\n", cond.depth, fof)
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
					var cond2 qeCond = cond
					cond2.dnf = !ao.isAnd()
					cond2.depth = cond.depth + 1

					ret := make([]Fof, len(ao.Fmls()))
					for i, f := range ao.Fmls() {
						ret[i] = fofq.gen(fofq.Qs(), f)
					}

					ao = ao.gen(ret).(FofAO)
					fmlq := g.qe_andor(ao, opt, cond2)
					for i := len(fs) - 2; i >= 0; i-- {
						fmlq = fs[i].gen(fs[i].Qs(), fmlq)
					}

					return g.qe(fmlq, opt, cond)
				}
			}
			break
		}
	}

	// もうがんばるしかない状態.
	return g.qe_prenex_main(fofq, opt, cond)
}

func (g *Ganrac) qe_prenex_main(fof FofQ, opt QEopt, cond qeCond) Fof {
	fmt.Printf("qepm [%4d] %v\n", cond.depth, fof)
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
	// Hong93 # {{{
	// 線形か2次の等式制約が含まれる場合.
	////////////////////////////////

	////////////////////////////////
	// VS を適用できるか. {{{
	////////////////////////////////

	////////////////////////////////
	// 非等式 QE {{{
	////////////////////////////////

	////////////////////////////////
	// CAD ではどうしようもないが, VS 2 次が使えるかも? # {{{
	////////////////////////////////

	////////////////////////////////
	// CAD
	// @TODO 前調査で多項式がおおかったら分配する、のも手ではないか.
	////////////////////////////////
	return g.qe_cad(fof, opt, cond)
}

func (g *Ganrac) qe_cad(fof FofQ, opt QEopt, cond qeCond) Fof {
	fmt.Printf("qecad[%4d] %v\n", cond.depth, fof)
	// 変数順序を入れ替える. :: 自由変数 -> 束縛変数
	maxvar := opt.varn

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
	for {
		qs := fq.Qs()
		for _, q := range qs {
			b[q] = false
		}
		if ff, ok := fq.Fml().(FofQ); ok {
			fq = ff
		} else {
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
	fmt.Printf("  cad[%4d] %v\n", cond.depth, fof2)
	cad, err := NewCAD(fof2, g)
	if err != nil {
		panic(fmt.Sprintf("cad.lift() input=%v\nerr=%v", fof2, err))
	}
	cad.Projection(0)
	err = cad.Lift()
	if err != nil {
		// @TODO well-oriented なら Hong-proj へ
		panic(fmt.Sprintf("cad.lift() input=%v\nerr=%v", fof, err))
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

func (g *Ganrac) qe_nonpreq(fofq FofQ, opt QEopt, cond qeCond) Fof {
	fmt.Printf("qenpr[%4d] %v\n", cond.depth, fofq)
	fs := make([]FofQ, 1)
	fs[0] = fofq
	for {
		fml := fofq.Fml()
		if fml.IsQuantifier() {
			fs = append(fs, fml.(FofQ))
		} else if fmlao, ok := fml.(FofAO); ok {
			fml = g.qe_andor(fmlao, opt, cond)

			// quantifier の再構築
			for i := len(fs) - 1; i >= 0; i-- {
				fml = fs[i].gen(fs[i].Qs(), fml)
			}
			return g.qe_prenex(fml.(FofQ), opt, cond)
		} else {
			panic("?")
		}
	}
}

func (g *Ganrac) qe_andor(fof FofAO, opt QEopt, cond qeCond) Fof {
	// fof: non-prenex-formula
	fmt.Printf("qeao [%4d] %v\n", cond.depth, fof)
	fmls := fof.Fmls()
	sort.Slice(fmls, func(i, j int) bool {
		return opt.fmlcmp(fmls[i], fmls[j])
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

		fmt.Printf("qeao [%4d,%d,1] %v\n", cond.depth, i, f)
		f = g.simplify(f, opt, cond2)
		fmt.Printf("qeao [%4d,%d,2] %v\n", cond.depth, i, f)
		f = g.qe(f, opt, cond2)
		fmt.Printf("qeao [%4d,%d,3] %v\n", cond.depth, i, f)
		fmls[i] = g.simplify(f, opt, cond2)
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
