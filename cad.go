package ganrac

import (
	"fmt"
	"io"
)

type ProjectionAlgo int

type sign_t int8

const (
	t_undef  = -1 // まだ評価していない
	t_false  = 0
	t_true   = 1
	t_other  = 2 // 兄弟の情報で親の真偽値が確定したのでもう評価しない
	q_forall = t_false
	q_exists = t_true
)

type qIntrval struct {
	// int, rat, binint が入ると想定
	inf, sup NObj
}

type AtomProj struct {
	Atom
	pl *ProjLink
}

type Cell struct {
	parent       *Cell
	children     []*Cell
	index        uint
	defpoly      *Poly
	intv         qIntrval  // 有理数=defpoly=nil か，bin-interval
	nintv        *Interval // 数値計算. defpoly=multivariate, de=true
	ex_deg       int       // 拡大次数
	signature    []sign_t
	multiplicity []int8
	lv           int8
	truth        int8
	de           bool
	sgn_of_left  sign_t
}

type cellStack struct {
	stack []*Cell
}

type ProjFactor struct {
	p       *Poly
	index   uint
	input   bool // 入力の論理式に含まれるか.
	coeff   []*ProjLink
	discrim *ProjLink
}

type ProjFactors struct {
	pf        []*ProjFactor
	resultant [][]*ProjLink
}

type ProjLink struct {
	sgn          sign_t
	multiplicity []uint
	projs        ProjFactors
}

type CADStat struct {
	fctr         int
	qrealroot    int
	irealroot    int
	irealroot_ok int
	sqrt         int
	sqrt_ok      int
	discriminant int
	resultant    int
	cell         int
	true_cell    int
	false_cell   int
	precision    int
	lift         int
}

type CAD struct {
	fml      Fof            // qff
	q        []int8         // quantifier
	proj     []*ProjFactors // [level]
	pl4const []*ProjLink    // 定数用 0, +, -
	stack    *cellStack
	root     *Cell
	g        *Ganrac
	stat     CADStat
}

func qeCAD(fml Fof) Fof {
	return fml
}

func (stat CADStat) Print(b io.Writer) {
	fmt.Fprintf(b, "CAD stat....\n")
	fmt.Fprintf(b, "===========\n")
	fmt.Fprintf(b, " - # of cells/true/false: %d / %d / %d\n", stat.cell, stat.true_cell, stat.false_cell)
	fmt.Fprintf(b, " - # of lifting         : %d\n", stat.lift)
	fmt.Fprintf(b, "\n")
	fmt.Fprintf(b, "CA stat....\n")
	fmt.Fprintf(b, "===========\n")
	fmt.Fprintf(b, " - discrim on proj      : %8d\n", stat.discriminant)
	fmt.Fprintf(b, " - resultant on proj    : %8d\n", stat.resultant)
	fmt.Fprintf(b, " - factorization over Z : %8d\n", stat.fctr)
	fmt.Fprintf(b, " - real root in Z[x]    : %8d\n", stat.qrealroot)
	fmt.Fprintf(b, " - real root in intv[x] : %8d / %d\n", stat.irealroot_ok, stat.irealroot)

}

func (c *CAD) Tag() uint {
	return TAG_CAD
}

func (c *CAD) String() string {
	return fmt.Sprintf("CAD[%v]", c.fml)
}

func NewCAD(prenex_formula Fof, g *Ganrac) (*CAD, error) {
	if g.ox == nil {
		return nil, fmt.Errorf("ox is required")
	}

	// @TODO 変数順序の変換....

	c := new(CAD)
	c.g = g
	v := prenex_formula.maxVar()
	c.q = make([]int8, v)
	for i := 0; i < len(c.q); i++ {
		c.q[i] = -1
	}
	c.fml = prenex_formula
	for {
		var qq []Level
		var qval int8
		switch f := c.fml.(type) {
		case *ForAll:
			qq = f.q
			qval = q_forall
			c.fml = f.fml
		case *Exists:
			qq = f.q
			qval = q_exists
			c.fml = f.fml
		default:
			goto _NEXT
		}

		max := Level(0)
		min := v
		for _, qi := range qq {
			c.q[qi] = qval
			if min > qi {
				min = qi
			}
			if max < qi {
				max = qi
			}
		}
		if int(max-min) != len(qq)-1 || max+1 != v {
			return nil, fmt.Errorf("CAD: invalid variable order [%d,%d,%d]", min, max, v)
		}

		v = min
	}
_NEXT:

	if !c.fml.IsQff() {
		return nil, fmt.Errorf("prenex formula is expected")
	}

	/////////////////////////////////////
	// projection の準備
	/////////////////////////////////////
	c.initProj(Level(len(c.q)))

	c.root = NewCell(c, nil, 0)
	c.stack = newCellStack()
	c.stack.push(c.root)

	return c, nil
}

func (c *CAD) initProj(v Level) {
	c.proj = make([]*ProjFactors, v)
	for i := Level(0); i < v; i++ {
		c.proj[i] = new(ProjFactors)
		c.proj[i].pf = make([]*ProjFactor, 0)
	}

	c.fml = clone4CAD(c.fml, c)
	c.pl4const = make([]*ProjLink, 3)
	for i, s := range []sign_t{0, 1, -1} {
		c.pl4const[i] = newProjLink()
		c.pl4const[i].sgn = s
	}
}

func clone4CAD(formula Fof, c *CAD) Fof {
	switch fml := formula.(type) {
	case *FmlAnd:
		var t Fof = trueObj
		for _, f := range fml.fml {
			t = NewFmlAnd(t, clone4CAD(f, c))
		}
		return t
	case *FmlOr:
		var t Fof = falseObj
		for _, f := range fml.fml {
			t = NewFmlOr(t, clone4CAD(f, c))
		}
		return t
	case *Atom:
		t := new(AtomProj)
		t.op = fml.op
		t.p = make([]*Poly, len(fml.p))
		t.pl = new(ProjLink)
		t.pl.sgn = 1
		for i, poly := range fml.p {
			pl2 := c.addPoly(poly, true)
			t.pl.merge(pl2)
			t.p[i] = poly
		}
		return t
	}
	panic("stop")
}

func NewCell(cad *CAD, parent *Cell, idx uint) *Cell {
	cell := new(Cell)
	cell.parent = parent
	cell.index = idx
	cell.ex_deg = 1
	cell.truth = t_undef
	if parent == nil {
		cell.lv = -1
	} else {
		cell.lv = parent.lv + 1
		cell.signature = make([]sign_t, len(cad.proj[cell.lv].pf))
		cell.multiplicity = make([]int8, len(cad.proj[cell.lv].pf))
	}
	return cell
}

func newCellStack() *cellStack {
	cs := new(cellStack)
	cs.stack = make([]*Cell, 0, 10000)
	return cs
}

func (cs *cellStack) empty() bool {
	return len(cs.stack) == 0
}

func (cs *cellStack) push(c *Cell) {
	cs.stack = append(cs.stack, c)
}

func (cs *cellStack) pop() *Cell {
	cell := cs.stack[len(cs.stack)-1]
	cs.stack = cs.stack[:len(cs.stack)-1]
	return cell
}

func (cad *CAD) Print(b io.Writer, args ...interface{}) error {
	if len(args) == 0 {
		fmt.Fprintf(b, "input=%v\n", cad.fml)
		return nil
	}
	s, ok := args[0].(*String)
	if !ok {
		return fmt.Errorf("invalid argument [expect string]")
	}

	switch s.s {
	case "stat":
		cad.stat.Print(b)
	case "proj":
	case "cells", "cell":
		cad.root.Print(b, args...)
	default:
		return fmt.Errorf("invalid argument")
	}

	return nil
}

func (cell *Cell) printSignature(b io.Writer) {
	ch := '('
	for j := 0; j < len(cell.signature); j++ {
		sgns := '+'
		if cell.signature[j] < 0 {
			sgns = '-'
		} else if cell.signature[j] == 0 {
			sgns = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHJIKLMNOPQRSTUVWXYZ")[cell.multiplicity[j]]
		}
		fmt.Fprintf(b, "%c%c", ch, sgns)
		ch = ' '
	}
	fmt.Fprintf(b, ")")
}

func (cell *Cell) printMultiplicity(b io.Writer) {
	ch := '('
	for j := 0; j < len(cell.multiplicity); j++ {
		if cell.multiplicity[j] == 0 {
			fmt.Fprintf(b, "%c ", ch)
		} else {
			fmt.Fprintf(b, "%c%d", ch, cell.multiplicity[j])
		}
		ch = ' '
	}
	fmt.Fprintf(b, ")")
}

func (cell *Cell) stringTruth() string {
	if cell.truth < 0 {
		return "?"
	}
	return []string{"f", "t", "."}[cell.truth]

}

func (cell *Cell) Print(b io.Writer, args ...interface{}) error {
	s := "cell"
	if len(args) > 0 {
		s2, ok := args[0].(*String)
		if !ok {
			return fmt.Errorf("invalid argument [expect string/kind]")
		}
		s = s2.s
	}

	for i := 1; i < len(args); i++ {
		ii, ok := args[i].(*Int)
		if !ok {
			return fmt.Errorf("invalid argument [expect integer]")
		}
		if ii.Sign() < 0 || !ii.IsInt64() || cell.children == nil || ii.Int64() >= int64(len(cell.children)) {
			return fmt.Errorf("invalid argument [invalid index]")
		}
		cell = cell.children[ii.Int64()]
	}

	switch s {
	case "cells":
		if cell.children == nil {
			return fmt.Errorf("invalid argument [no child]")
		}
		fmt.Fprintf(b, "signature(%v) :: %v\n", cell.Index(), args)
		for i, c := range cell.children {
			fmt.Fprintf(b, "%2d,%s,", i, c.stringTruth())
			if c.children == nil {
				fmt.Fprintf(b, "  ")
			} else {
				fmt.Fprintf(b, "%2d", len(c.children))
			}
			fmt.Fprintf(b, " ")
			c.printSignature(b)
			fmt.Fprintf(b, " ")
			c.printMultiplicity(b)
			if c.intv.inf != nil {
				fmt.Fprintf(b, " [% e,% e]", c.intv.inf.Float(), c.intv.sup.Float())
			} else if c.nintv != nil {
				fmt.Fprintf(b, " [% e,% e]", c.nintv.inf, c.nintv.sup)
			}
			if c.defpoly != nil {
				fmt.Fprintf(b, " %.50v", c.defpoly)
			} else if c.intv.inf == c.intv.sup && c.index%2 == 1 {
				fmt.Fprintf(b, " %v", c.intv.inf)
			}
			fmt.Fprintf(b, "\n")
		}
		return nil
	case "cell":
		fmt.Fprintf(b, "--- infomation about the cell %v ---\n", cell.Index())
		fmt.Fprintf(b, "lv=%d, de=%v, exdeg=%d, truth=%d sgn=%d\n",
			cell.lv, cell.de, cell.ex_deg, cell.truth, cell.sgn_of_left)
		var num int
		if cell.children == nil {
			num = -1
		} else {
			num = len(cell.children)
		}
		fmt.Fprintf(b, "# of children=%d\n", num)
		fmt.Fprintf(b, "def.poly     =%v\n", cell.defpoly)
		fmt.Fprintf(b, "signature    =")
		cell.printSignature(b)
		fmt.Fprintf(b, "\n")
		if cell.intv.inf != nil {
			fmt.Fprintf(b, "iso.intv     =[%v,%v]\n", cell.intv.inf, cell.intv.sup)
			fmt.Fprintf(b, "             =[%e,%e]\n", cell.intv.inf.Float(), cell.intv.sup.Float())
		}
		if cell.nintv != nil {
			fmt.Fprintf(b, "iso.nintv    =%f\n", cell.nintv)
			fmt.Fprintf(b, "             =%e\n", cell.nintv)
		}
	}

	return fmt.Errorf("invalid argument [kind]")
}
