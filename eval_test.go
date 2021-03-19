package ganrac

import (
	"strings"
	"testing"
	"fmt"
)


func TestEval(t *testing.T) {

	for i, s := range []struct {
		input string
		expect Coef
	} {
		{"1+x;", NewPolyInts(0, 1, 1)},
		{"2+x;", NewPolyInts(0, 2, 1)},
		{"0;", NewInt(0)},
		{"1;", NewInt(1)},
		{"1+2;", NewInt(3)},
		{"2*3;", NewInt(6)},
		{"2-5;", NewInt(-3)},
		{"init(x,y,z,t);", NewInt(0)},
		{"x;", NewPolyInts(0, 0, 1)},
		{"y;", NewPolyInts(1, 0, 1)},
		{"z;", NewPolyInts(2, 0, 1)},
		{"t;", NewPolyInts(3, 0, 1)},
		{"x+1;", NewPolyInts(0, 1, 1)},
		{"y+1;", NewPolyInts(1, 1, 1)},
		{"y+2*3;", NewPolyInts(1, 6, 1)},
	} {
		fmt.Printf("input=%s\n", s.input)
		u, err := Eval(strings.NewReader(s.input))
		fmt.Printf("output=%s\n", u)
		if err != nil && s.expect != nil {
			t.Errorf("%d: input=%s: expect=%v, actual=err:%s", i, s.input, s.expect, err)
			break
		}

		c, ok := u.(Coef)
		if ok {
			if !c.Equals(s.expect) {
				t.Errorf("%d: input=%s: expect=%v, actual(%d)=%v", i, s.input, s.expect, c.Tag(), c)
				break
			}
		} else {
			t.Errorf("%d: input=%s: I dont know!", i, s.input)
			break
		}
	}
}

