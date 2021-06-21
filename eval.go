package ganrac

import (
	"fmt"
	"os"
	"time"
)

func (g *Ganrac) evalStack(stack *pStack) (interface{}, error) {
	s, err := stack.Pop()
	// fmt.Printf("eval: s=%v\n", s)
	if err != nil {
		return nil, err
	}
	switch s.cmd {
	case plus, minus, mult, pow, div:
		return g.evalStackRObj2(stack, s)
	case ident:
		lv, err := var2lv(s.str)
		if err != nil {
			return nil, err
		}
		return NewPolyVar(lv), nil
	case name:
		return g.evalStackName(stack, s)
	case unaryminus:
		return g.evalStackRObj1(stack, s)
	case and, or:
		return g.evalStackFof2(stack, s)
	case geop:
		return g.evalStackAtom(stack, GE, s)
	case gtop:
		return g.evalStackAtom(stack, GT, s)
	case leop:
		return g.evalStackAtom(stack, LE, s)
	case ltop:
		return g.evalStackAtom(stack, LT, s)
	case eqop:
		return g.evalStackAtom(stack, EQ, s)
	case neop:
		return g.evalStackAtom(stack, NE, s)
	case call, list:
		return g.evalStackNvar(stack, s)
	case assign:
		return g.evalStackAssign(stack, s)
	case lb:
		return g.evalStackElem(stack, s)
	case number:
		bi := ParseInt(s.str, 10)
		if bi != nil {
			return bi, nil
		} else {
			return nil, fmt.Errorf("invalid number: %s", s.str)
		}
	case t_str:
		return g.evalStackString(stack, s)
	case initvar:
		return g.evalInitVar(stack, s.extra)
	case f_true:
		return trueObj, nil
	case f_false:
		return falseObj, nil
	case f_time:
		tm_start := time.Now()
		o, err := g.evalStack(stack)
		fmt.Fprintf(os.Stderr, "time: %.3f sec\n", time.Since(tm_start).Seconds())
		return o, err
	case vardol:
		return g.evalStackVarDol(stack, s)
	case eolq:
		for i := 0; i < s.extra; i++ {
			_, err := g.evalStack(stack)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	case eol:
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported [str=%s, cmd=%d]", s.str, s.cmd)
}

func (g *Ganrac) evalStackAtom(stack *pStack, op OP, node pNode) (Fof, error) {
	right, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	r, ok := right.(RObj)
	if !ok {
		return nil, fmt.Errorf("%s is not supported", node.str)
	}
	left, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	l, ok := left.(RObj)
	if !ok {
		return nil, fmt.Errorf("%s is not supported", node.str)
	}

	a := NewAtom(Sub(l, r), op)
	return a, nil
}

func (g *Ganrac) evalInitVar(stack *pStack, num int) (interface{}, error) {
	if num == 0 {
		v := NewList()
		for i := 0; i < len(varlist); i++ {
			v.Append(NewPolyVar(Level(i)))
		}
		return v, nil
	}

	vlist := make([]string, num)
	for i := num - 1; i >= 0; i-- {
		p, err := stack.Pop()
		if err != nil {
			return nil, err
		}
		vlist[i] = p.str
	}
	err := g.InitVarList(vlist)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (g *Ganrac) evalStackFof2(stack *pStack, node pNode) (interface{}, error) {
	right, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	r, ok := right.(Fof)
	if !ok {
		return nil, fmt.Errorf("%s: expected FOF", node.str)
	}
	left, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	l, ok := left.(Fof)
	if !ok {
		return nil, fmt.Errorf("%s: expected FOF", node.str)
	}
	switch node.cmd {
	case and:
		if err := l.valid(); err != nil {
			fmt.Printf("l=%v, err=%v, r=%v %p\n", l, err, r, l)
			panic("left")
		}
		if err := r.valid(); err != nil {
			fmt.Printf("r=%v, err=%v\n", r, err)
			panic("right")
		}
		return NewFmlAnd(l, r), nil
	case or:
		return NewFmlOr(l, r), nil
	}
	return nil, fmt.Errorf("%s is not supported", node.str)
}

func (g *Ganrac) evalStackRObj2(stack *pStack, node pNode) (interface{}, error) {
	right, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	r, ok := right.(RObj)
	if !ok {
		return nil, fmt.Errorf("%s is not supported", node.str)
	}
	left, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	l, ok := left.(RObj)
	if !ok {
		return nil, fmt.Errorf("%s is not supported", node.str)
	}
	switch node.cmd {
	case plus:
		return Add(l, r), nil
	case minus:
		return Sub(l, r), nil
	case mult:
		return Mul(l, r), nil
	case pow:
		c, ok := r.(*Int)
		if !ok || c.Sign() < 0 {
			return nil, fmt.Errorf("%s is not supported", node.str)
		}
		if c.IsZero() && l.IsZero() {
			return nil, fmt.Errorf("0^0 is not defined")
		}
		return l.Pow(c), nil
	case div:
		c, ok := r.(NObj)
		if !ok {
			return nil, fmt.Errorf("%s is not supported", node.str)
		}
		if c.IsZero() {
			return nil, fmt.Errorf("divide by zero")
		}
		return l.Div(c), nil
	}
	return nil, fmt.Errorf("%s is not supported", node.str)
}

func (g *Ganrac) evalStackRObj1(stack *pStack, node pNode) (interface{}, error) {
	right, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	r, ok := right.(RObj)
	if !ok {
		return nil, fmt.Errorf("%s is not supported", node.str)
	}
	switch node.cmd {
	case unaryminus:
		return r.Neg(), nil
	}
	return nil, fmt.Errorf("%s is not supported", node.str)
}

func (g *Ganrac) evalStackNvar(stack *pStack, node pNode) (interface{}, error) {
	args := make([]interface{}, node.extra)
	var err error
	for i := len(args) - 1; i >= 0; i-- {
		args[i], err = g.evalStack(stack)
		if err != nil {
			return nil, err
		}
	}

	switch node.cmd {
	case call:
		return g.callFunction(node.str, args)
	case list:
		return NewList(args...), nil
	}
	return nil, fmt.Errorf("%s is not supported", node.str)
}

func (g *Ganrac) evalStackString(stack *pStack, node pNode) (interface{}, error) {
	return NewString(node.str), nil
}

func (g *Ganrac) evalStackAssign(stack *pStack, node pNode) (interface{}, error) {
	v, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	g.varmap[node.str] = v
	return v, nil
}

func (g *Ganrac) evalStackVarDol(stack *pStack, node pNode) (interface{}, error) {
	bi := ParseInt(node.str, 10)
	if !bi.IsInt64() {
		return nil, fmt.Errorf("too large level: %v", bi)
	}
	b := bi.Int64()
	if b > 10000 {
		return nil, fmt.Errorf("too large level: %v", bi)
	}

	lv := Level(b)
	return NewPolyVar(lv), nil
}

func (g *Ganrac) evalStackName(stack *pStack, node pNode) (interface{}, error) {
	v, ok := g.varmap[node.str]
	if !ok {
		return zero, nil
	} else {
		return v, nil
	}
}

func (g *Ganrac) evalStackElem(stack *pStack, node pNode) (interface{}, error) {
	idx, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	pp, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	idxi, ok := idx.(*Int)
	if !ok {
		return nil, fmt.Errorf("index should be integer")
	}

	switch p := pp.(type) {
	case *List:
		return p.Get(idxi)
	case *FmlAnd:
		return p.Get(idxi)
	case *FmlOr:
		return p.Get(idxi)
	case *ForAll:
		return p.Get(idxi)
	case *Exists:
		return p.Get(idxi)
	}
	return nil, fmt.Errorf("index is not supported")
}
