package ganrac

import (
	"fmt"
	"io"
)

type ProjectionAlgo int

type sign_t int8

type Intv struct {
	l, u NObj
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
	intv         Intv
	ex_deg       int // 拡大次数
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

type CAD struct {
	fml      Fof            // qff
	q        []int8         // quantifier
	proj     []*ProjFactors // [level]
	pl4const []*ProjLink    // 定数用 0, +, -
	stack    *cellStack
	root     *Cell
	g        *Ganrac
}

func qeCAD(fml Fof) Fof {
	return fml
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
			qval = 0
			c.fml = f.fml
		case *Exists:
			qq = f.q
			qval = 1
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
	cell.truth = -1
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

func (cad *CAD) Print(b io.Writer, args []interface{}) error {
	if len(args) == 0 {
		fmt.Fprintf(b, "input=%v\n", cad.fml)
		return nil
	}
	s, ok := args[0].(*String)
	if !ok {
		return fmt.Errorf("invalid argument [expect string]")
	}

	switch s.s {
	case "proj":
	case "cells", "cell", "signature":
		cad.root.Print(b, args)
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
			sgns = '0'
		}
		fmt.Fprintf(b, "%c%c", ch, sgns)
		ch = ','
	}
	fmt.Fprintf(b, ")")
}

func (cell *Cell) printMultiplicity(b io.Writer) {
	ch := '('
	for j := 0; j < len(cell.multiplicity); j++ {
		fmt.Fprintf(b, "%c%d", ch, cell.multiplicity[j])
		ch = ','
	}
}

func (cell *Cell) Print(b io.Writer, args []interface{}) error {
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
	case "signature":
		if cell.children == nil {
			return fmt.Errorf("invalid argument [no child]")
		}
		fmt.Fprintf(b, "signature(%v) :: %v\n", cell.Index(), args)
		for i, c := range cell.children {
			fmt.Fprintf(b, "%2d,", i)
			if c.truth < 0 {
				fmt.Fprintf(b, " :")
			} else {
				fmt.Fprintf(b, "%c:", "ft"[c.truth])
			}
			c.printSignature(b)
			c.printMultiplicity(b)
			fmt.Fprintf(b, " [% e,% e]", c.intv.l.Float(), c.intv.u.Float())
			if c.defpoly != nil {
				fmt.Fprintf(b, " %.30v", c.defpoly)
			}
			fmt.Fprintf(b, "\n")
		}
		return nil
	case "cell":
		fmt.Fprintf(b, "--- infomation about the cell %v ---\n", cell.Index())
		fmt.Fprintf(b, "level        =%d\n", cell.lv)
		var num int
		if cell.children == nil {
			num = -1
		} else {
			num = len(cell.children)
		}
		fmt.Fprintf(b, "# of children=%d\n", num)
		fmt.Fprintf(b, "truth value  =%d\n", cell.truth)
		fmt.Fprintf(b, "def.poly     =%v\n", cell.defpoly)
		fmt.Fprintf(b, "iso.intv     =[%v,%v]\n", cell.intv.l, cell.intv.u)
		fmt.Fprintf(b, "             =[%e,%e]\n", cell.intv.l.Float(), cell.intv.u.Float())
	case "cells":
		if cell.children == nil {
			return fmt.Errorf("invalid argument [no child]")
		}
	}

	return fmt.Errorf("invalid argument [kind]")
}
