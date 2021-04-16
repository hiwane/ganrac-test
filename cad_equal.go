package ganrac

import (
	"fmt"
)

func (cell *Cell) isDE() bool {
	n := 0
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly != nil && len(c.defpoly.c) > 2 {
			n++
		}
	}
	return n > 1
}

func (cell *Cell) sym_zero_chk_node(cad *CAD, p *Poly) bool {
	for c := cell; c.lv >= 0; c = c.parent {
		if c.defpoly == nil {
			continue
		}
		fmt.Printf("sym_zero_chk_node(): p=%v\n", p)
		if len(c.defpoly.c) == 2 { // 1次である. 代入で消去
			deg := p.Deg(Level(c.lv))
			if deg == 0 {
				continue
			}
			coef1 := c.defpoly.c[1].Neg()
			dens := make([]RObj, deg+1)
			dens[0] = one
			dens[1] = coef1
			for i := 2; i <= deg; i++ {
				dens[i] = dens[i-1].Mul(coef1)
			}
			q := p.subst_frac(c.defpoly.c[0], dens, Level(c.lv))
			switch qq := q.(type) {
			case *Poly:
				p = qq
			default:
				if !qq.IsNumeric() {
					panic("??") // @DEBUG
				}
				return p.IsZero()
			}
		} else {
			if !p.isUnivariate() {
				panic("??") // @DEBUG
			}
			r := p.reduce(c.defpoly)
			fmt.Printf("sym_zero_chk_node(): q=%v\n", p)
			fmt.Printf("sym_zero_chk_node(): r=%v\n", r)
			return r.IsZero()
		}
	}
	if true {
		panic("to-ranai")
	}
	return p.IsZero()
}

func (cad *CAD) symbolic_equal(ci, cj *Cell) bool {
	if len(ci.defpoly.c) > len(cj.defpoly.c) {
		c := ci
		ci = cj
		cj = c
	}

	if !ci.parent.isDE() {
		if len(ci.defpoly.c) == 2 {
			return ci.sym_zero_chk_node(cad, cj.defpoly)
		} else if len(cj.defpoly.c) == 2 {
			return cj.sym_zero_chk_node(cad, ci.defpoly)
		}
	}

	ret := cad.symde_zero_chk(ci.defpoly, cj)
	fmt.Printf("symbolic_equal() ret=%d\n", ret)
	panic("stop")

	// return false
}

func (cad *CAD) symde_zero_chk(p *Poly, c *Cell) int {
	return 0
}
