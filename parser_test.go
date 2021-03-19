package ganrac

import (
	"fmt"
	"strings"
	"testing"
)

func test_parse(str string) (*pStack, error) {
	l := new(pLexer)
	l.Init(strings.NewReader(str))
	l.varmap = make(map[string]string)
	stack = new(pStack)
	yyParse(l)
	return stack, nil
}

func test_token2str(t int) string {
	if call <= t && t <= unaryplus {
		return yyToknames[t-call+3]
	}
	return fmt.Sprintf("unknown(%d)", t)
}

func TestToken2Str(t *testing.T) {
	if test_token2str(call) != "call" {
		t.Errorf("invalid call")
	}
	if test_token2str(pow) != "pow" {
		t.Errorf("invalid pow")
	}
	if test_token2str(unaryplus) != "unaryplus" {
		t.Errorf("invalid unaryplus")
	}
}

func TestParseValid(t *testing.T) {
	for k, s := range []struct {
		str   string
		stack []int
	}{
		{"1;", []int{number}},
		{"1+2;", []int{plus, number, number}},
		{"a+2;", []int{plus, number, ident}},
		{"1 > 2;", []int{gtop, number, number}},
		{"A;", []int{name}},
		{"a;", []int{ident}},
		{"a^3 + 1 > 0 && X >= 0;", []int{and, geop, number, name, gtop, number, plus, number, pow, number, ident}},
		{"-x+Y*3 > 0;", []int{gtop, number, plus, mult, number, name, unaryminus, ident}},
		{"[];", []int{list}},
		{"AAA = [];", []int{assign, list, name}},
		{"all([x], Y > x);", []int{call, ident, gtop, ident, name, list, ident}},
	} {
		stack, err := test_parse(s.str)
		if err != nil {
			t.Errorf("[%d]invalid input=\"%s\", err=%s", k, s.str, err.Error())
		}
		for i := 0; !stack.Empty() && i < len(s.stack); i++ {
			v, _ := stack.Pop()
			if v.cmd != s.stack[i] {
				t.Errorf("[%d]invalid input=\"%s\", expect[%d]=%s, actual=%s", k, s.str, i, test_token2str(s.stack[i]), test_token2str(v.cmd))
				goto _next
			}
		}
		if !stack.Empty() {
			t.Errorf("[%d]invalid input=\"%s\", stack is not empty", k, s.str)
		}
	_next:
	}
}
