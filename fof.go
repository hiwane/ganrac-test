package ganrac

import (
	"fmt"
)

// first-order formula
type Fof interface {
	GObj
	IsQff() bool
	Not() Fof
	Equals(f Fof) bool // 等化まではやらない. 形として同じもの
	hasFreeVar(lv Level) bool
	Subst(xs []RObj, lvs []Level) Fof
	valid() bool // for DEBUG
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

func (p *Atom) valid() bool {
	return p.p.valid() && 1 <= p.op && p.op <= 7
}

func (p *AtomT) valid() bool {
	return true
}
func (p *AtomF) valid() bool {
	return true
}

func (p *FmlAnd) valid() bool {
	if len(p.fml) < 2 {
		return false
	}
	for _, f := range p.fml {
		switch f.(type) {
		case *AtomT, *AtomF, *FmlAnd:
			return false
		}
		if !f.valid() {
			return false
		}
	}
	return true
}

func (p *FmlOr) valid() bool {
	if len(p.fml) < 2 {
		return false
	}
	for _, f := range p.fml {
		switch f.(type) {
		case *AtomT, *AtomF, *FmlOr:
			return false
		}
		if !f.valid() {
			return false
		}
	}
	return true
}

func (p *ForAll) valid() bool {
	if len(p.q) == 0 {
		return false
	}
	for _, lv := range p.q {
		if !p.fml.hasFreeVar(lv) {
			return false
		}
	}
	fml := p.fml
	switch fml.(type) {
	case *AtomT, *AtomF, *ForAll:
		return false
	}
	return p.fml.valid()
}

func (p *Exists) valid() bool {
	if len(p.q) == 0 {
		return false
	}
	for _, lv := range p.q {
		if !p.fml.hasFreeVar(lv) {
			return false
		}
	}
	fml := p.fml
	switch fml.(type) {
	case *AtomT, *AtomF, *Exists:
		return false
	}
	return p.fml.valid()
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
func (p *AtomF) String() string {
	return "false"
}
func (p *Atom) String() string {
	return fmt.Sprintf("%v%s0", p.p, op2str[p.op])
}

func stringFmlAndOr(fmls []Fof, op string) string {
	// @TODO
	return op
}

func (p *FmlAnd) String() string {
	return stringFmlAndOr(p.fml, "&&")
}

func (p *FmlOr) String() string {
	return stringFmlAndOr(p.fml, "||")
}

func stringFmlQ(fmls Fof, q string) string {
	// @TODO
	return q
}

func (p *ForAll) String() string {
	return stringFmlQ(p.fml, "all")
}

func (p *Exists) String() string {
	return stringFmlQ(p.fml, "ex")
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
		}
	case *AtomT:
		return qq
	case *AtomF:
		return pp
	}

	switch q := qq.(type) {
	case *FmlAnd:
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
