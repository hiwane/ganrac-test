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

func test(ox *ganrac.OpenXM) {

	if true {
		ox.ExecFunction("igcd", []interface{} {
			ganrac.NewInt(8), ganrac.NewInt(12)})
		s, _ := ox.PopCMO()
		if s.(*big.Int).Cmp(big.NewInt(4)) != 0 {
			panic("!")
		}
	}

	if true {
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

	test(ox)

	fmt.Println("finished")

	// for _, val := range []int64{5, -3, 132, 102400, -1032} {
	// 	ox.logger.Printf("@@ push ZZ [%d]", val)
	// 	z := big.NewInt(val)
	// 	err = ox.PushCMOZZ(z)
	// 	if err != nil {
	// 		return
	// 	}
	// 	m, _ := ox.PopCMO()
	// 	ox.logger.Printf("@@ get CMO=%x\n", m)
	// }
	if true {
		return
	}

	// cmo0 := []byte{00, 00, 00, 0x14,
	// 	00, 00, 00, 01, 00, 00, 00, 0xc}
	//
	// /* (CMO_ZZ,8) */
	// cmo1 := []byte{00, 00, 00, 0x14,
	// 	00, 00, 00, 01, 00, 00, 00, 8}
	//
	// /* (CMO_INT32,2); */
	// cmo2 := []byte{00, 00, 00, 02, 00, 00, 00, 02}
	//
	// /* (CMO_STRING,"igcd") */
	// cmo3 := []byte{00, 00, 00, 04, 00, 00, 00, 04,
	// 	0x69, 0x67, 0x63, 0x64}
	//
	// ox_push_cmo(dw, cmo0)
	// dw.Flush()
	// ox_pop_string(dw, connd)
	//
	// ox_push_cmo(dw, cmo0)
	// ox_push_cmo(dw, cmo1)
	// ox_push_cmo(dw, cmo2)
	// ox_push_cmo(dw, cmo3)
	// dw.Flush()
	//
	// ox_push_cmd(dw, SM_executeFunction)
	// dw.Flush()
	// s, err := ox_pop_string(dw, connd)
	// fmt.Printf("get %s\n", s)
	//
	// //	ox_push_string(connd, "(x+1)^2;")

}
