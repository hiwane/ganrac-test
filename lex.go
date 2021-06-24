package ganrac

import (
	"errors"
	"fmt"
	"strconv"
	"text/scanner"
)

var debug_print_trace = false

///////////////////////////////////////////////////////////////
// NODE
///////////////////////////////////////////////////////////////
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
type pLexer struct {
	scanner.Scanner
	s            string
	err          error
	varmap       map[string]string
	print_trace  bool
	stack        *pStack
	sones, sfuns []token
}

func newLexer(trace bool) *pLexer {
	p := new(pLexer)
	p.stack = new(pStack)
	p.print_trace = trace
	return p
}

type token struct {
	val   string
	label int
}

func (p *pLexer) isupper(ch rune) bool {
	return 'A' <= ch && ch <= 'Z'
}
func (p *pLexer) islower(ch rune) bool {
	return 'a' <= ch && ch <= 'z'
}
func (p *pLexer) isalpha(ch rune) bool {
	return p.isupper(ch) || p.islower(ch)
}
func (p *pLexer) isdigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
func (p *pLexer) isalnum(ch rune) bool {
	return p.isalpha(ch) || p.isdigit(ch)
}
func (p *pLexer) isletter(ch rune) bool {
	return p.isalpha(ch) || ch == '_'
}
func (p *pLexer) isspace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (l *pLexer) skip_space() {
	for {
		for l.isspace(l.Peek()) {
			l.Next()
		}
		if l.Peek() != '#' {
			break
		}
		for l.Peek() != '\n' { // 改行までコメント
			l.Next()
		}
	}
}

// 字句解析機
func (l *pLexer) Lex(lval *yySymType) int {
	l.skip_space()

	c := l.Peek()
	for i, s := range l.sones { // 記号系
		// if rune(s.val[0]) == c && (len(s.val) == 1 || l.Peek() == rune(s.val[1])) {
		if rune(s.val[0]) == c {
			l.Next()
			c2 := l.Peek()
			for ; i < len(l.sones); i++ {
				s2 := l.sones[i]
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

	if l.isdigit(l.Peek()) { // Integer
		var ret []rune
		for l.isdigit(l.Peek()) {
			ret = append(ret, l.Next())
		}
		lval.node = newPNode(string(ret), number, 0, l.Pos())
		return number
	}

	if l.isalpha(l.Peek()) { // 英字
		var ret []rune
		for l.isdigit(l.Peek()) || l.isletter(l.Peek()) {
			ret = append(ret, l.Next())
		}
		str := string(ret)
		for i := 0; i < len(l.sfuns); i++ {
			if str == l.sfuns[i].val {
				lval.node = newPNode(str, l.sfuns[i].label, 0, l.Pos())
				return l.sfuns[i].label
			}
		}

		if l.isupper(rune(str[0])) {
			lval.node = newPNode(str, name, 0, l.Pos())
		} else {
			lval.node = newPNode(str, ident, 0, l.Pos())
		}
		return lval.node.cmd
	}
	if l.Peek() == '"' {
		sbuf := make([]rune, 0)
		l.Next()
		for l.Peek() != '"' && l.Peek() != scanner.EOF {
			sbuf = append(sbuf, l.Next())
		}
		l.Next()
		lval.node = newPNode(string(sbuf), t_str, 0, l.Pos())
		return t_str
	}
	if l.Peek() == '$' {
		l.Next()
		var ret []rune
		for l.isdigit(l.Peek()) {
			ret = append(ret, l.Next())
		}
		cmd := vardol
		if len(ret) > 0 {
			lval.node = newPNode(string(ret), cmd, 0, l.Pos())
			return cmd
		}
	} else if l.Peek() == '@' {
		l.Next()
		cmd := varhist
		if l.Peek() == '@' {
			l.Next()
			lval.node = newPNode("1", cmd, 1, l.Pos())
			return cmd
		} else if '1' <= l.Peek() && l.Peek() <= '9' {
			c = l.Next()
			lval.node = newPNode(string(c), cmd, int(c-'0'), l.Pos())
			return cmd
		}
	}

	return int(c)
}

func (l *pLexer) Error(s string) {
	pos := l.Pos()
	if l.err == nil {
		l.err = errors.New(fmt.Sprintf("%s:Error:%s \n", pos.String(), s))
	}
}

func (l *pLexer) push(n pNode) {
	l.stack.Push(n)
}

func (l *pLexer) trace(s string) {
	if l.print_trace {
		fmt.Printf(s + "\n")
	}
}

func (l *pLexer) token2Str(t int) string {
	if call <= t && t <= unaryplus {
		return yyToknames[t-call+3]
	}
	return fmt.Sprintf("unknown(%d)", t)
}

func (l *pLexer) Parse() (*pStack, error) {
	yyParse(l)
	if l.err != nil {
		return nil, l.err
	}
	return l.stack, nil
}
