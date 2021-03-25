package main

import (
	"bufio"
	"fmt"
	"github.com/hiwane/ganrac"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func test_execfunc(ox *ganrac.OpenXM) {
	ox.ExecFunction("igcd", []interface{}{
		ganrac.NewInt(8), ganrac.NewInt(12)})
	s, _ := ox.PopCMO()
	if s.(*big.Int).Cmp(big.NewInt(4)) != 0 {
		panic("!")
	}
}

func test_list(ox *ganrac.OpenXM) {
	lst := ganrac.NewList([]interface{}{
		ganrac.NewInt(1025),
		ganrac.NewInt(4097),
	})
	ox.PushOxCMO(lst)
	s, _ := ox.PopCMO()

	sl, ok := s.(*ganrac.List)
	if !ok || sl.Len() != 2 {
		panic("why!")
	}
}

func test_rpoly(ox *ganrac.OpenXM) {
	p := ganrac.NewPolyInts(0, 1, 7, 8, 4, 5)
	ox.PushOxCMO(p)
	u, _ := ox.PopString()
	fmt.Printf("%s\n", u)

	p = ganrac.NewPolyCoef(0, ganrac.NewInt(1),
		ganrac.NewPolyInts(1, 71, 17),
		ganrac.NewPolyInts(2, 0, 43),
		ganrac.NewPolyCoef(1, ganrac.NewInt(11), ganrac.NewPolyInts(2, 0, 2)),
		ganrac.NewPolyInts(2, 0, -43),
		ganrac.NewInt(5))

	ox.ExecFunction("fctr", []interface{}{p})
	u, _ = ox.PopString()
	fmt.Printf("fctr1=%s\n", u)

	ox.ExecFunction("fctr", []interface{}{p})
	s, _ := ox.PopCMO()
	// fctr1=[[1,1],[5*x^5+7*x^4+(2*z*y+11)*x^3+(43*z+53)*x^2+(17*y+71)*x+1,1]]
	fmt.Printf("fctr2=%v\n", s)

	p = ganrac.NewPolyInts(1, 4, 0, -1)
	ox.ExecFunction("fctr", []interface{}{p})
	s, _ = ox.PopCMO()
	fmt.Printf("fctr3=%v\n", s)
}

func test(ox *ganrac.OpenXM) {

	test_rpoly(ox)
	test_execfunc(ox)
	test_list(ox)

	for _, ss := range []string{"hoge dayo!", "こんちわ"} {
		ox.PushOxCMO(ss)
		s, _ := ox.PopCMO()
		if s != ss {
			fmt.Printf("`%s` => `%s`\n", ss, s)
			panic("hyyy")
		}
	}
	for _, ss := range [][]string{
		{"1+3+4;", "8"},
		{"25 * 150003;", "3750075"},
		{"-5^15;", "-30517578125"},
		{"1234567 * 5443322;", "6720145711574"},
		{"1234567890^3*1234+1532532;", "2321988642787817098346984678532"},
	} {
		ox.PushOxCMO(ss[0])
		ox.PushOXCommand(ganrac.SM_executeStringByLocalParser)
		s, _ := ox.PopCMO()
		sb := s.(*big.Int).String()
		if ss[1] != sb {
			fmt.Printf("[%v]: `%s` => `%v`\n", ss[1] == sb, ss[0], s)
			panic("invalid!")
		}
		ox.PushOxCMO(s) // ZZ
		t, _ := ox.PopCMO()
		if s.(*big.Int).Cmp(t.(*big.Int)) != 0 {
			fmt.Printf("[%v]: `%v` => `%v`\n", s.(*big.Int).Cmp(t.(*big.Int)), s, t)
			panic("invalid!")
		}
		ox.PushOxCMO(s) // ZZ
		u, _ := ox.PopString()
		if u != ss[1] {
			fmt.Printf("u: `%v` => `%v`\n", s, u)
			panic("invalid!")
		}
	}

}

func main() {
	connc, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Printf("dial1234 %v\n", err)
		return
	}
	defer connc.Close()

	time.Sleep(time.Second * 1)

	connd, err := net.Dial("tcp", "localhost:4321")
	if err != nil {
		fmt.Printf("dial4321 %v\n", err)
		return
	}
	defer connd.Close()

	dw := bufio.NewWriter(connd)
	cw := bufio.NewWriter(connc)

	ox := ganrac.NewOpenXM(cw, dw, connc, connd, log.New(os.Stderr, "", log.LstdFlags))
	err = ox.Init()
	if err != nil {
		return
	}

	ganrac.InitVarList([]string{
		"x", "y", "z", "w", "a", "b", "c", "e", "f", "g", "h",
	})

	test(ox)

	fmt.Println("finished")
}
