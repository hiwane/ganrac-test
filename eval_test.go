package ganrac

import (
	"strings"
	"testing"
)

func TestEvalRobj(t *testing.T) {

	g := NewGANRAC()

	for i, s := range []struct {
		input  string
		expect RObj
	}{
		{"1+x;", NewPolyCoef(0, 1, 1)},
		{"2+x;", NewPolyCoef(0, 2, 1)},
		{"0;", zero},
		{"1;", one},
		{"1+2;", NewInt(3)},
		{"2*3;", NewInt(6)},
		{"2-5;", NewInt(-3)},
		{"init(x,y,z,t);", nil},
		{"x;", NewPolyCoef(0, 0, 1)},
		{"y;", NewPolyCoef(1, 0, 1)},
		{"z;", NewPolyCoef(2, 0, 1)},
		{"t;", NewPolyCoef(3, 0, 1)},
		{"x+1;", NewPolyCoef(0, 1, 1)},
		{"y+1;", NewPolyCoef(1, 1, 1)},
		{"y+2*3;", NewPolyCoef(1, 6, 1)},
		{"(x+1)+(x+3);", NewPolyCoef(0, 4, 2)},
		{"(x+1)+(3-x);", NewInt(4)},
		{"(x+1)-(+x+1);", zero},
		{"(x+1)+(-x-1);", zero},
		{"(x^2+3*x+1)+(x+5);", NewPolyCoef(0, 6, 4, 1)},
		{"(x^2+3*x+1)+(-3*x+5);", NewPolyCoef(0, 6, 0, 1)},
		{"(x^2+3*x+1)+(-x^2+5*x+8);", NewPolyCoef(0, 9, 8)},
		{"(x^2+3*x+1)+(-x^2-3*x+8);", NewInt(9)},
		{"(x^2+3*x+1)+(-x^2-3*x-1);", zero},
	} {
		u, err := g.Eval(strings.NewReader(s.input))
		if err != nil && s.expect != nil {
			t.Errorf("%d: input=%s: expect=%v, actual=err:%s", i, s.input, s.expect, err)
			break
		}

		if u == s.expect {
			continue
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

func TestEvalCallRObj(t *testing.T) {
	g := NewGANRAC()
	for i, s := range []struct {
		input  string
		expect RObj
	}{
		{"subst(4*x+3);", NewPolyCoef(0, 3, 4)},
		{"subst(x^2+x*y*3+2*y+5,x,3);", NewPolyCoef(1, 14, 11)},
		{"subst(x^2+x*y*3+2*y+5,y,3);", NewPolyCoef(0, 11, 9, 1)},
		{"subst((y-5)*x^3+y*5+x*3+5,y,5);", NewPolyCoef(0, 30, 3)},
		{"subst((y-5)*(x^3+x*3)+8,y,5);", NewInt(8)},
		{"subst((+y-5)*(x^3+x*3+8),y,5);", zero},
		{"subst(5*x+7*y+11*z+3,y,5);", NewPolyCoef(2, NewPolyCoef(0, 38, 5), NewInt(11))},
		{"subst(5*x+7*y+11*z+3,z,5);", NewPolyCoef(1, NewPolyCoef(0, 58, 5), NewInt(7))},
		{"subst(5*x+7*y+11*z+3,x,5);", NewPolyCoef(2, NewPolyCoef(1, 28, 7), NewInt(11))},
		{"subst(5*x+7*y+11*z+3,x,5,x,3);", NewPolyCoef(2, NewPolyCoef(1, 28, 7), NewInt(11))},
		{"subst(5*x+7*y+11*z+3,x,5,y,7);", NewPolyCoef(2, 5*5+7*7+11*00+3, 11)},
		{"subst(5*x+7*y+11*z+3,y,7,x,5);", NewPolyCoef(2, 5*5+7*7+11*00+3, 11)},
		{"subst(5*x+7*y+11*z+3,x,5,z,11);", NewPolyCoef(1, 5*5+7*0+11*11+3, 7)},
		{"subst(5*x+7*y+11*z+3,z,11,x,5);", NewPolyCoef(1, 5*5+7*0+11*11+3, 7)},
		{"subst(5*x+7*y+11*z+3,y,7,z,11);", NewPolyCoef(0, 5*0+7*7+11*11+3, 5)},
		{"subst(5*x+7*y+11*z+3,z,11,y,7);", NewPolyCoef(0, 5*0+7*7+11*11+3, 5)},
		{"subst(5*x+7*y+11*z+3,z,11,y,7,z,3);", NewPolyCoef(0, 5*0+7*7+11*11+3, 5)},
		{"deg(1,y);", zero},
		{"deg(x^2+x+1,x);", NewInt(2)},
		{"deg(x^2+x+1,y);", zero},
		{"deg(x^2+x+x*y+y^3+1,y);", NewInt(3)},
		{"deg(x^2+x+x*y+y^3+1,z);", zero},
		{"deg(x^2+x+x*z+z^3+1,y);", zero},
		{"deg(x^2+x+x*z^5+z^3+1,z);", NewInt(5)},
		{"deg(x^2+x+x*y+y^3+1+y^3*z^3+x^2*z^4+5,z);", NewInt(4)},
		{"coef(3,y,1);", zero},
		{"coef(3,y,0);", NewInt(3)},
		{"coef((x^2+z^2+x*z^3*3)*y^2+((8*z+3)*x^2+4*x+6*z+7),y,1);", zero},
		{"coef((x^2+z^2+x*z^3*3)*y^2+((8*z+3)*x^2+(4)*x+(6*z+7)),y,0);",
			NewPolyCoef(2, NewPolyCoef(0, 7, 4, 3), NewPolyCoef(0, 6, 0, 8))},
		//			NewPolyCoef(0, NewPolyCoef(2, 7, 6), NewInt(4), NewPolyCoef(2, 3, 8))},
		{"coef((x^2+z^2+x*z^3*3)*y^2+((8*z+3)*x^2+(4)*x+(6*z+7)),y,1);", zero},
		{"coef((x^2+z^2+x*z^3*3)*y^2+((8*z+3)*x^2+(4)*x+(6*z+7)),y,2);",
			NewPolyCoef(2, NewPolyCoef(0, 0, 0, 1), zero, one, NewPolyCoef(0, 0, 3))},
		//			NewPolyCoef(0, NewPolyCoef(2, 0, 0, 1), NewPolyCoef(2, 0, 0, 0, 3), one)},
	} {
		u, err := g.Eval(strings.NewReader(s.input))
		if err != nil && s.expect != nil {
			t.Errorf("%d: input=%s: expect=%v, actual=err:%s", i, s.input, s.expect, err)
			break
		}

		g, ok := u.(GObj)
		if !ok {
			t.Errorf("%d: input=%s: I dont know! not gobj: %v", i, s.input, g)
			break
		}

		c, ok := u.(RObj)
		if !ok {
			t.Errorf("%d: input=%s: I dont know! %d:%v", i, s.input, g.Tag(), u)
			break
		}

		if !c.Equals(s.expect) {
			t.Errorf("%d: input=%s: expect=%v, actual(%d)=%v", i, s.input, s.expect, c.Tag(), c)
			break
		}
	}
}

func TestEvalCallFof(t *testing.T) {
	g := NewGANRAC()
	for i, s := range []struct {
		input  string
		expect Fof
	}{
		{"subst(x>0, x,y);", NewAtom(NewPolyCoef(1, 0, 1), GT)},
		{"subst(x>0 && y>0, x,1);", NewAtom(NewPolyCoef(1, 0, 1), GT)},
		{"subst(x>0 && y>0, x,1,y,1);", NewBool(true)},
		{"subst(x>0 && y>0, x,1,y,-1);", NewBool(false)},
	} {
		u, err := g.Eval(strings.NewReader(s.input))
		if err != nil && s.expect != nil {
			t.Errorf("%d: input=%s: expect=%v, actual=err:%s", i, s.input, s.expect, err)
			break
		}

		g, ok := u.(GObj)
		if !ok {
			t.Errorf("%d: input=%s: I dont know! not gobj: %v", i, s.input, g)
			break
		}

		c, ok := u.(Fof)
		if !ok {
			t.Errorf("%d: input=%s: I dont know! %d:%v", i, s.input, g.Tag(), u)
			break
		}

		if !c.Equals(s.expect) {
			t.Errorf("%d: input=%s: expect=%v, actual(%d)=%v", i, s.input, s.expect, c.Tag(), c)
			break
		}
	}
}

func TestEvalFof(t *testing.T) {
	g := NewGANRAC()
	x_gt := NewAtom(NewPolyCoef(0, 0, 1), GT)
	for i, s := range []struct {
		input  string
		expect Fof
	}{
		{"x >= 0;", NewAtom(NewPolyCoef(0, 0, 1), GE)},
		{"x < 1;", NewAtom(NewPolyCoef(0, -1, 1), LT)},
		{"not(x == 1);", NewAtom(NewPolyCoef(0, -1, 1), NE)},
		{"not(2 == 1);", NewBool(true)},
		{"all([x], y > 0);", NewAtom(NewPolyCoef(1, 0, 1), GT)},
		{"all([x], x > 0);", NewQuantifier(true, []Level{0}, x_gt)},
		{"all([x, x, y, x, y], x > 0);", NewQuantifier(true, []Level{0}, x_gt)},
	} {
		u, err := g.Eval(strings.NewReader(s.input))
		if err != nil && s.expect != nil {
			t.Errorf("%d: input=%s: expect=%v, actual=err:%s", i, s.input, s.expect, err)
			break
		}

		g, ok := u.(GObj)
		if !ok {
			t.Errorf("%d: input=%s: I dont know! not gobj: %v", i, s.input, g)
			break
		}

		c, ok := u.(Fof)
		if !ok {
			t.Errorf("%d: input=%s: I dont know! %d:%v", i, s.input, g.Tag(), u)
			break
		}

		if !c.Equals(s.expect) {
			t.Errorf("%d: input=%s: expect=%v, actual(%d)=%v", i, s.input, s.expect, c.Tag(), c)
			break
		}
	}
}
