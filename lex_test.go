package ganrac

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {

	g := NewGANRAC()
	for _, s := range []struct {
		str   string
		token []int
	}{
		{"1", []int{number, -1}},
		{"1+;", []int{number, plus, eol, -1}},
		{"1+:", []int{number, plus, eolq, -1}},
		{":;", []int{eolq, eol, -1}},
		{"1+x;", []int{number, plus, ident, eol, -1}},
		{"1 + > 3 >= 3 ;;;   ", []int{number, plus, gtop, number, geop, number, eol, eol, eol, -1}},
		{"[<>>>=>====<>=)", []int{lb, ltop, gtop, gtop, geop, geop, eqop, assign, ltop, geop, rp, -1}},
		{">=1a!=5)", []int{geop, number, ident, neop, number, rp, -1}},
		{"&&||,,/*", []int{and, or, comma, comma, div, mult, -1}},
		{"A B a b true True false FALSE", []int{name, name, ident, ident, f_true, name, f_false, name}},
		{"x = 0", []int{ident, assign, number}},
		{"\"3\" 3 x", []int{t_str, number, ident}},
		{"\"3;\" 3 x", []int{t_str, number, ident}},
	} {
		l := g.genLexer(strings.NewReader(s.str))
		var lval yySymType

		for j, v := range s.token {
			u := l.Lex(&lval)
			if u != v {
				t.Errorf("input=`%s`: j=%d: expected=%s, actual=%s", s.str, j, l.token2Str(v), l.token2Str(u))
			}
		}
	}
}
