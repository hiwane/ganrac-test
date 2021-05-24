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
	q := ox.toGObj(qq).(RObj)
	n := len(p.c) - 1 // deg(p)
	if (n & 0x2) != 0 {
		q = q.Neg()
	}
	// 主係数で割る
	switch pc := p.c[n].(type) {
	case *Poly:
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
	qq, err := ox.PopCMO()
	if err != nil {
		fmt.Printf("resultant %s\n", err.Error())
	} else if qq == nil {
		fmt.Printf("resultant(%d)\np=%v\nq=%v\n", lv, p, q)
	}
	return ox.toGObj(qq).(RObj)
}

func (ox *OpenXM) Psc(p *Poly, q *Poly, lv Level, j int32) RObj {
	if !ox.psc_defined {
		str := `def psc(F, G, X, J) {
	local M, N, L, S, D, AI, BI, I;
    M = deg(F, X);
    N = deg(G, X);
	if (type(J) == 10) {
		J = int32ton(J);
	}
	L = M+N-2*J;
	S = newmat(L, L);

	for (D = M; D >= 0; D--) {
		AI = coef(F,D,X);
		for (I = 0; I < N - J && M-D+I < L-1; I++) {
			S[I][M-D+I] = AI;
		}
	}
	S[N-J-1][L-1] = F;
	for (I = N-J-2; I >= 0; I--) {
		S[I][L-1] = X * S[I+1][L-1];
	}

	for (D = N; D >= 0; D--) {
		BI = coef(G,D,X);
		for (I = 0; I < M - J && N-D+I < L-1; I++) {
			S[I+N-J][N-D+I] = BI;
		}
	}

	S[L-1][L-1] = G;
	for (I= M-J-2; I >= 0; I--) {
		S[I+N-J][L-1] = X * S[I+1+N-J][L-1];
	}
	return coef(det(S), J, X);
}`
		ox.ExecString(str)
		ox.psc_defined = true
	}

	err := ox.ExecFunction("psc", p, q, NewPolyVar(lv), j)
	if err != nil {
		fmt.Printf("err: psc1 %s\n", err.Error())
	}
	qq, err := ox.PopCMO()
	if err != nil {
		fmt.Printf("err: psc2 %s\n", err.Error())
		return nil
	} else if qq == nil {
		return zero
	}
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
