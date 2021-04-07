package ganrac

import (
	"fmt"
)

type ProjectionAlgo int

type sign int8

type Intv struct {
	l, u RObj
}

type AtomProj struct {
	Atom
	pl *ProjLink
}

type Cell struct {
	parent    *Cell
	children  []*Cell
	index     uint
	defpoly   *Poly
	intv      Intv
	signature []sign
}

type cellStack struct {
	stack []*Cell
}

type ProjFactor struct {
	p       *Poly
	index   int
	input   bool // 入力の論理式に含まれるか.
	coeff   []*ProjLink
	discrim *ProjLink
}

type ProjFactors struct {
	pf        []*ProjFactor
	resultant [][]*ProjLink
}

type ProjLink struct {
	sgn          int8
	multiplicity []uint
	projs        ProjFactors
}

type CAD struct {
	fml   Fof            // qff
	q     []int8         // quantifier
	proj  []*ProjFactors // [level]
	pl    []*ProjLink    // 定数用 0, +, -
	stack *cellStack
	root  *Cell
	g     *Ganrac
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
	fmt.Printf("varlist=%v\n", varlist)
	v := prenex_formula.maxVar()
	fmt.Printf("maxVar=%v\n", v)
	c.q = make([]int8, v)
	c.fml = prenex_formula
	for {
		var qq []Level
		var qval int8
		switch f := c.fml.(type) {
		case *ForAll:
			qq = f.q
			qval = 2
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

	return c, nil
}

func (c *CAD) initProj(v Level) {
	c.proj = make([]*ProjFactors, v)
	for i := Level(0); i < v; i++ {
		c.proj[i] = new(ProjFactors)
		c.proj[i].pf = make([]*ProjFactor, 0)
	}

	c.fml = clone4CAD(c.fml, c)
	c.pl = make([]*ProjLink, 3)
	for i, s := range []int8{0, 1, -1} {
		c.pl[i] = newProjLink()
		c.pl[i].sgn = s
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
