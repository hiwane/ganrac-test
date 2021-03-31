package ganrac

type simpler interface {
	simplBasic(neccon, sufcon Fof) Fof // 手抜き簡単化
	simplFctr(g *Ganrac) Fof           // CA を要求
}
