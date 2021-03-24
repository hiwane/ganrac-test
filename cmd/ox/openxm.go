package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	OX_COMMAND   int32 = 513
	OX_DATA      int32 = 514
	OX_SYNC_BALL int32 = 515 /* ball to interrupt */
	OX_NOTIFY    int32 = 516

	OX_DATA_WITH_SIZE              int32 = 521
	OX_DATA_ASIR_BINARY_EXPRESSION int32 = 522 /* This number should be changed*/
	OX_DATA_OPENMATH_XML           int32 = 523
	OX_DATA_OPENMATH_BINARY        int32 = 524
	OX_DATA_MP                     int32 = 525

	SM_popSerializedLocalObject int32 = 258
	SM_popCMO                   int32 = 262
	SM_popString                int32 = 263 /* result ==> string and send the string by CMO */

	SM_mathcap                                    int32 = 264
	SM_pops                                       int32 = 265
	SM_setName                                    int32 = 266
	SM_evalName                                   int32 = 267
	SM_executeStringByLocalParser                 int32 = 268
	SM_executeFunction                            int32 = 269
	SM_beginBlock                                 int32 = 270
	SM_endBlock                                   int32 = 271
	SM_shutdown                                   int32 = 272
	SM_setMathCap                                 int32 = 273
	SM_executeStringByLocalParserInBatchMode      int32 = 274
	SM_getsp                                      int32 = 275
	SM_dupErrors                                  int32 = 276
	SM_pushCMOtag                                 int32 = 277
	SM_executeFunctionAndPopCMO                   int32 = 278
	SM_executeFunctionAndPopSerializedLocalObject int32 = 279
	SM_executeFunctionWithOptionalArgument        int32 = 282

	SM_nop int32 = 300

	SM_control_kill                    int32 = 1024
	SM_control_to_debug_mode           int32 = 1025
	SM_control_exit_debug_mode         int32 = 1026
	SM_control_spawn_server            int32 = 1027
	SM_control_terminate_server        int32 = 1028
	SM_control_reset_connection_server int32 = 1029
	SM_control_reset_connection        int32 = 1030

	SM_PRIVATE int32 = 0x7fff0000 /*  2147418112  */

	LARGEID     int32 = 0x7f000000 /* 2130706432 */
	CMO_PRIVATE int32 = 0x7fff0000 /* 2147418112 */

	CMO_ERROR2  int32 = (LARGEID + 2)
	CMO_NULL    int32 = 1
	CMO_INT32   int32 = 2
	CMO_DATUM   int32 = 3
	CMO_STRING  int32 = 4
	CMO_MATHCAP int32 = 5
	CMO_LIST    int32 = 17

	CMO_ATTRIBUTE_LIST int32 = (LARGEID + 3)

	CMO_MONOMIAL32                 int32 = 19
	CMO_ZZ                         int32 = 20
	CMO_QQ                         int32 = 21
	CMO_ZERO                       int32 = 22
	CMO_DMS                        int32 = 23 /* Distributed monomial system */
	CMO_DMS_GENERIC                int32 = 24
	CMO_DMS_OF_N_VARIABLES         int32 = 25
	CMO_RING_BY_NAME               int32 = 26
	CMO_RECURSIVE_POLYNOMIAL       int32 = 27
	CMO_LIST_R                     int32 = 28
	CMO_INT32COEFF                 int32 = 30
	CMO_DISTRIBUTED_POLYNOMIAL     int32 = 31
	CMO_POLYNOMIAL_IN_ONE_VARIABLE int32 = 33
	CMO_RATIONAL                   int32 = 34
	CMO_COMPLEX                    int32 = 35

	CMO_64BIT_MACHINE_DOUBLE           int32 = 40
	CMO_ARRAY_OF_64BIT_MACHINE_DOUBLE  int32 = 41
	CMO_128BIT_MACHINE_DOUBLE          int32 = 42
	CMO_ARRAY_OF_128BIT_MACHINE_DOUBLE int32 = 43

	CMO_BIGFLOAT          int32 = 50
	CMO_IEEE_DOUBLE_FLOAT int32 = 51
	CMO_BIGFLOAT32        int32 = 52

	CMO_INDETERMINATE int32 = 60
	CMO_TREE          int32 = 61
	CMO_LAMBDA        int32 = 62 /* for function definition */

)

type Flusher interface {
	io.Writer
	Flush() error
}

type OpenXM struct {
	cw, dw Flusher
	cr, dr io.Reader
	serial int32
	border binary.ByteOrder
	logger *log.Logger
}

func NewOpenXM(controlw, dataw Flusher, controlr, datar io.Reader, logger *log.Logger) *OpenXM {
	ox := new(OpenXM)
	ox.cw = controlw
	ox.cr = controlr
	ox.dw = dataw
	ox.dr = datar
	ox.border = binary.LittleEndian
	ox.logger = logger
	return ox
}

func (ox *OpenXM) dataRead(v interface{}) error {
	return binary.Read(ox.dr, ox.border, v)
}

func (ox *OpenXM) dataReadInt32() (int32, error) {
	var n int32
	err := binary.Read(ox.dr, ox.border, &n)
	return n, err
}

func (ox *OpenXM) dataWrite(v interface{}) error {
	if false {
		buf := new(bytes.Buffer)
		binary.Write(buf, ox.border, v)
		b := buf.Bytes()
		fmt.Printf(" ==> ")
		for i := 0; i < len(b); i++ {
			fmt.Printf("%02x ", b[i])
		}
		fmt.Printf("\n")
	}
	return binary.Write(ox.dw, ox.border, v)
}

func (ox *OpenXM) init() error {
	// byte order negotiation
	ox.logger.Printf("ox_init() start")
	b := make([]byte, 1)
	n, err := ox.cr.Read(b)
	if err != nil {
		ox.logger.Printf("<--  read cport %v\n", err)
		return err
	}
	ox.logger.Printf("<--  cread.n=%d: %x\n", n, b[0])
	n, err = ox.dr.Read(b)
	if err != nil {
		ox.logger.Printf("<--  read dport %v\n", err)
		return err
	}
	ox.logger.Printf("<--  dread.n=%d: %x\n", n, b[0])
	b[0] = 1	// Little endian
	n, err = ox.dw.Write(b)
	if err != nil {
		ox.logger.Printf(" --> write cport %v\n", err)
		return err
	}
	ox.dw.Flush()
	n, err = ox.cw.Write(b)
	if err != nil {
		ox.logger.Printf(" --> write cport %v\n", err)
		return err
	}
	ox.cw.Flush()
	ox.logger.Printf(" --> write.<cd>1\n")
	ox.logger.Printf("ox_init() finished")
	return nil
}

func (ox *OpenXM) PushOXCommand(sm_command int32) error {
	var err error
	err = ox.PushOXTag(OX_COMMAND)
	ox.logger.Printf("PushOXCommand(tag:%d): err=%s", sm_command, err)
	if err != nil {
		ox.logger.Printf("push ox command(tag:%d) failed: %s", sm_command, err.Error())
		return err
	}
	err = ox.dataWrite(sm_command)
	if err != nil {
		ox.logger.Printf("push ox command(sm:%d) failed: %s", sm_command, err.Error())

		return err
	}
	return nil
}

func (ox *OpenXM) PushOXTag(tag int32) error {
	var err error
	err = ox.dataWrite(tag)
	if err != nil {
		ox.logger.Printf("PushOXTag(tag:%d,tag) failed: %s", tag, err.Error())
		return err
	}
	ox.serial++
	err = ox.dataWrite(ox.serial)
	if err != nil {
		ox.logger.Printf("PushOXTag(tag:%d, serial) failed: %s", tag, err.Error())
		return err
	}
	ox.logger.Printf(" --> PushOXTag(%d,%d)", tag, ox.serial)
	// no flush
	return nil
}

func (ox *OpenXM) pushCMOTag(cmo int32) error {
	err := ox.PushOXTag(OX_DATA)
	ox.logger.Printf(" --> write CMOTag[%d]: err=%s", cmo, err)
	if err != nil {
		ox.logger.Printf("pushCMOTag(%d,oxtag) failed: %s", cmo, err.Error())
		return err
	}
	err = ox.dataWrite(cmo)
	if err != nil {
		ox.logger.Printf("pushCMOTag(%d,cmotag) failed: %s", cmo, err.Error())
		return err
	}
	return nil
}

func (ox *OpenXM) PushCMO(vv interface{}) error {
	switch v := vv.(type) {
	case int32:
		return ox.PushCMOInt32(v)
	case *big.Int:
		return ox.PushCMOZZ(v)
	case string:
		return ox.PushCMOString(v)
	}

	return fmt.Errorf("PushCMO(): unsupported cmo")
}

func (ox *OpenXM) PushCMOString(s string) error {
	const fname = "PushCMOString"
	err := ox.pushCMOTag(CMO_STRING)
	if err != nil {
		ox.logger.Printf("%s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	b := []byte(s)
	var m int32 = int32(len(b))
	ox.logger.Printf("%s() m=%d, s=%s", fname, m, s)
	err = ox.dataWrite(&m)
	err = ox.dataWrite(b)
	return err
}

func (ox *OpenXM) PushCMOInt32(n int32) error {
	ox.logger.Printf("PushCMOInt32(%d)", n)
	err := ox.pushCMOTag(CMO_INT32)
	if err != nil {
		ox.logger.Printf("PushCMOInt32(%d,cmotag) failed: %s", n, err.Error())
		return err
	}
	err = ox.dataWrite(&n)
	ox.logger.Printf(" --> write CMOint32[%d]", n)
	if err != nil {
		ox.logger.Printf("PushCMOInt32(%d,body) failed: %s", n, err.Error())
		return err
	}
	return nil
}

func (ox *OpenXM) PushCMOZZ(z *big.Int) error {
	const fname = "PushCMOZZ"
	err := ox.pushCMOTag(CMO_ZZ)
	if err != nil {
		ox.logger.Printf("%s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	const intSize = 32 << (^uint(0) >> 63)
	b := z.Bits()
	bb := make([]uint32, 0, len(b) * intSize)
	if intSize == 32 {
		for i := 0; i < len(b); i++ {
			bb = append(bb, uint32(b[i]))
		}
	} else {	// 64
		for i := 0; i < len(b) - 1; i++ {
			bb = append(bb, uint32(b[i]))
			bb = append(bb, uint32((b[i] >> 32) & 0xffffffff))
		}
		i := len(b) - 1
		bb = append(bb, uint32(b[i]))
		vv := uint32((b[i] >> 32) & 0xffffffff)
		if vv != 0 {
			bb = append(bb, uint32(vv))
		}
	}
	b = nil
	m := int32(len(bb))
	if z.Sign() < 0 {
		m *= -1
	}
	err = ox.dataWrite(&m)
	if err != nil {
		ox.logger.Printf("%s(len:%d) failed: %s", fname, m, err.Error())
		return err
	}
	for i := 0; i < len(bb); i++ {
		err = ox.dataWrite(bb[i])
		if err != nil {
			ox.logger.Printf("%s(body:%d/%d:%08x) failed: %s", fname, i, m, b[i], err.Error())
			return err
		}
	}

	return nil
}

func (ox *OpenXM) getCMOInt32() (int32, error) {
	var c int32
	err := ox.dataRead(&c)
	if err != nil {
		ox.logger.Printf("getCMOInt32() failed: %s", err.Error())
	}
	return c, err
}

func (ox *OpenXM) getCMOString() (string, error) {
	n, err := ox.getCMOInt32()
	if err != nil {
		ox.logger.Printf("getCMOString(len) failed: %s", err.Error())
		return "", err
	}
	b := make([]byte, n)
	err = ox.dataRead(b)
	if err != nil {
		ox.logger.Printf("getCMOString(body) failed: %s", err.Error())
		return "", err
	}
	return string(b), err
}

func (ox *OpenXM) getCMOZZ() (*big.Int, error) {
	const fname = "getCMOZZ"
	slen, err := ox.getCMOInt32()
	if err != nil {
		ox.logger.Printf("%s(len) failed: %s", fname, err.Error())
		return nil, err
	}
	m := int(slen)
	if m < 0 {
		m *= -1
	}

	z := new(big.Int)
	for i := 0; i < m; i++ {
		var u uint32
		err := binary.Read(ox.dr, ox.border, &u)
		ox.logger.Printf("%s(body:%d/%d) u=%08x: %d", fname, i, m, u, u)
		if err != nil {
			ox.logger.Printf("%s(body:%d/%d) failed: %s", fname, i, m, err.Error())
		}
		uu := big.NewInt(int64(u))
		uu.Lsh(uu, uint(32 * i))
		z.Add(z, uu)
	}
	if slen < 0 {
		z.Neg(z)
	}
	return z, nil
}

func (ox *OpenXM) getCMO() (interface{}, error) {
	var tag int32
	err := ox.dataRead(&tag)
	if err != nil {
		ox.logger.Printf("getCMO(tag) failed: %s", err.Error())
		return nil, err
	}
	ox.logger.Printf("<--  getCMO() tag=%d", tag)

	switch tag {
	case CMO_ZERO:
		return int32(0), nil
	case CMO_NULL:
		return nil, nil
	case CMO_INT32:
		return ox.getCMOInt32()
	case CMO_STRING:
		return ox.getCMOString()
	case CMO_ZZ:
		return ox.getCMOZZ()
	}

	return 1, nil
}

func (ox *OpenXM) PopOXTag() (int32, int32, error) {
	tag, err := ox.dataReadInt32()
	if err != nil {
		ox.logger.Printf("PopOXTag(oxtag) failed: %s", err.Error())
		return 0, 0, err
	}
	serial, err := ox.dataReadInt32()
	if err != nil {
		ox.logger.Printf("PopOXTag(serial) failed: %s", err.Error())
		return 0, 0, err
	}
	return tag, serial, nil
}

func (ox *OpenXM) PopCMO() (interface{}, error) {
	ox.logger.Printf("PopCMO() start\n")
	err := ox.PushOXCommand(SM_popCMO)
	if err != nil {
		ox.logger.Printf("PopCMO(send-command) failed: %s", err.Error())
		return "", err
	}
	err = ox.dw.Flush()
	if err != nil {
		ox.logger.Printf("PopCMO(flush) failed: %s", err.Error())
		return "", err
	}
	tag, serial, err := ox.PopOXTag()
	ox.logger.Printf("PopCMO() receive: tag=%d, serial=%d\n", tag, serial)
	if tag != OX_DATA {
		return "", fmt.Errorf("PopCMO() unexpected tag=%d", tag)
	}
	v, err := ox.getCMO()
	if err != nil {
		ox.logger.Printf("PopCMO(get) failed: %s", err.Error())
		return "", err
	}
	return v, nil
}

func (ox *OpenXM) PopString() (string, error) {
	ox.logger.Printf("PopString() start\n")
	err := ox.PushOXCommand(SM_popString)
	if err != nil {
		ox.logger.Printf("PopString(send-command) failed: %s", err.Error())
		return "", err
	}
	err = ox.dw.Flush()
	if err != nil {
		ox.logger.Printf("PopString(flush) failed: %s", err.Error())
		return "", err
	}
	tag, serial, err := ox.PopOXTag()
	if err != nil {
		ox.logger.Printf("PopString(oxtag) failed: %s", err.Error())
		return "", err
	}
	if tag != OX_DATA && false {
		ox.logger.Printf("PopString(oxtag) unexpected: tag=%d, serial=%d", tag, serial)
		return "", fmt.Errorf("invalid OX tag")
	}

	v, err := ox.getCMO()
	if err != nil {
		ox.logger.Printf("PopString(get) failed: %s", err.Error())
		return "", err
	}
	return v.(string), err
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

	ox := NewOpenXM(cw, dw, connc, connd, log.New(os.Stderr, "", log.LstdFlags))
	err = ox.init()
	if err != nil {
		return
	}

	// for _, ss := range []string{"hoge dayo!", "こんちわ"} {
	// 	ox.PushCMOString(ss)
	// 	s, _ := ox.PopCMO()
	// 	fmt.Printf("`%s` => `%s`\n", ss, s)
	// }
	// 
	for _, ss := range [][]string{
			{"1+3+4;", "8"},
			{"25 * 150003;", "3750075"},
			{"-5^15;", "-30517578125"},
			{"1234567 * 5443322;", "6720145711574"},
			{"1234567890^3*1234+1532532;", "2321988642787817098346984678532"},
		} {
		ox.PushCMOString(ss[0])
		ox.PushOXCommand(SM_executeStringByLocalParser)
		s, _ := ox.PopCMO()
		sb := s.(*big.Int).String()
		if ss[1] != sb {
			fmt.Printf("[%v]: `%s` => `%v`\n", ss[1] == sb, ss[0], s)
			panic("invalid!")
		}
		ox.PushCMO(s)	// ZZ
		t, _ := ox.PopCMO()
		if s.(*big.Int).Cmp(t.(*big.Int)) != 0 {
			fmt.Printf("[%v]: `%v` => `%v`\n", s.(*big.Int).Cmp(t.(*big.Int)), s, t)
			panic("invalid!")
		}
		ox.PushCMO(s)	// ZZ
		u, _ := ox.PopString()
		if u != ss[1] {
			fmt.Printf("u: `%v` => `%v`\n", s, u)
			panic("invalid!")
		}
	}

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

	ox.logger.Printf("finished\n")
}
