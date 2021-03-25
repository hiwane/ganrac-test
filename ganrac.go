package ganrac

import (
	"io"
)

type func_table struct {
	name     string
	min, max int
	f        func(name string, args []interface{}) (interface{}, error)
	descript string
	help     string
}

type Ganrac struct {
	varmap             map[string]interface{}
	sones, sfuns       []token
	builtin_func_table []func_table
}

func NewGANRAC() *Ganrac {
	g := new(Ganrac)
	g.varmap = make(map[string]interface{}, 100)
	g.setBuiltinFuncTable()
	g.InitVarList([]string{"x", "y", "z", "t"})

	g.sones = []token{
		{"+", plus},
		{"-", minus},
		{"*", mult},
		{"/", div},
		{"^", pow},
		{"[", lb},
		{"]", rb},
		//	{"{", lc},
		//	{"}", rc},
		{"(", lp},
		{")", rp},
		{",", comma},
		{";", eol},
		{"==", eqop},
		{"=", assign},
		{"!=", neop},
		{"<=", leop},
		{"<", ltop},
		{">=", geop},
		{">", gtop},
		{"&&", and},
		{"||", or},
	}

	g.sfuns = []token{
		// {"impl", impl},
		// {"repl", repl},
		// {"equiv", equiv},
		// {"not", not},
		// {"all", all},
		// {"ex", ex},
		{"init", initvar},
		{"true", f_true},
		{"false", f_false},
	}

	return g
}

func (g *Ganrac) genLexer(r io.Reader) *pLexer {
	lexer := newLexer(true)
	lexer.Init(r)
	lexer.sones = g.sones
	lexer.sfuns = g.sfuns
	return lexer
}

func (g *Ganrac) parse(r io.Reader) (*pStack, error) {
	lexer := g.genLexer(r)
	yyParse(lexer)
	if lexer.err != nil {
		return nil, lexer.err
	}
	return lexer.stack, nil
}

func (g *Ganrac) Eval(r io.Reader) (interface{}, error) {
	stack, err := g.parse(r)
	if err != nil {
		return nil, err
	}
	pp, err := g.evalStack(stack)
	if err != nil {
		return nil, err
	}
	if pp == nil {
		return nil, nil
	}
	switch p := pp.(type) {
	case Fof:
		err = p.valid()
		if err != nil {
			return nil, err
		}
	case RObj:
		err = p.valid()
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return pp, nil
}
