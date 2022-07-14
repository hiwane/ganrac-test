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
	n := p.deg()
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
		return zero
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
	for (I = N-J-1; I >= 0 && I-(N-J-1)+J >= 0; I--) {
		S[I][L-1] = coef(F, I-(N-J-1)+J, X);
	}

	for (D = N; D >= 0; D--) {
		BI = coef(G,D,X);
		for (I = 0; I < M - J && N-D+I < L-1; I++) {
			S[I+N-J][N-D+I] = BI;
		}
	}
	for (I= M-J-1; I >= 0; I--) {
		S[I+N-J][L-1] = coef(G, I-(M-J-1)+J, X);
	}
	return det(S);
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

func (ox *OpenXM) Sres(p *Poly, q *Poly, lv Level, k int32) RObj {
	if !ox.sres_defined {
		str := `def comb(A,B) {
	for (I=1, C=1; I<=B; I++) {
		C *= (A-I+1)/I;
	}
	return C;
}`
		ox.ExecString(str)

		str = `def sres(F, G, X, K) {
    M = deg(F, X);
    N = deg(G, X);

	if (type(K) == 10) {
		K = int32ton(K);
	}
	L = N - K;
	S = newmat(L+1, L+1);
	CMK = comb(M, K+1);

	for (J = 0; J < L; J++) {
		S[0][J] = coef(F, M-J, X);
		for (I = 1; J+I < L; I++) {
			S[I][I+J] = S[0][J];
		}
		if (0 <= L-J && L-J <= L) {
			S[L-J][L] = (CMK - comb(M-J, K+1)) * S[0][J];
		}
		S[L][J] = coef(G, N-J, X);
	}
	J = L;
	S[0][L] = (CMK - comb(M-J, K+1)) * coef(F, M-J, X);
	S[L][L] = CMK * coef(G, K, X);
	return det(S);
}`
		ox.ExecString(str)
		ox.sres_defined = true
	}

	err := ox.ExecFunction("sres", p, q, NewPolyVar(lv), k)
	if err != nil {
		fmt.Printf("err: sres1 %s\n", err.Error())
	}
	qq, err := ox.PopCMO()
	if err != nil {
		fmt.Printf("err: sres2 %s\n", err.Error())
		return nil
	} else if qq == nil {
		return zero
	}
	return ox.toGObj(qq).(RObj)
}

func (ox *OpenXM) GB(p *List, vars *List, n int) *List {
	// グレブナー基底
	var err error

	if n == 0 {
		err = ox.ExecFunction("nd_gr", p, vars, zero, zero)
	} else {
		// block order
		vn := NewList(
			NewList(zero, NewInt(int64(vars.Len()-n))),
			NewList(zero, NewInt(int64(n))))

		err = ox.ExecFunction("nd_gr", p, vars, zero, vn)
	}
	if err != nil {
		panic(fmt.Sprintf("gr failed: %v", err.Error()))
	}
	s, err := ox.PopCMO()
	if err != nil {
		fmt.Printf("gr failed: %v", err.Error())
		return nil
	}

	gob := ox.toGObj(s)
	return gob.(*List)
}

func (ox *OpenXM) Reduce(p *Poly, gb *List, vars *List, n int) (RObj, bool) {

	var err error

	if n == 0 {
		err = ox.ExecFunction("p_true_nf", p, gb, vars, zero)
	} else {
		// block order
		vn := NewList(
			NewList(zero, NewInt(int64(vars.Len()-n))),
			NewList(zero, NewInt(int64(n))))

		err = ox.ExecFunction("p_true_nf", p, gb, vars, vn)
	}
	if err != nil {
		panic(fmt.Sprintf("p_nf failed: %v", err.Error()))
	}
	s, err := ox.PopCMO()
	if err != nil {
		fmt.Printf("p_nf failed: %v", err.Error())
		return nil, false
	}

	gob := ox.toGObj(s).(*List)

	m, _ := gob.Geti(1)
	mm, ok := m.(NObj)
	if !ok {
		panic(fmt.Sprintf("p_nf failed:\np=%v\ngb=%v\nret=%v\nden=%v", p, gb, gob, m))
	}
	sgn := mm.Sign() < 0

	m, _ = gob.Geti(0)

	return m.(RObj), sgn
}
