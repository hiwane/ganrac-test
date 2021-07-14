package ganrac

/////////////////////////////////////
// Formula Simplification for Real Quantifier Elimination
// Using Geometric Invariance
// H. Iwane, H. Anai, ISSAC 2017
// https://doi.org/10.1145/3087604.3087627
//
// homogeneous formula ver. (scale invariant)
// related: simpl_tran
// related: simpl_rot
/////////////////////////////////////

import (
	"fmt"
)

func (p *AtomT) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	return p
}

func (p *AtomF) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	return p
}

func (p *ForAll) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	panic("?")
}

func (p *Exists) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	panic("?")
}

func (p *FmlAnd) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	fs := make([]Fof, p.Len())
	for i, f := range p.Fmls() {
		fs[i] = f.homo_reconstruct(lv, lvs, sgn)
	}
	return p.gen(fs)
}

func (p *FmlOr) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	fs := make([]Fof, p.Len())
	for i, f := range p.Fmls() {
		fs[i] = f.homo_reconstruct(lv, lvs, sgn)
	}
	return p.gen(fs)
}

func (atom *Atom) homo_reconstruct(lv Level, lvs Levels, sgn int) Fof {
	ps := make([]*Poly, len(atom.p))
	tdeg := 0
	for i, p := range atom.p {
		ps[i] = p
		for _, v := range lvs {
			if p.hasVar(v) {
				deg := p.tdeg(lvs)
				if deg == 0 {
					panic("?")
				}
				tdeg += deg
				ps[i] = p.homo_reconstruct(lv, lvs, deg).(*Poly)
				break
			}
		}
	}
	if tdeg == 0 {
		return atom
	} else if sgn > 0 || tdeg%2 == 0 {
		return newAtoms(ps, atom.op)
	} else {
		return newAtoms(ps, atom.op.neg())
	}
}

func (p *Poly) homo_reconstruct(lv Level, lvs Levels, tdeg int) RObj {
	u := 0
	if lvs.contains(p.lv) {
		u = 1
	}

	var q RObj = zero
	for i, cc := range p.c {
		d := tdeg - u*i
		if cc.IsZero() {
			continue
		} else if c, ok := cc.(*Poly); ok {
			cc = c.homo_reconstruct(lv, lvs, d)
		} else if d > 0 {
			xn := newPolyVarn(lv, d)
			xn.c[d] = cc
			cc = xn
		}
		if i > 0 {
			x := newPolyVarn(p.lv, i)
			q = Add(q, Mul(x, cc))
		} else {
			q = Add(q, cc)
		}
		if err := q.valid(); err != nil {
			panic(fmt.Sprintf("err=%v\ncc=%v\nq=%v\n", err, cc, q))
		}
	}
	return q
}

func (p *AtomT) get_homo_cond(conds [][]int, b []int) [][]int {
	return conds
}

func (p *AtomF) get_homo_cond(conds [][]int, b []int) [][]int {
	return conds
}

func (p *FmlAnd) get_homo_cond(conds [][]int, b []int) [][]int {
	for _, f := range p.Fmls() {
		conds = f.get_homo_cond(conds, b)
	}
	return conds
}

func (p *FmlOr) get_homo_cond(conds [][]int, b []int) [][]int {
	for _, f := range p.Fmls() {
		conds = f.get_homo_cond(conds, b)
	}
	return conds
}

func (p *ForAll) get_homo_cond(conds [][]int, b []int) [][]int {
	return p.Fml().get_homo_cond(conds, b)
}

func (p *Exists) get_homo_cond(conds [][]int, b []int) [][]int {
	return p.Fml().get_homo_cond(conds, b)
}

func (atom *Atom) get_homo_cond(conds [][]int, ret []int) [][]int {
	// homo になるときの条件を満たす条件の一覧を構築する.
	// Param:
	//   conds: これまで構築した条件一覧
	//   ret: homo 対象にならない変数に 0 を設定
	for _, p := range atom.p {
		c := p.constantTerm()
		if !c.IsZero() {
			bb := make([]bool, p.lv+1)
			atom.Indets(bb)
			for i, b := range bb {
				if b {
					ret[i] = 0
				}
			}
		} else {
			deg := make([]int, len(ret))
			p.lm(deg)

			conds = p.get_homo_cond(conds, deg)
		}
	}

	// 定数項は 0.
	return conds
}

func (p *Poly) get_homo_cond(conds [][]int, deg []int) [][]int {
	for i, cc := range p.c {
		switch c := cc.(type) {
		case *Poly:
			deg[p.lv] -= i
			conds = c.get_homo_cond(conds, deg)
			deg[p.lv] += i
		default:
			if !c.IsZero() {
				cs := make([]int, len(deg))
				copy(cs, deg)
				cs[p.lv] -= i
				conds = append(conds, cs)
			}
		}
	}
	return conds
}

func (qeopt QEopt) homo_solve(amat [][]int, bmat, x []int) bool {
	// a x = b を解く.
	// x \in [0, 1] .... (そのうち [0, 1, ....] ? k-homo, quasi-homo)
	// x=-1 は undefined を表す
	// returns: true if solved

	for len(amat) > 0 {
		up := false
		pos := make([]int, len(amat))
		neg := make([]int, len(amat))
		newa := make([][]int, 0, len(amat))
		newb := make([]int, 0, len(bmat))
		b := make([]int, len(bmat))
		copy(b, bmat)
		for i, _cond := range amat {
			cond := make([]int, len(_cond))
			copy(cond, _cond)

			mpos := 0
			mneg := 0
			for j, c := range cond {
				if x[j] >= 0 {
					b[i] -= cond[j] * x[j]
					cond[j] = 0
				} else if c > 0 {
					pos[i]++
					mpos += c
				} else if c < 0 {
					neg[i]++
					mneg += c
				}
			}

			if pos[i] == 0 && neg[i] == 0 {
				if b[i] != 0 {
					return false // 不成立
				}
			} else if pos[i] == 1 && neg[i] == 1 && b[i] != mpos && b[i] != mneg ||
				mpos < b[i] || mneg > b[i] {
				// 正と負の成分が 1個ずつかつ，b が非ゼロならどちらかの成分と一致しないと.
				// 正の成分の和が b[i] を超えられないなら，偽
				return false
			} else if b[i] == 0 && (neg[i] == 0 || pos[i] == 0 ||
				pos[i] == 1 && neg[i] == 1 && mpos+mneg != 0) {
				// 全部 0 が確定
				for j, c := range cond {
					if c != 0 {
						x[j] = 0
					}
				}
				up = true
			} else if b[i] == mpos {
				// 正の成分全部1
				for j, c := range cond {
					if c > 0 {
						x[j] = 1
					} else if c < 0 {
						x[j] = 0
					}
				}
				up = true
			} else if b[i] == mneg {
				// 正の成分全部1
				for j, c := range cond {
					if c < 0 {
						x[j] = 1
					} else if c > 0 {
						x[j] = 0
					}
				}
				up = true
			} else {
				newa = append(newa, cond)
				newb = append(newb, b[i])
			}
		}

		if len(newa) == 0 {
			amat = newa
			break
		}

		amat = newa
		bmat = newb
		if !up {
			break
		}
	}

	if len(amat) == 0 {
		return true
	}

	// もう諦めて仮定法....
	for j, c := range amat[0] {
		if c != 0 {
			if x[j] >= 0 {
				panic(fmt.Sprintf("j=%v, x=%v\na=%v\n", j, x, amat))
			}
			xx := make([]int, len(x))
			for _, v := range []int{1, 0} {
				// @TODO 自由変数から試すべき.
				copy(xx, x)
				xx[j] = v
				u := qeopt.homo_solve(amat, bmat, xx)
				if u {
					copy(x, xx)
					return true
				}
			}
		}
	}
	return false
}

func (qeopt QEopt) qe_homo_free(fof FofQ, cond qeCond, d []int, lv Level) Fof {
	// lv: free variable.
	qeopt.log(cond, 2, "qehom", "<%s> %v %v\n", varstr(lv), fof, d)

	// cond を新しい変数に置き換える
	varn := qeopt.varn
	xs := make([]RObj, 0, len(d))
	lvs := make([]Level, 0, len(d))

	for i := 0; i < len(d); i++ {
		if d[i] > 0 {
			lvs = append(lvs, Level(i))
			xs = append(xs, NewPolyVar(qeopt.varn))
			qeopt.varn++
		}
	}

	var cond2 qeCond = cond
	cond2.depth++

	var fret Fof = falseObj
	for _, ss := range []struct {
		val int
		vi  *Int
		op  OP
	}{
		// 0, +1, -1 の順序を変更してはならない.
		{0, zero, EQ},
		{+1, one, GT},
		{-1, mone, LT},
	} {
		if ss.val > 0 {
			// TODO 新しい変数との関係を neccon に追加する
			for i := 0; i < len(xs); i++ {
				cond2.neccon = cond2.neccon.Subst(xs[i], lvs[i])
				cond2.sufcon = cond2.sufcon.Subst(xs[i], lvs[i])
			}
		}

		fp := fof.Subst(ss.vi, lv)
		if ss.val < 0 {
			for _, v := range lvs {
				if v != lv {
					xx := NewPolyVar(v).Neg()
					fp = fp.Subst(xx, v)
				}
			}
		}
		fp = qeopt.qe(fp, cond2)

		// 変数を元に戻す..... lv 変数を復旧する
		if ss.val != 0 {
			fp = fp.homo_reconstruct(lv, lvs, ss.val)
		}

		fp = NewFmlAnd(fp, NewAtom(NewPolyVar(lv), ss.op))
		fret = NewFmlOr(fret, fp)
	}

	// varn を戻す
	qeopt.varn = varn
	return fret
}

func (qeopt QEopt) qe_homo_quan(fof FofQ, cond qeCond, d []int) Fof {
	if !fof.isPrenex() { // めんどい
		return nil
	}
	fqs := make([]FofQ, 0)

	// 一番外側の束縛変数を探す．
	var f Fof = fof
	var lv Level = -1
	for {
		if fq, ok := f.(FofQ); ok {
			for _, q := range fq.Qs() {
				if d[q] > 0 {
					lv = q
					goto _found
				}
			}
			fqs = append(fqs, fq)
			f = fq.Fml()
		} else {
			return nil
		}
	}

_found:
	cond.depth++
	fs := make([]Fof, 3)
	for i, v := range []RObj{zero, one, mone} {
		fs[i] = f.Subst(v, lv)
		fs[i] = qeopt.qe(fs[i], cond)
	}
	switch f.(type) {
	case *ForAll:
		f = newFmlAnds(fs[0], fs[1], fs[2])
	case *Exists:
		f = newFmlOrs(fs[0], fs[1], fs[2])
	default:
		panic("")
	}

	return qeopt.reconstruct(fqs, f, cond)
}

func (qeopt QEopt) qe_homo(fof FofQ, cond qeCond) Fof {
	conds := make([][]int, 0)
	d := make([]int, int(qeopt.varn))
	for i := range d {
		d[i] = -1
	}
	conds = fof.get_homo_cond(conds, d)
	b := make([]int, len(conds))
	o := qeopt.homo_solve(conds, b, d)
	if !o {
		return nil
	}
	found := false
	lv := Level(-1)
	for i, x := range d {
		if x > 0 {
			found = true
			if fof.hasFreeVar(Level(i)) {
				lv = Level(i)
				break
			}
		}
	}
	if !found {
		return nil
	}
	if lv >= 0 {
		return qeopt.qe_homo_free(fof, cond, d, lv)
	} else {
		return qeopt.qe_homo_quan(fof, cond, d)
	}
}
