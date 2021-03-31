package ganrac

func (ox *OpenXM) Factor(p *Poly) *List {
	// 因数分解
	ox.ExecFunction("fctr", p)
	s, _ := ox.PopCMO()
	gob := ox.toGObj(s)
	return gob.(*List)
}
