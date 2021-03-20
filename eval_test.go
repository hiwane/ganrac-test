package ganrac

import (
	"fmt"
	"strings"
	"testing"
)

func TestEval(t *testing.T) {

	for i, s := range []struct {
		input  string
		expect RObj
	}{
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
		{"(x+1)+(x+3);", NewPolyInts(0, 4, 2)},
		{"(x+1)+(3-x);", NewInt(4)},
		{"(x+1)-(+x+1);", NewInt(0)},
		{"(x+1)+(-x-1);", NewInt(0)},
		{"(x^2+3*x+1)+(x+5);", NewPolyInts(0, 6, 4, 1)},
		{"(x^2+3*x+1)+(-3*x+5);", NewPolyInts(0, 6, 0, 1)},
		{"(x^2+3*x+1)+(-x^2+5*x+8);", NewPolyInts(0, 9, 8)},
		{"(x^2+3*x+1)+(-x^2-3*x+8);", NewInt(9)},
		{"(x^2+3*x+1)+(-x^2-3*x-1);", NewInt(0)},
	} {
		fmt.Printf("input=%s\n", s.input)
		u, err := Eval(strings.NewReader(s.input))
		fmt.Printf("output=%v, err=%v\n", u, err)
		if err != nil && s.expect != nil {
			t.Errorf("%d: input=%s: expect=%v, actual=err:%s", i, s.input, s.expect, err)
			break
		}

		c, ok := u.(RObj)
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
