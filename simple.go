package ganrac

type simpler interface {
	simplBasic(neccon, sufcon Fof) Fof // 手抜き簡単化
	simplFctr(g *Ganrac) Fof           // CA を要求

	// symbolic-numeric simplification
	simplNum(g *Ganrac, true_region, false_region *NumRegion) (Fof, *NumRegion, *NumRegion)
}

func (g *Ganrac) simplFof(c Fof) Fof {
	c = c.simplFctr(g)
	c = c.simplBasic(trueObj, falseObj)
	c, _, _ = c.simplNum(g, nil, nil)
	return c
}
