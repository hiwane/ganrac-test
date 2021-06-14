package ganrac

import (
	//	"fmt"
	"sort"
)

/////////////////////////////////////
// 論理式の簡単化.
// 共通部分を括りだして，簡単化する.
//
// normalize が実施済みであること.
/////////////////////////////////////

func (p *AtomT) simplComm() Fof {
	return p
}

func (p *AtomF) simplComm() Fof {
	return p
}

func (p *Atom) simplComm() Fof {
	return p
}

func (p *FmlAnd) simplComm() Fof {
	q := p.Not()
	r := q.simplComm()
	if q == r {
		return p
	} else {
		return r.Not()
	}
}

func (p *FmlOr) simplComm() Fof {
	fmls := make([]Fof, 0, len(p.fml))
	for _, f := range p.fml {
		f = f.simplComm()
		if f == trueObj {
			return f
		}
		fmls = append(fmls, f)
	}

	sort.Slice(fmls, func(i, j int) bool {
		return fmls[i].FmlLess(fmls[j])
	})

	i := 0
	for i < len(fmls) {
		if _, ok := fmls[i].(*FmlAnd); ok {
			break
		}
		i++
	}

	im := i
	for im < len(fmls) {
		if _, ok := fmls[im].(*FmlAnd); !ok {
			break
		}
		im++
	}

	// A        || (A && C) ==> A && C
	// (A && B) || (A && C) ==> A && (B || C)
	for ; i < im; i++ {
		a := fmls[i].(*FmlAnd)
		for j := i + 1; j < im; j++ {
			a_u := make([]bool, len(a.fml))
			b := fmls[j].(*FmlAnd)
			b_u := make([]bool, len(b.fml))

			m := 0
			ix := -1
			for ii, af := range a.fml {
				for jj, bf := range b.fml {
					if af.Equals(bf) {
						a_u[ii] = true
						b_u[jj] = true
						m++
						break
					}
				}
				if !a_u[ii] {
					ix = ii
				}
			}
			if m == len(a.fml) {
				// 全部含まれた.... 不要
				fmls[i] = falseObj
				break
			} else if m+1 == len(a.fml) {
				ff := make([]Fof, len(a.fml))
				copy(ff, a.fml)

				bg := make([]Fof, 0, len(b.fml)-m)
				for iu, u := range b_u {
					if !u {
						bg = append(bg, b.fml[iu])
					}
				}
				ff[ix] = NewFmlOr(a.fml[ix], newFmlAnds(bg...))
				fmls[i] = falseObj
				fmls[j] = newFmlAnds(ff...)
				break
			}
		}
	}

	return newFmlOrs(fmls...)
}

func (p *ForAll) simplComm() Fof {
	fml := p.fml.simplComm()
	if fml == p.fml {
		return p
	}
	return NewQuantifier(true, p.q, fml)
}

func (p *Exists) simplComm() Fof {
	fml := p.fml.simplComm()
	if fml == p.fml {
		return p
	}
	return NewQuantifier(false, p.q, fml)
}
