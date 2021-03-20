package ganrac

import (
	"errors"
	"fmt"
	"strconv"
	"text/scanner"
)

///////////////////////////////////////////////////////////////
// NODE
///////////////////////////////////////////////////////////////
var debug_print_trace = false

type pNode struct {
	cmd   int
	extra int
	str   string
	pos   scanner.Position
}

func newPNode(str string, cmd, extra int, pos scanner.Position) pNode {
	return pNode{cmd: cmd, str: str, extra: extra, pos: pos}
}

func (n *pNode) String() string {
	return n.str + ":" + strconv.Itoa(n.cmd)
}

///////////////////////////////////////////////////////////////
// STACK
///////////////////////////////////////////////////////////////

type pStack struct {
	v []pNode
}

func (s *pStack) Pop() (pNode, error) {
	if len(s.v) <= 0 {
		return pNode{}, errors.New("empty stack")
	}
	v := s.v[len(s.v)-1]
	s.v = s.v[:len(s.v)-1]
	return v, nil
}

func (s *pStack) Popn(n int) *pStack {
	stack := new(pStack)
	m := len(s.v) - n
	stack.v = append(stack.v, s.v[m:]...)
	s.v = s.v[:m]
	return stack
}

func (s *pStack) Push(v pNode) {
	s.v = append(s.v, v)
}

func (s *pStack) Pushn(v *pStack) {
	s.v = append(s.v, v.v...)
}

func (s *pStack) Empty() bool {
	return len(s.v) == 0
}

func (s *pStack) Len() int {
	return len(s.v)
}

func (s *pStack) String() string {
	ret := ""
	for i := 0; i < len(s.v); i++ {
		if i == 0 {
			ret = s.v[i].String()
		} else {
			ret = ret + " " + s.v[i].String()
		}
	}
	return ret
}

///////////////////////////////////////////////////////////////
// LEXer
///////////////////////////////////////////////////////////////

var stack *pStack

type pLexer struct {
	scanner.Scanner
	s           string
	err         error
	varmap      map[string]string
	print_trace bool
}

type token struct {
	val   string
	label int
}

var sones = []token{
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

var sfuns = []token{
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

func isupper(ch rune) bool {
	return 'A' <= ch && ch <= 'Z'
}
func islower(ch rune) bool {
	return 'a' <= ch && ch <= 'z'
}
func isalpha(ch rune) bool {
	return isupper(ch) || islower(ch)
}
func isdigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
func isalnum(ch rune) bool {
	return isalpha(ch) || isdigit(ch)
}
func isletter(ch rune) bool {
	return isalpha(ch) || ch == '_'
}
func isspace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (l *pLexer) skip_space() {
	for {
		for isspace(l.Peek()) {
			l.Next()
		}
		if l.Peek() != '#' {
			break
		}
		l.Next()
		for l.Peek() != '\n' { // 改行までコメント
			l.Next()
		}
	}
}

// 字句解析機
func (l *pLexer) Lex(lval *yySymType) int {
	l.skip_space()

	c := l.Peek()
	for i, s := range sones { // 記号系
		// if rune(s.val[0]) == c && (len(s.val) == 1 || l.Peek() == rune(s.val[1])) {
		if rune(s.val[0]) == c {
			l.Next()
			c2 := l.Peek()
			for ; i < len(sones); i++ {
				s2 := sones[i]
				if rune(s2.val[0]) == c && (len(s2.val) == 1 || rune(s2.val[1]) == c2) {
					lval.node = newPNode(s2.val, s2.label, 0, l.Pos())
					if len(s2.val) > 1 {
						l.Next()
					}
					return s2.label
				}
			}
			return int(c)
		}
	}

	if isdigit(l.Peek()) { // Integer
		var ret []rune
		for isdigit(l.Peek()) {
			ret = append(ret, l.Next())
		}
		lval.node = newPNode(string(ret), number, 0, l.Pos())
		return number
	}

	if isalpha(l.Peek()) { // 英字
		var ret []rune
		for isdigit(l.Peek()) || isletter(l.Peek()) {
			ret = append(ret, l.Next())
		}
		str := string(ret)
		for i := 0; i < len(sfuns); i++ {
			if str == sfuns[i].val {
				lval.node = newPNode(str, sfuns[i].label, 0, l.Pos())
				return sfuns[i].label
			}
		}

		if isupper(rune(str[0])) {
			lval.node = newPNode(str, name, 0, l.Pos())
		} else {
			lval.node = newPNode(str, ident, 0, l.Pos())
		}
		return lval.node.cmd
	}

	return int(c)
}

func (l *pLexer) Error(s string) {
	pos := l.Pos()
	if l.err == nil {
		l.err = errors.New(fmt.Sprintf("%s:Error:%s \n", pos.String(), s))
	}
}

func yyytrace(s string) {
	if debug_print_trace {
		fmt.Printf(s + "\n")
	}
}

func yyyToken2Str(t int) string {
	if call <= t && t <= unaryplus {
		return yyToknames[t-call+3]
	}
	return fmt.Sprintf("unknown(%d)", t)
}
