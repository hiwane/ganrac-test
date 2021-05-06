package ganrac

type simpler interface {
	simplBasic(neccon, sufcon Fof) Fof // 手抜き簡単化
	simplFctr(g *Ganrac) Fof           // CA を要求

	// symbolic-numeric simplification
	simplNum(g *Ganrac, true_region, false_region *NumRegion) (Fof, *NumRegion, *NumRegion)
}
