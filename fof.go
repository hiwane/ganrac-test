package ganrac

import (
	"fmt"
	"sort"
)

var trueObj = new(AtomT)
var falseObj = new(AtomF)

// first-order formula
type Fof interface {
	GObj
	indeter
	equaler // 等価まではやらない. 形として同じもの
	simpler
	fofTag() uint
	IsQff() bool
	IsQuantifier() bool
	isPrenex() bool
	isEven(lv Level) int // is だが bool 復帰ではない
	redEven(lv Level, v, sgn int) Fof
	Not() Fof
	hasFreeVar(lv Level) bool
	hasVar(lv Level) bool
	maxVar() Level
	numAtom() int
	normalize() Fof
	sotd() int
	vsDeg(lv Level) int // atom の因数分解された多項式の最大次数
	Subst(xs RObj, lv Level) Fof
	replaceVar(xs []RObj, lvs []Level) Fof // xs は *Poly 次数1 主係数1 定数項0 な変数
	valid() error                          // for DEBUG. 実装として適切な形式になっているか
	Deg(lv Level) int
	FmlLess(a Fof) bool
	nonPrenex() Fof
	varShift(lv Level) Fof

	// QE..
	fof_vser      // Virtual Substitution
	fof_quad_eqer // Quadratic Equation
}

type FofQ interface {
	// quantifier
	Fof
	gen(lvv []Level, fml Fof) Fof
	Qs() []Level
	Fml() Fof
	isForAll() bool
}

type FofAO interface {
	// and/or
	Fof
	gen(fml []Fof) Fof
	Fmls() []Fof
	isAnd() bool
}

type FofTFbase struct {
	Fof
}

// AtomT, AtomF, Atom, FmlAnd, FmlOr, ForAll, Exists

type OP uint8

const (
	LT       OP = 0x1
	EQ       OP = 0x2
	GT       OP = 0x4
	LE       OP = LT | EQ
	GE       OP = GT | EQ
	NE       OP = GT | LT
	OP_TRUE  OP = GT | EQ | LT
	OP_FALSE OP = 0x0

	FTAG_TRUE  uint = 0x101
	FTAG_FALSE uint = 0x102
	FTAG_ATOM  uint = 0x103
	FTAG_AND   uint = 0x104
	FTAG_OR    uint = 0x105
	FTAG_ALL   uint = 0x106
	FTAG_EX    uint = 0x107
)

///////////////////////
// Fof define
///////////////////////

type AtomT struct {
	FofTFbase
}

type AtomF struct {
	FofTFbase
}

type Atom struct {
	// p1*p2*...*pn op 0
	p         []*Poly
	pmul      *Poly // = p1*p2*...*p2: cache
	op        OP
	factorizd bool
}

type FmlAnd struct {
	fml []Fof
}

type FmlOr struct {
	fml []Fof
}

type ForAll struct {
	q   []Level // quantifier
	fml Fof
}

type Exists struct {
	q   []Level // quantifier
	fml Fof
}

func (op OP) strict() OP {
	return op & NE
}

func (op OP) neg() OP {
	// 負  :        1 <--> 4, 2 <-->2, 3 <--> 6, 5 <-->5
	if op == EQ || op == NE {
		return op
	} else {
		return op ^ (LT | GT)
	}
}

func (op OP) not() OP {
	// 否定
	return 7 - op
}

func (op OP) String() string {
	return []string{"@false@", "<", "==", "<=", ">", "!=", ">=", "@true@"}[op]
}

///////////////////////
// FofQ
///////////////////////

func (p *ForAll) gen(q []Level, fml Fof) Fof {
	return NewQuantifier(true, q, fml)
}

func (p *Exists) gen(q []Level, fml Fof) Fof {
	return NewQuantifier(false, q, fml)
}

func (p *ForAll) Fml() Fof {
	return p.fml
}

func (p *Exists) Fml() Fof {
	return p.fml
}

func (p *ForAll) Qs() []Level {
	return p.q
}

func (p *Exists) Qs() []Level {
	return p.q
}

func (p *ForAll) isForAll() bool {
	return true
}

func (p *Exists) isForAll() bool {
	return false
}

///////////////////////
// AO / And-Or
///////////////////////

func (p *FmlAnd) Fmls() []Fof {
	return p.fml
}

func (p *FmlOr) Fmls() []Fof {
	return p.fml
}

func (p *FmlAnd) gen(fml []Fof) Fof {
	return newFmlAnds(fml...)
}

func (p *FmlOr) gen(fml []Fof) Fof {
	return newFmlOrs(fml...)
}

func (p *FmlAnd) isAnd() bool {
	return true
}

func (p *FmlOr) isAnd() bool {
	return false
}

///////////////////////
// FOF function
///////////////////////

func (p *Atom) IsQff() bool {
	return true
}

func (p *AtomT) IsQff() bool {
	return true
}

func (p *AtomF) IsQff() bool {
	return true
}

func isQffAndOr(fmls []Fof) bool {
	for _, f := range fmls {
		if !f.IsQff() {
			return false
		}
	}
	return true
}

func (p *FmlAnd) IsQff() bool {
	return isQffAndOr(p.fml)
}

func (p *FmlOr) IsQff() bool {
	return isQffAndOr(p.fml)
}

func (p *ForAll) IsQff() bool {
	return false
}

func (p *Exists) IsQff() bool {
	return false
}

func (p *Atom) isPrenex() bool {
	return true
}

func (p *AtomT) isPrenex() bool {
	return true
}

func (p *AtomF) isPrenex() bool {
	return true
}

func (p *FmlAnd) isPrenex() bool {
	return p.IsQff()
}

func (p *FmlOr) isPrenex() bool {
	return p.IsQff()
}

func (p *ForAll) isPrenex() bool {
	return p.fml.isPrenex()
}

func (p *Exists) isPrenex() bool {
	return p.fml.isPrenex()
}

func (p *Atom) IsQuantifier() bool {
	return false
}

func (p *AtomT) IsQuantifier() bool {
	return false
}

func (p *AtomF) IsQuantifier() bool {
	return false
}

func (p *FmlAnd) IsQuantifier() bool {
	return false
}

func (p *FmlOr) IsQuantifier() bool {
	return false
}

func (p *ForAll) IsQuantifier() bool {
	return true
}

func (p *Exists) IsQuantifier() bool {
	return true
}

func (p *Atom) hasFreeVar(lv Level) bool {
	for _, pp := range p.p {
		if pp.hasVar(lv) {
			return true
		}
	}
	return false
}

func (p *AtomT) hasFreeVar(lv Level) bool {
	return false
}

func (p *AtomF) hasFreeVar(lv Level) bool {
	return false
}

func hasFreeVarFmls(lv Level, fmls []Fof) bool {
	for _, f := range fmls {
		if f.hasFreeVar(lv) {
			return true
		}
	}
	return false
}

func (p *FmlAnd) hasFreeVar(lv Level) bool {
	return hasFreeVarFmls(lv, p.fml)
}

func (p *FmlOr) hasFreeVar(lv Level) bool {
	return hasFreeVarFmls(lv, p.fml)
}

func (p *ForAll) hasFreeVar(lv Level) bool {
	for _, x := range p.q {
		if x == lv {
			return false
		}
	}
	return p.fml.hasFreeVar(lv)
}

func (p *Exists) hasFreeVar(lv Level) bool {
	for _, x := range p.q {
		if x == lv {
			return false
		}
	}
	return p.fml.hasFreeVar(lv)
}

func (p *Atom) hasVar(lv Level) bool {
	for _, pp := range p.p {
		if pp.hasVar(lv) {
			return true
		}
	}
	return false
}

func (p *AtomT) hasVar(lv Level) bool {
	return false
}

func (p *AtomF) hasVar(lv Level) bool {
	return false
}

func hasVarFmls(lv Level, fmls []Fof) bool {
	for _, f := range fmls {
		if f.hasVar(lv) {
			return true
		}
	}
	return false
}

func (p *FmlAnd) hasVar(lv Level) bool {
	return hasVarFmls(lv, p.fml)
}

func (p *FmlOr) hasVar(lv Level) bool {
	return hasVarFmls(lv, p.fml)
}

func (p *ForAll) hasVar(lv Level) bool {
	return p.fml.hasVar(lv)
}

func (p *Exists) hasVar(lv Level) bool {
	return p.fml.hasVar(lv)
}

func (p *Atom) maxVar() Level {
	lv := Level(0)
	for _, pp := range p.p {
		m := pp.maxVar()
		if m > lv {
			lv = m
		}
	}
	return lv
}

func (p *AtomT) maxVar() Level {
	return Level(0)
}

func (p *AtomF) maxVar() Level {
	return Level(0)
}

func (p *FmlAnd) maxVar() Level {
	lv := Level(0)
	for _, f := range p.fml {
		m := f.maxVar()
		if m > lv {
			lv = m
		}
	}
	return lv
}

func (p *FmlOr) maxVar() Level {
	lv := Level(0)
	for _, f := range p.fml {
		m := f.maxVar()
		if m > lv {
			lv = m
		}
	}
	return lv
}

func (p *ForAll) maxVar() Level {
	return p.fml.maxVar()
}

func (p *Exists) maxVar() Level {
	return p.fml.maxVar()
}

func (p *AtomT) numAtom() int {
	return 1
}
func (p *AtomF) numAtom() int {
	return 1
}
func (p *Atom) numAtom() int {
	return 1
}

func (p *FmlAnd) numAtom() int {
	n := 0
	for _, f := range p.fml {
		n += f.numAtom()
	}
	return n
}

func (p *FmlOr) numAtom() int {
	n := 0
	for _, f := range p.fml {
		n += f.numAtom()
	}
	return n
}

func (p *ForAll) numAtom() int {
	return p.fml.numAtom()
}

func (p *Exists) numAtom() int {
	return p.fml.numAtom()
}

func (p *Atom) fofTag() uint {
	return FTAG_ATOM
}

func (p *AtomT) fofTag() uint {
	return FTAG_TRUE
}
func (p *AtomF) fofTag() uint {
	return FTAG_FALSE
}

func (p *FmlAnd) fofTag() uint {
	return FTAG_AND
}

func (p *FmlOr) fofTag() uint {
	return FTAG_OR
}

func (p *ForAll) fofTag() uint {
	return FTAG_ALL
}

func (p *Exists) fofTag() uint {
	return FTAG_EX
}

func (p *Atom) Tag() uint {
	return TAG_FOF
}

func (p *AtomT) Tag() uint {
	return TAG_FOF
}

func (p *AtomF) Tag() uint {
	return TAG_FOF
}

func (p *FmlAnd) Tag() uint {
	return TAG_FOF
}

func (p *FmlOr) Tag() uint {
	return TAG_FOF
}

func (p *ForAll) Tag() uint {
	return TAG_FOF
}

func (p *Exists) Tag() uint {
	return TAG_FOF
}

func (p *Atom) valid() error {
	for i, pp := range p.p {
		err := pp.valid()
		if err != nil {
			return err
		}
		c := pp.LeadinfCoef()
		if c.Sign() <= 0 {
			return fmt.Errorf("%dth lc is poitive: %v", i, p)
		}
	}
	if 1 <= p.op && p.op <= 7 {
		return nil
	} else {
		return fmt.Errorf("op is invalid: %d", p.op)
	}
}

func (p *AtomT) valid() error {
	return nil
}

func (p *AtomF) valid() error {
	return nil
}

func validFmlAndOr(name string, fml []Fof) error {
	if len(fml) < 2 {
		return fmt.Errorf("len(%s) should be greater than 1.", name)
	}
	for _, f := range fml {
		switch f.(type) { // 容易に簡単化できるので許さない
		case *AtomT, *AtomF:
			return fmt.Errorf("%s: invalid element. %v", name, f)
		}
		err := f.valid()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *FmlAnd) valid() error {
	for _, f := range p.fml {
		// and の入れ子は許さない
		switch f.(type) {
		case *FmlAnd:
			return fmt.Errorf("and is in and")
		}
	}
	return validFmlAndOr("and", p.fml)
}

func (p *FmlOr) valid() error {
	for i, f := range p.fml {
		// or の入れ子は許さない
		switch f.(type) {
		case *FmlOr:
			return fmt.Errorf("or is in or [%d] `%v`\n", i, f)
		}
	}
	return validFmlAndOr("or", p.fml)
}

func validQuantifier(name string, q []Level, fml Fof) error {
	if len(q) == 0 {
		return fmt.Errorf("quantifier %s() is empty", name)
	}
	for _, lv := range q {
		// 限量できるのは，子論理式の自由変数のみ
		if !fml.hasFreeVar(lv) {
			return fmt.Errorf("quantifier %s(lv=%d) in redundant", name, lv)
		}
	}
	// 限量子の重複は許さない.. @TODO

	switch fml.(type) {
	case *AtomT, *AtomF:
		return fmt.Errorf("quantifier %s() has boolean", name)
	}

	return fml.valid()
}

func (p *ForAll) valid() error {
	err := validQuantifier("all", p.q, p.fml)
	if err != nil {
		return err
	}
	switch p.fml.(type) {
	case *ForAll:
		return fmt.Errorf("all is in all")
	}
	return nil
}

func (p *Exists) valid() error {
	err := validQuantifier("ex", p.q, p.fml)
	if err != nil {
		return err
	}
	switch p.fml.(type) {
	case *Exists:
		return fmt.Errorf("ex is in ex")
	}
	return nil
}

func (p *Atom) Equals(qq interface{}) bool {
	q, ok := qq.(*Atom)
	if !ok {
		return false
	}
	if p.op == q.op {
		if len(p.p) != len(q.p) {
			return false
		}
		for i := 0; i < len(p.p); i++ {
			if !p.p[i].Equals(q.p[i]) {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

func (p *AtomT) Equals(qq interface{}) bool {
	_, ok := qq.(*AtomT)
	return ok
}

func (p *AtomF) Equals(qq interface{}) bool {
	_, ok := qq.(*AtomF)
	return ok
}

func (p *FmlAnd) Equals(qq interface{}) bool {
	q, ok := qq.(*FmlAnd)
	if !ok {
		return false
	}
	if len(q.fml) != len(p.fml) {
		return false
	}
	for i := 0; i < len(p.fml); i++ {
		if !p.fml[i].Equals(q.fml[i]) {
			return false
		}
	}
	return true
}

func (p *FmlOr) Equals(qq interface{}) bool {
	q, ok := qq.(*FmlOr)
	if !ok {
		return false
	}
	if len(q.fml) != len(p.fml) {
		return false
	}
	for i := 0; i < len(p.fml); i++ {
		if !p.fml[i].Equals(q.fml[i]) {
			return false
		}
	}
	return true
}

func (p *ForAll) Equals(qq interface{}) bool {
	q, ok := qq.(*ForAll)
	if !ok {
		return false
	}
	if len(p.q) != len(q.q) {
		return false
	}
	for _, v := range p.q {
		found := false
		for _, u := range q.q {
			if v == u {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return p.fml.Equals(q.fml)
}

func (p *Exists) Equals(qq interface{}) bool {
	q, ok := qq.(*Exists)
	if !ok {
		return false
	}
	if len(p.q) != len(q.q) {
		return false
	}
	for _, v := range p.q {
		found := false
		for _, u := range q.q {
			if v == u {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return p.fml.Equals(q.fml)
}

func (p *AtomT) String() string {
	return "true"
}

func (p *AtomT) Format(s fmt.State, format rune) {
	switch format {
	case FORMAT_TEX:
		fmt.Fprintf(s, "\\ltrue")
	case FORMAT_DUMP, 'v', 's':
		fmt.Fprintf(s, "%strue%s", esc_sgr(34), esc_sgr(0))
	case FORMAT_SRC:
		fmt.Fprintf(s, "trueObj")
	case FORMAT_QEPCAD:
		fmt.Fprintf(s, "TRUE")
	default:
		p.Format(s, format)
	}
}

func (p *AtomF) String() string {
	return "false"
}

func (p *AtomF) Format(s fmt.State, format rune) {
	switch format {
	case FORMAT_TEX:
		fmt.Fprintf(s, "\\lfalse")
	case FORMAT_DUMP, 'v', 's':
		fmt.Fprintf(s, "%sfalse%s", esc_sgr(34), esc_sgr(0))
	case FORMAT_SRC:
		fmt.Fprintf(s, "falseObj")
	case FORMAT_QEPCAD:
		fmt.Fprintf(s, "FALSE")
	default:
		p.Format(s, format)
	}
}

func (p *Atom) String() string {
	return fmt.Sprintf("%v", p)
}

func (p *Atom) Format(b fmt.State, format rune) {
	switch format {
	case 'v', 's': // 通常コース
		if len(p.p) == 1 {
			p.p[0].Format(b, format)
			fmt.Fprintf(b, "%s0", p.op)
		} else {
			for i, pp := range p.p {
				if i != 0 {
					fmt.Fprintf(b, "*")
				}
				fmt.Fprintf(b, "(%v)", pp)
			}
			fmt.Fprintf(b, "%s0", p.op)
		}
	case FORMAT_TEX: // tex
		if len(p.p) == 1 {
			p.p[0].Format(b, format)
		} else {
			for _, f := range p.p {
				fmt.Fprintf(b, "(")
				f.Format(b, format)
				fmt.Fprintf(b, ")")
			}
		}
		fmt.Fprintf(b, " %s 0", []string{"@false@", "<", "=", "\\leq", ">", "\\neq", "\\ge", "@true@"}[p.op])
	case FORMAT_QEPCAD:
		if len(p.p) == 1 {
			p.p[0].Format(b, format)
		} else {
			for _, f := range p.p {
				fmt.Fprintf(b, "(")
				f.Format(b, format)
				fmt.Fprintf(b, ")")
			}
		}
		fmt.Fprintf(b, " %s 0", []string{"@false@", "<", "=", "<=", ">", "/=", ">=", "@true@"}[p.op])
	case FORMAT_DUMP: // dump
		fmt.Fprintf(b, "(atom %d (", len(p.p))
		for _, pp := range p.p {
			pp.Format(b, format)
		}
		fmt.Fprintf(b, ") %d)", p.op)
	case FORMAT_SRC:

		if len(p.p) == 1 {
			fmt.Fprintf(b, "NewAtom(")
			p.p[0].write_src(b)
		} else {
			fmt.Fprintf(b, "NewAtoms([]RObj{")
			for i, f := range p.p {
				if i != 0 {
					fmt.Fprintf(b, ",")
				}
				f.write_src(b)
			}
			fmt.Fprintf(b, "}")
		}
		fmt.Fprintf(b, ", ")
		switch p.op {
		case LT:
			fmt.Fprintf(b, "LT")
		case EQ:
			fmt.Fprintf(b, "EQ")
		case LE:
			fmt.Fprintf(b, "LE")
		case GT:
			fmt.Fprintf(b, "GT")
		case GE:
			fmt.Fprintf(b, "GE")
		case NE:
			fmt.Fprintf(b, "NE")
		}
		fmt.Fprintf(b, ")")

	default:
		p.Format(b, format)
	}
}

func (p *FmlAnd) Format(b fmt.State, format rune) {
	switch format {
	case 's', 'v': //
		for i, f := range p.fml {
			if i != 0 {
				fmt.Fprintf(b, " && ")
			}

			if _, ok := f.(*FmlOr); ok {
				fmt.Fprintf(b, "%s(%s", esc_sgr(33), esc_sgr(0))
				f.Format(b, format)
				fmt.Fprintf(b, "%s)%s", esc_sgr(33), esc_sgr(0))
			} else {
				f.Format(b, format)
			}
		}
	case FORMAT_TEX: // Tex
		for i, f := range p.fml {
			if i != 0 {
				fmt.Fprintf(b, " \\land ")
			}

			if _, ok := f.(*FmlOr); ok {
				fmt.Fprintf(b, "(")
				f.Format(b, format)
				fmt.Fprintf(b, ")")
			} else {
				f.Format(b, format)
			}
		}
	case FORMAT_QEPCAD:
		for i, f := range p.fml {
			if i == 0 {
				fmt.Fprintf(b, " [ ")
			} else {
				fmt.Fprintf(b, " /\\ ")
			}

			if _, ok := f.(*FmlOr); ok {
				fmt.Fprintf(b, "[ ")
				f.Format(b, format)
				fmt.Fprintf(b, " ]")
			} else {
				f.Format(b, format)
			}
		}
		fmt.Fprintf(b, " ]")
	case FORMAT_DUMP: // dump
		fmt.Fprintf(b, "(&& %d (", len(p.fml))
		for _, f := range p.fml {
			fmt.Fprintf(b, " ")
			f.Format(b, format)
		}
		fmt.Fprintf(b, "))")
	case FORMAT_SRC: // source
		fmt.Fprintf(b, "newFmlAnds(")
		for i, f := range p.fml {
			if i != 0 {
				fmt.Fprintf(b, ",")
			}
			f.Format(b, format)
		}
		fmt.Fprintf(b, ")")

	default:
		p.Format(b, format)
	}
}

func (p *FmlOr) Format(b fmt.State, format rune) {
	switch format {
	case 's', 'v': //
		for i, f := range p.fml {
			if i != 0 {
				fmt.Fprintf(b, " || ")
			}
			// @TODO () は不要だが，明確化のため
			if _, ok := f.(*FmlAnd); ok {
				if coloredFml {
					fmt.Fprintf(b, "%s(%s", esc_sgr(35), esc_sgr(0))
					f.Format(b, format)
					fmt.Fprintf(b, "%s)%s", esc_sgr(35), esc_sgr(0))
				} else {
					f.Format(b, format)
				}
			} else {
				f.Format(b, format)
			}
		}
	case FORMAT_TEX: // Tex
		for i, f := range p.fml {
			if i != 0 {
				fmt.Fprintf(b, " \\lor ")
			}

			f.Format(b, format)
		}
	case FORMAT_QEPCAD:
		for i, f := range p.fml {
			if i == 0 {
				fmt.Fprintf(b, " [ ")
			} else {
				fmt.Fprintf(b, " \\/ ")
			}

			if _, ok := f.(*FmlAnd); ok {
				fmt.Fprintf(b, "[ ")
				f.Format(b, format)
				fmt.Fprintf(b, " ]")
			} else {
				f.Format(b, format)
			}
		}
		fmt.Fprintf(b, " ]")
	case FORMAT_DUMP: // dump
		fmt.Fprintf(b, "(|| %d (", len(p.fml))
		for _, f := range p.fml {
			fmt.Fprintf(b, " ")
			f.Format(b, format)
		}
		fmt.Fprintf(b, "))")
	case FORMAT_SRC: // source
		fmt.Fprintf(b, "newFmlOrs(")
		for i, f := range p.fml {
			if i != 0 {
				fmt.Fprintf(b, ",")
			}
			f.Format(b, format)
		}
		fmt.Fprintf(b, ")")

	default:
		p.Format(b, format)
	}
}

func (p *FmlAnd) String() string {
	return fmt.Sprintf("%v", p)
}

func (p *FmlOr) String() string {
	return fmt.Sprintf("%v", p)
}

func (p *ForAll) Format(b fmt.State, format rune) {
	switch format {
	case 's', 'v': //
		fmt.Fprintf(b, "%sall(%s", esc_sgr(31), esc_sgr(0))
		for i, lv := range p.q {
			if i == 0 {
				fmt.Fprintf(b, "[")
			} else {
				fmt.Fprintf(b, ",")
			}
			fmt.Fprintf(b, "%s", varstr(lv))
		}
		fmt.Fprintf(b, "], ")
		p.fml.Format(b, format)
		fmt.Fprintf(b, "%s)%s", esc_sgr(31), esc_sgr(0))
	case FORMAT_TEX:
		for _, lv := range p.q {
			fmt.Fprintf(b, "\\forall %s ", varstr(lv))
		}
		p.fml.Format(b, format)
	case FORMAT_QEPCAD:
		for _, lv := range p.q {
			fmt.Fprintf(b, "(A %s)", varstr(lv))
		}
		p.fml.Format(b, format)
	case FORMAT_DUMP: // dump
		fmt.Fprintf(b, "(all ")
		for i, lv := range p.q {
			if i == 0 {
				fmt.Fprintf(b, "[")
			} else {
				fmt.Fprintf(b, ",")
			}
			fmt.Fprintf(b, "%d", lv)
		}
		fmt.Fprintf(b, "], ")
		p.fml.Format(b, format)
		fmt.Fprintf(b, ")")
	case FORMAT_SRC: // source
		fmt.Fprintf(b, "NewQuantifier(true, []Level{")
		for i, lv := range p.q {
			if i != 0 {
				fmt.Fprintf(b, ",")
			}
			fmt.Fprintf(b, "%d", lv)
		}
		fmt.Fprintf(b, "}, ")
		p.fml.Format(b, format)
		fmt.Fprintf(b, ")")
	default:
		p.Format(b, format)
	}
}

func (p *Exists) Format(b fmt.State, format rune) {
	switch format {
	case 's', 'v': //
		fmt.Fprintf(b, "%sex(%s", esc_sgr(32), esc_sgr(0))
		for i, lv := range p.q {
			if i == 0 {
				fmt.Fprintf(b, "[")
			} else {
				fmt.Fprintf(b, ",")
			}
			fmt.Fprintf(b, "%s", varstr(lv))
		}
		fmt.Fprintf(b, "], ")
		p.fml.Format(b, format)
		fmt.Fprintf(b, "%s)%s", esc_sgr(32), esc_sgr(0))
	case FORMAT_TEX: // Tex
		for _, lv := range p.q {
			fmt.Fprintf(b, "\\exists %s ", varstr(lv))
		}
		p.fml.Format(b, format)
	case FORMAT_QEPCAD:
		for _, lv := range p.q {
			fmt.Fprintf(b, "(E %s)", varstr(lv))
		}
		p.fml.Format(b, format)
	case FORMAT_DUMP: // dump
		fmt.Fprintf(b, "(ex ")
		for i, lv := range p.q {
			if i == 0 {
				fmt.Fprintf(b, "[")
			} else {
				fmt.Fprintf(b, ",")
			}
			fmt.Fprintf(b, "%d", lv)
		}
		fmt.Fprintf(b, "], ")
		p.fml.Format(b, format)
		fmt.Fprintf(b, ")")
	case FORMAT_SRC: // source
		fmt.Fprintf(b, "NewQuantifier(false, []Level{")
		for i, lv := range p.q {
			if i != 0 {
				fmt.Fprintf(b, ",")
			}
			fmt.Fprintf(b, "%d", lv)
		}
		fmt.Fprintf(b, "}, ")
		p.fml.Format(b, format)
		fmt.Fprintf(b, ")")
	default:
		p.Format(b, format)
	}
}

func (p *ForAll) String() string {
	return fmt.Sprintf("%v", p)
}

func (p *Exists) String() string {
	return fmt.Sprintf("%v", p)
}

func (p *AtomT) Not() Fof {
	return falseObj
}

func (p *AtomF) Not() Fof {
	return trueObj
}

func (p *Atom) Not() Fof {
	return newAtoms(p.p, p.op.not())
}

func (p *FmlAnd) Not() Fof {
	q := new(FmlOr)
	q.fml = make([]Fof, len(p.fml))
	for i := len(p.fml) - 1; i >= 0; i-- {
		q.fml[i] = p.fml[i].Not()
	}
	return q
}

func (p *FmlOr) Not() Fof {
	q := new(FmlAnd)
	q.fml = make([]Fof, len(p.fml))
	for i := len(p.fml) - 1; i >= 0; i-- {
		q.fml[i] = p.fml[i].Not()
	}
	return q
}

func (p *ForAll) Not() Fof {
	q := new(Exists)
	q.q = p.q
	q.fml = p.fml.Not()
	return q
}

func (p *Exists) Not() Fof {
	q := new(ForAll)
	q.q = p.q
	q.fml = p.fml.Not()
	return q
}

func (p *AtomT) Subst(xs RObj, lvs Level) Fof {
	return p
}

func (p *AtomF) Subst(xs RObj, lvs Level) Fof {
	return p
}

func (p *Atom) Subst(xs RObj, lvs Level) Fof {
	op := p.op
	pp := make([]RObj, 0, len(p.p))
	s := 1
	for _, q := range p.p {
		qc := q.Subst(xs, lvs)
		if qc.IsNumeric() {
			s *= qc.Sign()
			if s == 0 {
				return NewAtom(qc, op)
			}
		} else {
			pp = append(pp, qc.(*Poly))
		}
	}
	if len(pp) == 0 {
		return NewAtom(NewInt(int64(s)), op)
	}
	if s < 0 {
		op = op.neg()
	}
	return NewAtoms(pp, op)
}

func (p *FmlAnd) Subst(xs RObj, lvs Level) Fof {
	q := new(FmlAnd)
	q.fml = make([]Fof, 0, len(p.fml))
	for i := 0; i < len(p.fml); i++ {
		fml := p.fml[i].Subst(xs, lvs)
		switch fml.(type) {
		case *AtomT:
			break
		case *AtomF:
			return fml
		default:
			q.fml = append(q.fml, fml)
		}
	}
	if len(q.fml) == 0 {
		return NewBool(true)
	} else if len(q.fml) == 1 {
		return q.fml[0]
	}
	return q
}

func (p *FmlOr) Subst(xs RObj, lvs Level) Fof {
	q := new(FmlOr)
	q.fml = make([]Fof, 0, len(p.fml))
	for i := 0; i < len(p.fml); i++ {
		fml := p.fml[i].Subst(xs, lvs)
		switch fml.(type) {
		case *AtomF:
			break
		case *AtomT:
			return fml
		default:
			q.fml = append(q.fml, fml)
		}
	}
	if len(q.fml) == 0 {
		return NewBool(true)
	} else if len(q.fml) == 1 {
		return q.fml[0]
	}
	return q
}

func substQuantifier(forex bool, fml Fof, qorg []Level, lvs []Level) Fof {
	qq := make([]Level, 0, len(qorg))
	for _, v := range qorg {
		found := false
		for _, u := range lvs {
			if u == v {
				found = true
			}
		}
		if !found {
			qq = append(qq, v)
		}
	}
	return NewQuantifier(forex, qq, fml)
}

func (p *ForAll) Subst(xs RObj, lvs Level) Fof {
	fml := p.fml.Subst(xs, lvs)
	return p.gen(p.q, fml)
}

func (p *Exists) Subst(xs RObj, lvs Level) Fof {
	fml := p.fml.Subst(xs, lvs)
	if err := fml.valid(); err != nil {
		panic(err.Error())
	}
	return p.gen(p.q, fml)
}

func (p *AtomT) replaceVar(xs []RObj, lvs []Level) Fof {
	return p
}

func (p *AtomF) replaceVar(xs []RObj, lvs []Level) Fof {
	return p
}

func (p *Atom) replaceVar(xs []RObj, lvs []Level) Fof {
	pp := make([]RObj, len(p.p))
	for i, q := range p.p {
		pp[i] = q
		for j := 0; j < len(xs); j++ {
			pp[i] = pp[i].Subst(xs[j], lvs[j])
		}
	}
	return NewAtoms(pp, p.op)
}

func (p *FmlAnd) replaceVar(xs []RObj, lvs []Level) Fof {
	q := new(FmlAnd)
	q.fml = make([]Fof, len(p.fml))
	for i := 0; i < len(p.fml); i++ {
		q.fml[i] = p.fml[i].replaceVar(xs, lvs)
	}
	return q
}

func (p *FmlOr) replaceVar(xs []RObj, lvs []Level) Fof {
	q := new(FmlOr)
	q.fml = make([]Fof, len(p.fml))
	for i := 0; i < len(p.fml); i++ {
		q.fml[i] = p.fml[i].replaceVar(xs, lvs)
	}
	return q
}

func (p *ForAll) replaceVar(xs []RObj, lvs []Level) Fof {
	fml := p.fml.replaceVar(xs, lvs)
	qq := make([]Level, len(p.q))
	copy(qq, p.q)
	for i, q := range p.q {
		for j, lv := range lvs {
			if lv == q {
				qq[i] = xs[j].(*Poly).lv
				break
			}
		}
	}
	return NewQuantifier(true, qq, fml)
}

func (p *Exists) replaceVar(xs []RObj, lvs []Level) Fof {
	fml := p.fml.replaceVar(xs, lvs)
	qq := make([]Level, len(p.q))
	copy(qq, p.q)
	for i, q := range p.q {
		for j, lv := range lvs {
			if lv == q {
				qq[i] = xs[j].(*Poly).lv
				break
			}
		}
	}
	return NewQuantifier(false, qq, fml)
}

func NewBool(b bool) Fof {
	if b {
		return trueObj
	} else {
		return falseObj
	}
}

func newAtoms(p []*Poly, op OP) *Atom {
	a := new(Atom)
	a.p = p
	a.op = op
	if op <= 0 || op > 7 {
		panic("invalid op")
	}
	return a
}

func NewAtoms(pp []RObj, op OP) Fof {
	s := 1
	polys := make([]*Poly, 0, len(pp))
	for _, pi := range pp {
		if pi.IsNumeric() {
			sgn := pi.Sign()
			if sgn < 0 {
				s *= -1
			} else if sgn == 0 {
				return NewBool((op & EQ) != 0)
			}
			continue
		}

		// pi is poly
		p := pi.(*Poly)
		if p.Sign() < 0 {
			p = p.Neg().(*Poly)
			s *= -1
		}
		polys = append(polys, p)
	}

	if len(polys) == 0 {
		// 全部 数だった.
		if s < 0 {
			return NewBool((op & LT) != 0)
		} else {
			return NewBool((op & GT) != 0)
		}
	}

	if s < 0 {
		op = op.neg()
	}

	a := new(Atom)
	a.p = polys
	a.op = op
	return a
}

func NewAtom(p RObj, op OP) Fof {
	if p.IsNumeric() {
		s := p.Sign()
		if s < 0 {
			return NewBool((op & 1) != 0)
		} else if s > 0 {
			return NewBool((op & 4) != 0)
		} else {
			return NewBool((op & 2) != 0)
		}
	}
	a := new(Atom)

	a.p = []*Poly{p.(*Poly)}
	if a.p[0].Sign() < 0 { // 正規化. 主係数を正にする.
		a.p[0] = a.p[0].Neg().(*Poly)
		if op != EQ && op != NE {
			a.op = op ^ (LT | GT)
		} else {
			a.op = op
		}
	} else {
		a.op = op
	}

	// 整数化, 原始化 @TODO
	return a
}

func newFmlAnds(pp ...Fof) Fof {
	var q Fof = trueObj
	for _, p := range pp {
		q = NewFmlAnd(q, p)
	}
	return q
}

func newFmlOrs(pp ...Fof) Fof {
	var q Fof = falseObj
	for _, p := range pp {
		q = NewFmlOr(q, p)
	}
	return q
}

func NewFmlAnd(pp Fof, qq Fof) Fof {
	switch p := pp.(type) {
	case *FmlAnd:
		switch q := qq.(type) {
		case *FmlAnd:
			// どっちも and
			r := new(FmlAnd)
			r.fml = make([]Fof, len(p.fml)+len(q.fml))
			copy(r.fml, p.fml)
			for i := 0; i < len(q.fml); i++ {
				r.fml[len(p.fml)+i] = q.fml[i]
			}
			return r
		case *AtomT:
			return pp
		case *AtomF:
			return qq
		default:
			r := new(FmlAnd)
			r.fml = make([]Fof, len(p.fml)+1)
			copy(r.fml, p.fml)
			r.fml[len(p.fml)] = q
			if err := r.valid(); err != nil {
				fmt.Printf("pp=%v\nqq=%v\nrr=%v\n", pp, qq, r)
				panic("stop")
			}
			return r
		}
	case *AtomT:
		return qq
	case *AtomF:
		return pp
	}

	switch q := qq.(type) {
	case *FmlAnd:
		// p は or か q か
		r := new(FmlAnd)
		r.fml = make([]Fof, len(q.fml)+1)
		copy(r.fml[1:], q.fml)
		r.fml[0] = pp
		if err := r.valid(); err != nil {
			fmt.Printf("pp=%v\nqq=%v\nrr=%v\n", pp, qq, r)
			panic("stop")
		}
		return r
	case *AtomT:
		return pp
	case *AtomF:
		return qq
	}

	r := new(FmlAnd)
	r.fml = make([]Fof, 2)
	r.fml[0] = pp
	r.fml[1] = qq
	if err := r.valid(); err != nil {
		fmt.Printf("pp=%v\nqq=%v\nrr=%v\n", pp, qq, r)
		panic("stop")
	}
	return r
}

func NewFmlOr(pp Fof, qq Fof) Fof {
	switch p := pp.(type) {
	case *FmlOr:
		switch q := qq.(type) {
		case *FmlOr:
			// どっちも and
			r := new(FmlOr)
			r.fml = make([]Fof, len(p.fml)+len(q.fml))
			copy(r.fml, p.fml)
			for i := 0; i < len(q.fml); i++ {
				r.fml[len(p.fml)+i] = q.fml[i]
			}
			return r
		case *AtomT:
			return qq
		case *AtomF:
			return pp
		default:
			r := new(FmlOr)
			r.fml = make([]Fof, len(p.fml)+1)
			copy(r.fml, p.fml)
			r.fml[len(p.fml)] = q
			return r
		}
	case *AtomT:
		return pp
	case *AtomF:
		return qq
	}

	switch q := qq.(type) {
	case *FmlOr:
		r := new(FmlOr)
		r.fml = make([]Fof, len(q.fml)+1)
		copy(r.fml[1:], q.fml)
		r.fml[0] = pp
		return r
	case *AtomT:
		return qq
	case *AtomF:
		return pp
	}

	r := new(FmlOr)
	r.fml = make([]Fof, 2)
	r.fml[0] = pp
	r.fml[1] = qq
	return r
}

func NewExists(lv []Level, fml Fof) Fof {
	return NewQuantifier(false, lv, fml)
}

func NewQuantifier(forex bool, lvv []Level, fml Fof) Fof {
	// forex: true -> forall, false -> exists
	lvs := make([]Level, 0)
	if err := fml.valid(); err != nil {
		fmt.Printf("fml invalid %s: %v\n", err, fml)
		panic(err)
	}
	for _, lv := range lvv {
		if fml.hasFreeVar(lv) {
			flag := false
			for _, ll := range lvs { // lvv 内の重複を削除
				if ll == lv {
					flag = true
				}
			}
			if !flag {
				lvs = append(lvs, lv)
			}
		}
	}
	if len(lvs) == 0 {
		return fml
	}
	if forex {
		q := new(ForAll)
		if qq, ok := fml.(*ForAll); ok {
			q.q = append(lvs, qq.q...)
			q.fml = qq.fml
		} else {
			q.q = lvs
			q.fml = fml
		}
		return q
	} else {
		q := new(Exists)
		if qq, ok := fml.(*Exists); ok {
			q.q = append(lvs, qq.q...)
			q.fml = qq.fml
		} else {
			q.q = lvs
			q.fml = fml
		}
		q.q = lvs
		q.fml = fml
		return q
	}
}

func newFmlImplies(f1, f2 Fof) Fof {
	return NewFmlOr(f1.Not(), f2)
}

func newFmlEquiv(f1, f2 Fof) Fof {
	return NewFmlAnd(newFmlImplies(f1, f2), newFmlImplies(f2, f1))
}

func (p *FmlAnd) Len() int {
	return len(p.fml)
}

func (p *FmlOr) Len() int {
	return len(p.fml)
}

func getFmlAndOr(fml []Fof, idx *Int) (Fof, error) {
	if idx.Sign() < 0 || !idx.IsInt64() {
		return nil, fmt.Errorf("index out of range")
	}
	m := idx.Int64()
	if m >= int64(len(fml)) {
		return nil, fmt.Errorf("index out of range")
	}
	for i, f := range fml {
		fmt.Printf("fml[%d]=%v: %d\n", i, f, f.Tag())
	}
	return fml[m], nil
}

func (p *FmlAnd) Get(idx *Int) (Fof, error) {
	return getFmlAndOr(p.fml, idx)
}

func (p *FmlOr) Get(idx *Int) (Fof, error) {
	return getFmlAndOr(p.fml, idx)
}

func getFmlQuantifier(q []Level, fml Fof, idx *Int) (interface{}, error) {
	if idx.IsZero() {
		return q, nil
	}
	if idx.IsOne() {
		return fml, nil
	}

	return nil, fmt.Errorf("invalid index")
}

func (p *ForAll) Get(idx *Int) (interface{}, error) {
	return getFmlQuantifier(p.q, p.fml, idx)
}

func (p *Exists) Get(idx *Int) (interface{}, error) {
	return getFmlQuantifier(p.q, p.fml, idx)
}

func (p *Atom) Indets(b []bool) {
	for _, pp := range p.p {
		pp.Indets(b)
	}
}

func (p *AtomT) Indets(b []bool) {
}

func (p *AtomF) Indets(b []bool) {
}

func (p *FmlAnd) Indets(b []bool) {
	for _, f := range p.fml {
		f.Indets(b)
	}
}

func (p *FmlOr) Indets(b []bool) {
	for _, f := range p.fml {
		f.Indets(b)
	}
}

func (p *ForAll) Indets(b []bool) {
	p.fml.Indets(b)
}

func (p *Exists) Indets(b []bool) {
	p.fml.Indets(b)
}

func (p *Atom) Deg(lv Level) int {
	n := 0
	for _, pp := range p.p {
		n += pp.Deg(lv)
	}
	return n
}

func (p *AtomT) Deg(lv Level) int {
	return 0
}

func (p *AtomF) Deg(lv Level) int {
	return 0
}

func (p *FmlAnd) Deg(lv Level) int {
	m := -1
	for _, f := range p.fml {
		d := f.Deg(lv)
		if d > m {
			m = d
		}
	}
	return m
}

func (p *FmlOr) Deg(lv Level) int {
	m := -1
	for _, f := range p.fml {
		d := f.Deg(lv)
		if d > m {
			m = d
		}
	}
	return m
}

func (p *ForAll) Deg(lv Level) int {
	return p.fml.Deg(lv)
}

func (p *Exists) Deg(lv Level) int {
	return p.fml.Deg(lv)
}

func (p *AtomT) sotd() int {
	return 0
}

func (p *AtomF) sotd() int {
	return 0
}

func (p *Atom) sotd() int {
	m := 0
	for _, f := range p.p {
		m += f.sotd()
	}
	return m
}

func (p *Poly) sotd() int {
	m := 0
	for i, cc := range p.c {
		if c, ok := cc.(*Poly); ok {
			m += i * c.sotd()
		} else {
			m += i
		}
	}
	return m
}

func (p *FmlAnd) sotd() int {
	m := 0
	for _, f := range p.fml {
		m += f.sotd()
	}
	return m
}

func (p *FmlOr) sotd() int {
	m := 0
	for _, f := range p.fml {
		m += f.sotd()
	}
	return m
}

func (p *ForAll) sotd() int {
	return p.fml.sotd()
}

func (p *Exists) sotd() int {
	return p.fml.sotd()
}

func (p *Atom) vsDeg(lv Level) int {
	n := 0
	for _, pp := range p.p {
		d := pp.Deg(lv)
		if d > n {
			n = d
		}
	}
	return n
}

func (p *AtomT) vsDeg(lv Level) int {
	return 0
}

func (p *AtomF) vsDeg(lv Level) int {
	return 0
}

func (p *FmlAnd) vsDeg(lv Level) int {
	m := -1
	for _, f := range p.fml {
		d := f.vsDeg(lv)
		if d > m {
			m = d
		}
	}
	return m
}

func (p *FmlOr) vsDeg(lv Level) int {
	m := -1
	for _, f := range p.fml {
		d := f.vsDeg(lv)
		if d > m {
			m = d
		}
	}
	return m
}

func (p *ForAll) vsDeg(lv Level) int {
	return p.fml.vsDeg(lv)
}

func (p *Exists) vsDeg(lv Level) int {
	return p.fml.vsDeg(lv)
}

func (p *Atom) FmlLess(qq Fof) bool {
	if pt, qt := p.fofTag(), qq.fofTag(); pt != qt {
		return pt < qt
	}
	q := qq.(*Atom)
	if p.op != q.op {
		return p.op < q.op
	}
	if len(p.p) != len(q.p) {
		return len(p.p) < len(q.p)
	}
	for i := 0; i < len(p.p); i++ {
		if p.p[i].lv != q.p[i].lv {
			return p.p[i].lv < q.p[i].lv
		}
	}
	for i := 0; i < len(p.p); i++ {
		if len(p.p[i].c) != len(q.p[i].c) {
			return len(p.p[i].c) < len(q.p[i].c)
		}
	}
	for i := 0; i < len(p.p); i++ {
		if !p.p[i].Equals(q.p[i]) {
			return p.p[i].Less(q.p[i])
		}
	}

	return false
}

func (p *AtomT) FmlLess(q Fof) bool {
	return true
}

func (p *AtomF) FmlLess(q Fof) bool {
	return p.fofTag() < q.fofTag()
}

func (p *FmlAnd) FmlLess(q Fof) bool {
	if p.fofTag() != q.fofTag() {
		return p.fofTag() < q.fofTag()
	}
	return len(p.fml) < len(q.(*FmlAnd).fml)
}

func (p *FmlOr) FmlLess(q Fof) bool {
	if p.fofTag() != q.fofTag() {
		return p.fofTag() < q.fofTag()
	}
	return len(p.fml) < len(q.(*FmlOr).fml)
}

func (p *ForAll) FmlLess(q Fof) bool {
	return p.fofTag() < q.fofTag()
}

func (p *Exists) FmlLess(q Fof) bool {
	return p.fofTag() < q.fofTag()
}

func (p *Atom) nonPrenex() Fof {
	return p
}

func (p *AtomT) nonPrenex() Fof {
	return p
}

func (p *AtomF) nonPrenex() Fof {
	return p
}

func (p *FmlAnd) nonPrenex() Fof {
	fml := make([]Fof, len(p.fml))
	for i, f := range p.fml {
		fml[i] = f.nonPrenex()
	}
	return newFmlAnds(fml...)
}

func (p *FmlOr) nonPrenex() Fof {
	fml := make([]Fof, len(p.fml))
	for i, f := range p.fml {
		fml[i] = f.nonPrenex()
	}
	return newFmlOrs(fml...)
}

func nonPrenexQ(forex bool, qq []Level, sfml Fof) Fof {
	var ofmls []Fof
	switch fml := sfml.(type) {
	case *FmlAnd:
		ofmls = fml.fml
	case *FmlOr:
		ofmls = fml.fml
	default:
		return NewQuantifier(forex, qq, sfml)
	}

	fmli := make([]Fof, 0, len(ofmls))
	fmlo := make([]Fof, 0, len(ofmls))
	for i := range ofmls {
		f := ofmls[i].nonPrenex()
		has := false
		for _, q := range qq {
			if f.hasFreeVar(q) {
				has = true
				break
			}
		}
		if has {
			fmli = append(fmli, f)
		} else {
			fmlo = append(fmlo, f)
		}
	}

	var fmliq Fof
	switch sfml.(type) {
	case *FmlAnd:
		fmliq = newFmlAnds(fmli...)
		fmliq = NewQuantifier(forex, qq, fmliq)
		return NewFmlAnd(fmliq, newFmlAnds(fmlo...))
	case *FmlOr:
		fmliq = newFmlOrs(fmli...)
		fmliq = NewQuantifier(forex, qq, fmliq)
		return NewFmlOr(fmliq, newFmlOrs(fmlo...))
	}
	panic("?")
}

func (p *ForAll) nonPrenex() Fof {
	if fmland, ok := p.fml.(*FmlAnd); ok {
		fml := make([]Fof, len(fmland.fml))
		for i, f := range fmland.fml {
			fml[i] = nonPrenexQ(true, p.q, f)
		}
		return newFmlAnds(fml...)
	} else {
		return nonPrenexQ(true, p.q, p.fml)
	}
}

func (p *Exists) nonPrenex() Fof {
	if fmlor, ok := p.fml.(*FmlOr); ok {
		fml := make([]Fof, len(fmlor.fml))
		for i, f := range fmlor.fml {
			fml[i] = nonPrenexQ(false, p.q, f)
		}
		return newFmlOrs(fml...)
	} else {
		return nonPrenexQ(false, p.q, p.fml)
	}
}

func (p *AtomT) varShift(lv Level) Fof {
	return p
}

func (p *AtomF) varShift(lv Level) Fof {
	return p
}

func (p *Poly) varShift(lv Level) *Poly {
	q := NewPoly(p.lv+lv, len(p.c))
	for i, cc := range p.c {
		switch c := cc.(type) {
		case *Poly:
			q.c[i] = c.varShift(lv)
		default:
			q.c[i] = c
		}
	}
	return q
}

func (p *Atom) varShift(lv Level) Fof {
	q := make([]RObj, len(p.p))
	for i, pp := range p.p {
		q[i] = pp.varShift(lv)
	}
	return NewAtoms(q, p.op)
}

func (p *FmlAnd) varShift(lv Level) Fof {
	q := make([]Fof, len(p.fml))
	for i, f := range p.fml {
		q[i] = f.varShift(lv)
	}
	return newFmlAnds(q...)
}

func (p *FmlOr) varShift(lv Level) Fof {
	q := make([]Fof, len(p.fml))
	for i, f := range p.fml {
		q[i] = f.varShift(lv)
	}
	return newFmlOrs(q...)
}

func (p *ForAll) varShift(lv Level) Fof {
	q := make([]Level, len(p.q))
	for i, qi := range p.q {
		q[i] = qi + lv
	}
	fml := p.fml.varShift(lv)
	return NewQuantifier(true, q, fml)
}

func (p *Exists) varShift(lv Level) Fof {
	q := make([]Level, len(p.q))
	for i, qi := range p.q {
		q[i] = qi + lv
	}
	fml := p.fml.varShift(lv)
	return NewQuantifier(false, q, fml)
}

func (p *Atom) normalize() Fof {
	sort.Slice(p.p, func(i, j int) bool {
		return p.p[i].Less(p.p[j])
	})

	return p
}

func (p *AtomT) normalize() Fof {
	return p
}

func (p *AtomF) normalize() Fof {
	return p
}

func (p *FmlAnd) normalize() Fof {
	f := p.fml
	sort.Slice(f, func(i, j int) bool {
		return f[i].FmlLess(f[j])
	})
	return p
}

func (p *FmlOr) normalize() Fof {
	f := p.fml
	sort.Slice(f, func(i, j int) bool {
		return f[i].FmlLess(f[j])
	})
	return p
}

func (p *ForAll) normalize() Fof {
	p.fml = p.fml.normalize()
	return p
}

func (p *Exists) normalize() Fof {
	p.fml = p.fml.normalize()
	return p
}

func FofImpl(f1, f2 Fof) Fof {
	return NewFmlOr(f1.Not(), f2)
}

func FofEquiv(f1, f2 Fof) Fof {
	return NewFmlAnd(FofImpl(f1, f2), FofImpl(f2, f1))
}

func (a *Atom) getPoly() *Poly {
	if a.pmul == nil {
		var f RObj
		for i, p := range a.p {
			if i == 0 {
				f = p
			} else {
				f = f.Mul(p)
			}
		}
		a.pmul = f.(*Poly)
	}
	return a.pmul
}

func (a *Atom) isUnivariate() bool {
	lv := Level(-1)
	for _, p := range a.p {
		if !p.isUnivariate() {
			return false
		} else if lv < 0 {
			lv = p.lv
		} else if lv != p.lv {
			return false
		}
	}
	return true
}
