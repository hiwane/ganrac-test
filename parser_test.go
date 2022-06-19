package ganrac

import (
	"fmt"
	"strings"
	"testing"
)

func TestToken2Str(t *testing.T) {
	l := new(pLexer)
	if l.token2Str(call) != "call" {
		t.Errorf("invalid call")
	}
	if l.token2Str(pow) != "pow" {
		t.Errorf("invalid pow")
	}
	if l.token2Str(unaryplus) != "unaryplus" {
		t.Errorf("invalid unaryplus")
	}
}

func TestParseValid(t *testing.T) {
	g := NewGANRAC()
	for k, s := range []struct {
		str   string
		stack []int
	}{
		{"{};", []int{eol, dict}},
		{"{a: 1};", []int{eol, dict, ident, number}},
		{"{a: 1, b: \"hoge\"};", []int{eol, dict, ident, t_str, ident, number}},
		{"{\"c\": x^2};", []int{eol, dict, t_str, pow, number, ident}},
		{"a+2;", []int{eol, plus, number, ident}},
		{";", []int{eolq}},
		{":", []int{eolq}},
		{"1:", []int{eolq, number}},
		{"a^3 + 1 < 0 && x >= 0;", []int{eol, and, geop, number, ident, ltop, number, plus, number, pow, number, ident}},
		{"1+x;", []int{eol, plus, ident, number}},
		{"1;", []int{eol, number}},
		{"(1);", []int{eol, number}},
		{"1+2;", []int{eol, plus, number, number}},
		{"1+2*3;", []int{eol, plus, mult, number, number, number}},
		{"(1+2)*3;", []int{eol, mult, number, plus, number, number}},
		{"1 > 2;", []int{eol, gtop, number, number}},
		{init_var_funcname + "(x,t,z);", []int{eol, initvar, ident, ident, ident}},
		{"A;", []int{eol, name}},
		{"a;", []int{eol, ident}},
		{"-x+y*3 > 0;", []int{eol, gtop, number, plus, mult, number, ident, unaryminus, ident}},
		{"AAA = [];", []int{eol, assign, list, name}},
		{"all([x], 3 > x);", []int{eol, call, gtop, ident, number, list, ident}},
		{"A = 0;", []int{eol, assign, number, name}},
		{"func();", []int{eol, call}},
		{"help();", []int{eol, call}},
		{"help(\"all\");", []int{eol, call, t_str}},
	} {
		stack, err := g.parse(strings.NewReader(s.str))
		if err != nil {
			t.Errorf("[%d]invalid\ninput=\"%s\", err=%s", k, s.str, err.Error())
			continue
		}
		m := stack.Len()
		for i := 0; !stack.Empty() && i < len(s.stack); i++ {
			v, _ := stack.Pop()
			if v.cmd != s.stack[i] {
				t.Errorf("[%d, %d] invalid\ninput=\"%s\", expect[%d]=%d, actual=%d", k, m, s.str, i, (s.stack[i]), (v.cmd))
				goto _next
			}
		}
		if !stack.Empty() {
			rr := ""
			for !stack.Empty() {
				v, _ := stack.Pop()
				rr += fmt.Sprintf("\ncmd=%d", v.cmd)
			}

			t.Errorf("[%d,%d]invalid input=\"%s\", stack is not empty. v=%s", k, m, s.str, rr)
		}
	}
_next:
}
