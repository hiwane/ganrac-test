package ganrac

import (
	"fmt"
)

func (z *Poly) RootBound() (NObj, error) {
	if !z.isUnivariate() {
		return nil, fmt.Errorf("supported only for univariate polynomial")
	}

	var m NObj
	m = z.c[0].(NObj)
	for i := 1; i < len(z.c)-1; i++ {
		if m.CmpAbs(z.c[i].(NObj)) < 0 {
			m = z.c[i].(NObj)
		}
	}
	m = m.Div(z.c[len(z.c)-1].(NObj)).(NObj)
	return m.Abs().AddInt(1), nil
}

func (z *Poly) rootBound2Exp() *Int {
	m, _ := z.RootBound()
	r := one
	for r.Cmp(m) < 0 {
		r = r.Mul(two).(*Int)
	}
	return r
}
