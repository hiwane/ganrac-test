package ganrac

import (
	"fmt"
	"io"
	"strings"
)

// first-order formula
type Fof interface {
	GObj
	Indeter
	IsQff() bool
	Not() Fof
	Equals(f Fof) bool // 等化まではやらない. 形として同じもの
	hasFreeVar(lv Level) bool
	Subst(xs []RObj, lvs []Level) Fof
	valid() error // for DEBUG
	write(b io.Writer)
}

type OP uint8

const (
	LT OP = 0x1
	EQ OP = 0x2
	GT OP = 0x4
	LE OP = LT | EQ
	GE OP = GT | EQ
	NE OP = GT | LT
)

var op2str = []string{"false", "<", "=", "<=", ">", "!=", ">=", "true"}

var trueObj = new(AtomT)
var falseObj = new(AtomF)

type AtomT struct {
}

type AtomF struct {
}

type Atom struct {
	p  *Poly
	op OP
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

func (p *Atom) hasFreeVar(lv Level) bool {
	return p.p.hasVar(lv)
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
	err := p.p.valid()
	if err != nil {
		return err
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
	for _, f := range p.fml {
		// or の入れ子は許さない
		switch f.(type) {
		case *FmlOr:
			return fmt.Errorf("or is in or")
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

func (p *Atom) Equals(qq Fof) bool {
	q, ok := qq.(*Atom)
	if !ok {
		return false
	}
	if p.op == q.op {
		return p.p.Equals(q.p)
	} else if p.op+q.op == 7 {
		return p.p.lv == q.p.lv && len(p.p.c) == len(q.p.c) && p.p.Add(q.p).IsZero()
	} else {
		return false
	}
}

func (p *AtomT) Equals(qq Fof) bool {
	_, ok := qq.(*AtomT)
	return ok
}
func (p *AtomF) Equals(qq Fof) bool {
	_, ok := qq.(*AtomF)
	return ok
}

func (p *FmlAnd) Equals(qq Fof) bool {
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

func (p *FmlOr) Equals(qq Fof) bool {
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

func (p *ForAll) Equals(qq Fof) bool {
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

func (p *Exists) Equals(qq Fof) bool {
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
func (p *AtomT) write(b io.Writer) {
	fmt.Fprintf(b, p.String())
}

func (p *AtomF) String() string {
	return "false"
}
func (p *AtomF) write(b io.Writer) {
	fmt.Fprintf(b, p.String())
}

func (p *Atom) String() string {
	return fmt.Sprintf("%v%s0", p.p, op2str[p.op])
}
func (p *Atom) write(b io.Writer) {
	fmt.Fprintf(b, "%v%s0", p.p, op2str[p.op])
}

func writeFmlAndOr(b io.Writer, fmls []Fof, op string) {
	for i, f := range fmls {
		if i != 0 {
			fmt.Fprintf(b, " %s ", op)
		}
		f.write(b)
	}
}

func (p *FmlAnd) write(b io.Writer) {
	writeFmlAndOr(b, p.fml, "&&")
}

func (p *FmlOr) write(b io.Writer) {
	writeFmlAndOr(b, p.fml, "||")
}

func (p *FmlAnd) String() string {
	var b strings.Builder
	p.write(&b)
	return b.String()
}

func (p *FmlOr) String() string {
	var b strings.Builder
	p.write(&b)
	return b.String()
}

func writeFmlQ(b io.Writer, lvs []Level, fml Fof, q string) {
	fmt.Fprintf(b, "%s(", q)
	for i, lv := range lvs {
		if i == 0 {
			fmt.Fprintf(b, "[")
		} else {
			fmt.Fprintf(b, ",")
		}
		fmt.Fprintf(b, "%s", varlist[lv].v)
	}
	fmt.Fprintf(b, "], ")
	fml.write(b)
	fmt.Fprintf(b, ")")
}

func (p *ForAll) write(b io.Writer) {
	writeFmlQ(b, p.q, p.fml, "all")
}

func (p *Exists) write(b io.Writer) {
	writeFmlQ(b, p.q, p.fml, "ex")
}

func (p *ForAll) String() string {
	var b strings.Builder
	p.write(&b)
	return b.String()
}

func (p *Exists) String() string {
	var b strings.Builder
	p.write(&b)
	return b.String()
}

func (p *AtomT) Not() Fof {
	return falseObj
}

func (p *AtomF) Not() Fof {
	return trueObj
}

func (p *Atom) Not() Fof {
	return NewAtom(p.p, 7-p.op)
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
	q := new(FmlOr)
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

func (p *AtomT) Subst(xs []RObj, lvs []Level) Fof {
	return p
}

func (p *AtomF) Subst(xs []RObj, lvs []Level) Fof {
	return p
}

func (p *Atom) Subst(xs []RObj, lvs []Level) Fof {
	return NewAtom(p.p.Subst(xs, lvs, 0), p.op)
}

func (p *FmlAnd) Subst(xs []RObj, lvs []Level) Fof {
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

func (p *FmlOr) Subst(xs []RObj, lvs []Level) Fof {
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

func (p *ForAll) Subst(xs []RObj, lvs []Level) Fof {
	fml := p.fml.Subst(xs, lvs)
	return substQuantifier(true, fml, p.q, lvs)
}

func (p *Exists) Subst(xs []RObj, lvs []Level) Fof {
	fml := p.fml.Subst(xs, lvs)
	return substQuantifier(false, fml, p.q, lvs)
}

func NewBool(b bool) Fof {
	if b {
		return trueObj
	} else {
		return falseObj
	}
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
	a.p = p.(*Poly)
	if a.p.Sign() < 0 { // 正規化. 主係数を正にする.
		a.p = a.p.Neg().(*Poly)
		a.op = 7 - op
	} else {
		a.op = op
	}
	return a
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
		case *Atom:
			r := new(FmlAnd)
			r.fml = make([]Fof, len(p.fml)+1)
			copy(r.fml, p.fml)
			r.fml[len(p.fml)] = q
			return r
		}
	case *AtomT:
		return qq
	case *AtomF:
		return pp
	case *Atom:
		switch q := qq.(type) {
		case *FmlAnd:
			r := new(FmlAnd)
			r.fml = make([]Fof, len(q.fml)+1)
			copy(r.fml[1:], q.fml)
			r.fml[0] = p
			return r
		}
	}

	switch q := qq.(type) {
	case *FmlAnd:
		// p は or か q か
		r := new(FmlAnd)
		r.fml = make([]Fof, len(q.fml)+1)
		copy(r.fml, q.fml)
		r.fml[len(q.fml)] = pp
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
		case *Atom:
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
	case *Atom:
		switch q := qq.(type) {
		case *FmlOr:
			r := new(FmlOr)
			r.fml = make([]Fof, len(q.fml)+1)
			copy(r.fml[1:], q.fml)
			r.fml[0] = p
			return r
		}
	}

	switch q := qq.(type) {
	case *FmlOr:
		r := new(FmlOr)
		r.fml = make([]Fof, len(q.fml)+1)
		copy(r.fml, q.fml)
		r.fml[len(q.fml)] = pp
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

func NewQuantifier(forex bool, lvv []Level, fml Fof) Fof {
	lvs := make([]Level, 0)
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
	fmt.Printf("idx=%v, m=%d, len=%d\n", idx, m, len(fml))
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
	p.p.Indets(b)
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
