package ganrac

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

var init_var_funcname string = "vars"

type func_table struct {
	name     string
	min, max int
	f        func(g *Ganrac, name string, args []interface{}) (interface{}, error)
	ox       bool
	descript string
	help     string
}

type Ganrac struct {
	varmap             map[string]interface{}
	sones, sfuns       []token
	history            []interface{}
	builtin_func_table []func_table
	ox                 *OpenXM
	logger             *log.Logger
	verbose            int
	verbose_cad        int
}

func NewGANRAC() *Ganrac {
	g := new(Ganrac)
	g.varmap = make(map[string]interface{}, 100)
	g.InitVarList([]string{
		"x", "y", "z", "w", "a", "b", "c", "d", "e", "f", "g", "h",
	})
	g.setBuiltinFuncTable()
	g.logger = log.New(ioutil.Discard, "", 0)
	g.sones = []token{
		{"+", plus},
		{"-", minus},
		{"*", mult},
		{"/", div},
		{"^", pow},
		{"[", lb},
		{"]", rb},
		{"{", lc},
		{"}", rc},
		{"(", lp},
		{")", rp},
		{",", comma},
		{";", eol},
		{":", eolq},
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
		{init_var_funcname, initvar},
		{"time", f_time},
		{"true", f_true},
		{"false", f_false},
	}

	return g
}

func (g *Ganrac) addHisto(o interface{}) {
	g.history = append(g.history, o)
	if len(g.history) > 10 {
		g.history = g.history[len(g.history)-10:]
	}
}

func (g *Ganrac) genLexer(r io.Reader) *pLexer {
	lexer := newLexer(false)
	// yyErrorVerbose = true
	// yyDebug = 5
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

func (g *Ganrac) SetLogger(logger *log.Logger) {
	g.logger = logger
}

func (g *Ganrac) ConnectOX(cw, dw Flusher, cr, dr io.Reader) error {
	g.ox = NewOpenXM(cw, dw, cr, dr, g.logger)
	return g.ox.Init()
}

func (g *Ganrac) log(lv int, format string, a ...interface{}) {
	if lv <= g.verbose {
		fmt.Printf(format, a...)
	}
}

func maxint(a, b int) int {
	if a < b {
		return b
	} else {
		return a
	}
}
