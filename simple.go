package ganrac

type reduce_info struct {
	depth int
	q     []Level
	qn    int
	vars  *List
	varb  []bool
	eqns  *List
}

type simpler interface {
	simplBasic(neccon, sufcon Fof) Fof           // 手抜き簡単化
	simplComm() Fof                              // 共通部分の括りだし
	simplFctr(g *Ganrac) Fof                     // CA を要求
	simplReduce(g *Ganrac, inf *reduce_info) Fof // 等式制約による簡約化

	// symbolic-numeric simplification
	simplNum(g *Ganrac, true_region, false_region *NumRegion) (Fof, *NumRegion, *NumRegion)
	get_homo_cond(conds [][]int, c []int) [][]int
	homo_reconstruct(lv Level, lvs Levels, sgn int) Fof
}

func (g *Ganrac) simplFof(c Fof, neccon, sufcon Fof) Fof {
	g.log(3, "simpl %v\n", c)
	c = c.simplFctr(g)
	c.normalize()
	inf := newReduceInfo()
	c = c.simplReduce(g, inf)

	for {
		cold := c
		c = c.simplComm()
		c = c.simplBasic(neccon, sufcon)
		c, _, _ = c.simplNum(g, nil, nil)
		if c.Equals(cold) {
			break
		}
	}
	return c
}
