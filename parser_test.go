package ganrac

import (
	"strings"
	"testing"
)

func test_parse(str string) (*pStack, error) {
	l := new(pLexer)
	l.Init(strings.NewReader(str))
	l.varmap = make(map[string]string)
	return parse(l)
}

func TestToken2Str(t *testing.T) {
	if yyyToken2Str(call) != "call" {
		t.Errorf("invalid call")
	}
	if yyyToken2Str(pow) != "pow" {
		t.Errorf("invalid pow")
	}
	if yyyToken2Str(unaryplus) != "unaryplus" {
		t.Errorf("invalid unaryplus")
	}
}

func TestParseValid(t *testing.T) {
	for k, s := range []struct {
		str   string
		stack []int
	}{
		{"a+2;", []int{plus, number, ident}},
		{"a^3 + 1 < 0 && x >= 0;", []int{and, geop, number, ident, ltop, number, plus, number, pow, number, ident}},
		{"1+x;", []int{plus, ident, number}},
		{"1;", []int{number}},
		{"(1);", []int{number}},
		{"1+2;", []int{plus, number, number}},
		{"1+2*3;", []int{plus, mult, number, number, number}},
		{"(1+2)*3;", []int{mult, number, plus, number, number}},
		{"1 > 2;", []int{gtop, number, number}},
		{"init(x,t,z);", []int{initvar, ident, ident, ident}},
		{"A;", []int{name}},
		{"a;", []int{ident}},
		{"-x+y*3 > 0;", []int{gtop, number, plus, mult, number, ident, unaryminus, ident}},
		{"AAA = [];", []int{assign, list}},
		{"all([x], 3 > x);", []int{call, gtop, ident, number, list, ident}},
		{"A = 0;", []int{assign, number}},
	} {
		stack, err := test_parse(s.str)
		if err != nil {
			t.Errorf("[%d]invalid input=\"%s\", err=%s", k, s.str, err.Error())
			continue
		}
		m := stack.Len()
		for i := 0; !stack.Empty() && i < len(s.stack); i++ {
			v, _ := stack.Pop()
			if v.cmd != s.stack[i] {
				t.Errorf("[%d,%d]invalid input=\"%s\", expect[%d]=%s, actual=%s", k, m, s.str, i, yyyToken2Str(s.stack[i]), yyyToken2Str(v.cmd))
				goto _next
			}
		}
		if !stack.Empty() {
			t.Errorf("[%d,%d]invalid input=\"%s\", stack is not empty", k, m, s.str)
		}
	}
_next:
}
