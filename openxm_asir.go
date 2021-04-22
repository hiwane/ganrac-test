package ganrac

import (
	"fmt"
)

func (ox *OpenXM) Gcd(p, q *Poly) RObj {
	ox.ExecFunction("gcd", p, q)
	s, _ := ox.PopCMO()
	gob := ox.toGObj(s)
	return gob.(RObj)
}

func (ox *OpenXM) Factor(p *Poly) *List {
	// 因数分解
	ox.ExecFunction("fctr", p)
	s, _ := ox.PopCMO()
	gob := ox.toGObj(s)
	return gob.(*List)
}

func (ox *OpenXM) Discrim(p *Poly, lv Level) RObj {
	dp := p.diff(lv)
	ox.ExecFunction("res", NewPolyVar(lv), p, dp)
	qq, _ := ox.PopCMO()
	fmt.Printf("qq=%v\n", qq)
	q := ox.toGObj(qq).(RObj)
	fmt.Printf("q=%v\n", q)
	n := len(p.c) - 1 // deg(p)
	if (n & 0x2) != 0 {
		q = q.Neg()
	}
	// 主係数で割る
	switch pc := p.c[n].(type) {
	case *Poly:
		fmt.Printf("poly...\n")
		return q.(*Poly).sdiv(pc)
	case NObj:
		return q.Div(pc)
	default:
		fmt.Printf("discrim: %v, pc=%v\n", p, pc)
	}
	return nil
}

func (ox *OpenXM) Resultant(p *Poly, q *Poly, lv Level) RObj {
	ox.ExecFunction("res", NewPolyVar(lv), p, q)
	qq, _ := ox.PopCMO()
	return ox.toGObj(qq).(RObj)
}

func (ox *OpenXM) GB(p *List, v uint) *List {
	// グレブナー基底
	var err error

	b := make([]bool, v)
	p.Indets(b)
	vars := NewList()
	for i := uint(0); i < v; i++ {
		if b[i] {
			vars.Append(NewPolyVar(Level(i)))
		}
	}

	err = ox.ExecFunction("gr", p, vars, one)
	if err != nil {
		panic(fmt.Sprintf("gr failed: %v", err.Error()))
	}
	s, err := ox.PopCMO()
	if err != nil {
		fmt.Sprintf("gr failed: %v", err.Error())
		return nil
	}

	gob := ox.toGObj(s)
	return gob.(*List)
}
