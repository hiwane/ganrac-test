package ganrac

import (
	"testing"
)

func TestAtom(t *testing.T) {
	for _, s := range []struct {
		op1, op2 OP
	}{
		{LE, GT}, // 3, 4
		{LT, GE}, // 1, 6
		{EQ, NE}, // 2, 5
	} {
		pp := NewAtom(NewPolyInts(0, 0, 1), s.op1)
		p, ok := pp.(*Atom)
		if !ok {
			t.Errorf("invalid atom %v", pp)
			return
		}
		qq := pp.Not()
		q, ok := qq.(*Atom)
		if !ok {
			t.Errorf("invalid atom not(%v)=%v", pp, qq)
			return
		}

		if p.p != q.p || q.op != s.op2 {
			t.Errorf("invalid atom not(%v)=%v", pp, qq)
			return
		}

		rr := q.Not()
		r, ok := rr.(*Atom)
		if !ok {
			t.Errorf("invalid atom not(%v)=%v", qq, rr)
			return
		}

		if p.p != r.p || r.op != p.op {
			t.Errorf("invalid atom not(%v)=%v", qq, rr)
			return
		}
	}
}
