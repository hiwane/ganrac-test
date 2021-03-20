package ganrac

import (
	"fmt"
	"io"
)

func Eval(r io.Reader) (interface{}, error) {
	lexer := new(pLexer)
	lexer.Init(r)

	stack = new(pStack)
	yyParse(lexer)

	return evalStack(stack)
}

func evalStack(stack *pStack) (interface{}, error) {
	s, err := stack.Pop()
	if err != nil {
		return nil, err
	}
	switch s.cmd {
	case initvar:
		return evalInitVar(stack, s.extra)
	case plus, minus, mult, pow:
		return evalStackRObj2(stack, s)
	case unaryminus:
		return evalStackRObj1(stack, s)
	case number:
		bi := ParseInt(s.str, 10)
		if bi != nil {
			return bi, nil
		} else {
			return nil, fmt.Errorf("invalid number: %s", s.str)
		}
	case ident:
		lv, err := var2lv(s.str)
		if err != nil {
			return nil, err
		}
		return NewPolyInts(lv, 0, 1), nil
	}
	return nil, fmt.Errorf("unsupported [str=%s, cmd=%d]", s.str, s.cmd)
}

func evalInitVar(stack *pStack, num int) (interface{}, error) {
	vlist := make([]string, num)
	for i := num - 1; i >= 0; i-- {
		p, err := stack.Pop()
		if err != nil {
			return nil, err
		}
		vlist[i] = p.str
	}
	err := InitVar(vlist)
	if err != nil {
		return nil, err
	}
	return NewInt(0), nil
}

func evalStackRObj2(stack *pStack, node pNode) (interface{}, error) {
	right, err := evalStack(stack)
	if err != nil {
		return nil, err
	}
	r, ok := right.(RObj)
	if !ok {
		return nil, fmt.Errorf("%s is not supported", node.str)
	}
	left, err := evalStack(stack)
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
		return l.Pow(c), nil
	}
	return nil, fmt.Errorf("%s is not supported", node.str)
}

func evalStackRObj1(stack *pStack, node pNode) (interface{}, error) {
	right, err := evalStack(stack)
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
