package ganrac

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {

	for _, s := range []struct {
		str   string
		token []int
	}{
		{"1", []int{number, -1}},
		{"1 + > 3 >= 3 ;;;   ", []int{number, plus, gtop, number, geop, number, eol, eol, eol, -1}},
		{"[<>>>=>====<>=)", []int{lb, ltop, gtop, gtop, geop, geop, eqop, assign, ltop, geop, rp, -1}},
		{">=1a!=5)", []int{geop, number, ident, neop, number, rp, -1}},
		{"&&||,,/*", []int{and, or, comma, comma, div, mult, -1}},
		{"A B a b true True false FALSE", []int{name, name, ident, ident, f_true, name, f_false, name}},
	} {
		l := new(pLexer)
		l.Init(strings.NewReader(s.str))
		l.varmap = make(map[string]string)
		var lval yySymType

		for j, v := range s.token {
			u := l.Lex(&lval)
			if u != v {
				t.Errorf("input=%s: j=%d: expected=%d, actual=%d", s.str, j, v, u)
			}
		}
	}
}
