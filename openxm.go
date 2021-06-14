package ganrac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/big"
)

// http://www.math.sci.kobe-u.ac.jp/OpenXM/Current/index-spec.html
const (
	OX_COMMAND   int32 = 513 // 0x201
	OX_DATA      int32 = 514 // 0x202
	OX_SYNC_BALL int32 = 515 // ball to interrupt
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
	cw, dw       Flusher
	cr, dr       io.Reader
	serial       int32
	border       binary.ByteOrder
	logger       *log.Logger
	psc_defined  bool
	sres_defined bool
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
		s := " ==>"
		for i := 0; i < len(b); i++ {
			s += fmt.Sprintf(" %02x", b[i])
		}
		ox.logger.Printf("%s ... <%d,%d>", s, len(b), b[0])
	}
	return binary.Write(ox.dw, ox.border, v)
}

func (ox *OpenXM) Init() error {
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
	b[0] = 1 // Little endian
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

	err = ox.ExecFunction("load", "gr")
	if err != nil {
		ox.logger.Printf(" --> load failed %v\n", err)
		return err
	}
	return err
}

func (ox *OpenXM) PushOXCommand(sm_command int32) error {
	var err error
	err = ox.PushOXTag(OX_COMMAND)
	//ox.logger.Printf("PushOXCommand(tag:%d): err=%s", sm_command, err)
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
	// ox.logger.Printf(" --> PushOXTag(%d,%d)", tag, ox.serial)
	// no flush
	return nil
}

func (ox *OpenXM) sendCMOTag(cmo int32) error {
	err := ox.dataWrite(cmo)
	if err != nil {
		ox.logger.Printf("sendCMOTag(%d,cmotag) failed: %s", cmo, err.Error())
		return err
	}
	return nil
}

func (ox *OpenXM) PushOxCMO(vv interface{}) error {
	const fname = "PushOxCMO"
	err := ox.PushOXTag(OX_DATA)
	// ox.logger.Printf(" --> write CMOTag[%d]: err=%s", cmo, err)
	if err != nil {
		ox.logger.Printf("%s(oxtag) failed: %s", fname, err.Error())
		return err
	}

	err = ox.sendCMO(vv, nil)
	if err != nil {
		ox.logger.Printf("%s(cmo) failed: %s", fname, err.Error())
		return err

	}
	return err
}

func (ox *OpenXM) sendCMO(vv interface{}, lvmap map[Level]int32) error {
	// Remarks: OX_tag is already sent.
	const fname = "sendCMO"
	switch v := vv.(type) {
	case int32:
		if v == 0 {
			return ox.sendCMOZero()
		} else {
			return ox.sendCMOInt32(v)
		}
	case *big.Int:
		if v.Sign() == 0 {
			return ox.sendCMOZero()
		} else {
			return ox.sendCMOZZ(v)
		}
	case *big.Rat:
		return ox.sendCMOQQ(v)
	case *Int:
		if v.Sign() == 0 {
			return ox.sendCMOZero()
		} else {
			return ox.sendCMOZZ(v.n)
		}
	case *Rat:
		return ox.sendCMOQQ(v.n)
	case string:
		return ox.sendCMOString(v)
	case *String:
		return ox.sendCMOString(v.s)
	case *List:
		return ox.sendCMOList(v)
	case *Poly:
		if lvmap == nil {
			var err error
			lvmap, err = ox.sendCMORecPoly(v)
			if err != nil {
				ox.logger.Printf(" --> %s(recpoly) failed: %s", fname, err.Error())
				return err
			}
		}

		return ox.sendCMOPoly(v, lvmap)
	}

	return fmt.Errorf(" --> %s(): unsupported cmo %v", fname, vv)
}

func (ox *OpenXM) sendCMOList(v *List) error {
	const fname = "sendCMOList"
	ox.logger.Printf(" --> %s(cmotag) start", fname)
	err := ox.sendCMOTag(CMO_LIST)
	if err != nil {
		ox.logger.Printf(" --> %s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	var m int32 = int32(v.Len())
	err = ox.dataWrite(&m)
	if err != nil {
		ox.logger.Printf(" --> %s(len:%d) failed: %s", fname, m, err.Error())
		return err
	}
	for i := 0; i < v.Len(); i++ {
		o, _ := v.Geti(i)
		ox.logger.Printf(" --> %s(cmotag) v[%d]=%v", fname, i, o)
		err := ox.sendCMO(o, nil)
		if err != nil {
			ox.logger.Printf(" --> %s(%d) failed: %s", fname, i, err.Error())
			return err
		}
	}
	return nil
}

func (ox *OpenXM) sendCMORecPoly(p *Poly) (map[Level]int32, error) {
	const fname = "sendCMORecPoly"
	err := ox.sendCMOTag(CMO_RECURSIVE_POLYNOMIAL)
	if err != nil {
		return nil, err
	}

	b := make([]bool, len(varlist))
	p.Indets(b)

	var cnt int32 = 0
	lvmap := make(map[Level]int32, len(varlist))
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] {
			lvmap[Level(i)] = cnt
			cnt++
		}
	}
	ox.logger.Printf(" --> %s() cnt=%d\n", fname, cnt)
	ox.sendCMOTag(CMO_LIST)
	err = ox.dataWrite(&cnt)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] {
			ox.sendCMOString(varstr(Level(i)))
			cnt--
		}
	}
	return lvmap, err
}

func (ox *OpenXM) sendCMOPoly(p *Poly, lvmap map[Level]int32) error {
	// すでに sendCMORecPoly() は呼ばれている
	err := ox.sendCMOTag(CMO_POLYNOMIAL_IN_ONE_VARIABLE)
	if err != nil {
		return err
	}

	var cnt int32 = 1 // 非ゼロ項数. 主係数非ゼロは確定
	for i := 0; i < len(p.c)-1; i++ {
		if !p.c[i].IsZero() {
			cnt++
		}
	}
	ox.logger.Printf(" --> sendCMOPoly() #mono=%d, p.lv=%d -> %d\n", cnt, p.lv, lvmap[p.lv])
	if err := p.valid(); err != nil {
		panic(fmt.Sprintf("sendcmo: #mono=%d, p.lv=%d\nerr=%v\np=%v", cnt, p.lv, err, p))
	}
	err = ox.dataWrite(&cnt)
	err = ox.dataWrite(lvmap[p.lv])
	for i := int32(len(p.c) - 1); cnt > 0; i-- {
		if !p.c[i].IsZero() {
			cnt--
			err = ox.dataWrite(&i)
			ox.sendCMO(p.c[i], lvmap)
		}
	}
	return err
}

func (ox *OpenXM) sendCMOString(s string) error {
	const fname = "sendCMOString"
	err := ox.sendCMOTag(CMO_STRING)
	if err != nil {
		ox.logger.Printf(" --> %s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	b := []byte(s)
	var m int32 = int32(len(b))
	// ox.logger.Printf(" --> %s() m=%d, s=%s", fname, m, s)
	err = ox.dataWrite(&m)
	err = ox.dataWrite(b)
	return err
}

func (ox *OpenXM) sendCMOZero() error {
	const fname = "sendCMOZero"
	err := ox.sendCMOTag(CMO_ZERO)
	if err != nil {
		ox.logger.Printf(" --> %s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	return nil
}

func (ox *OpenXM) sendCMOInt32(n int32) error {
	// ox.logger.Printf("PushCMOInt32(%d)", n)
	const fname = "sendCMOInt32"
	err := ox.sendCMOTag(CMO_INT32)
	if err != nil {
		ox.logger.Printf(" --> %s(%d,cmotag) failed: %s", fname, n, err.Error())
		return err
	}
	err = ox.dataWrite(&n)
	// ox.logger.Printf(" --> write CMOint32[%d]", n)
	if err != nil {
		ox.logger.Printf(" --> %s(%d,body) failed: %s", fname, n, err.Error())
		return err
	}
	return nil
}

func (ox *OpenXM) send_bigint(z *big.Int) error {
	const fname = "send_bigint"
	const intSize = 32 << (^uint(0) >> 63)
	b := z.Bits()
	bb := make([]uint32, 0, len(b)*intSize)
	if intSize == 32 {
		for i := 0; i < len(b); i++ {
			bb = append(bb, uint32(b[i]))
		}
	} else { // 64
		for i := 0; i < len(b)-1; i++ {
			bb = append(bb, uint32(b[i]))
			bb = append(bb, uint32((b[i]>>32)&0xffffffff))
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
	err := ox.dataWrite(&m)
	if err != nil {
		ox.logger.Printf(" --> %s(len:%d) failed: %s", fname, m, err.Error())
		return err
	}
	for i := 0; i < len(bb); i++ {
		err = ox.dataWrite(bb[i])
		if err != nil {
			ox.logger.Printf(" --> %s(body:%d/%d:%08x) failed: %s", fname, i, m, b[i], err.Error())
			return err
		}
	}
	return nil
}

func (ox *OpenXM) sendCMOQQ(z *big.Rat) error {
	const fname = "sendCMOQQ"
	err := ox.sendCMOTag(CMO_QQ)
	if err != nil {
		ox.logger.Printf(" --> %s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	ox.send_bigint(z.Num())
	return ox.send_bigint(z.Denom())
}

func (ox *OpenXM) sendCMOZZ(z *big.Int) error {
	const fname = "sendCMOZZ"
	err := ox.sendCMOTag(CMO_ZZ)
	if err != nil {
		ox.logger.Printf(" --> %s(cmotag) failed: %s", fname, err.Error())
		return err
	}
	return ox.send_bigint(z)
}

func (ox *OpenXM) recvCMOInt32() (int32, error) {
	var c int32
	err := ox.dataRead(&c)
	if err != nil {
		ox.logger.Printf("recvCMOInt32() failed: %s", err.Error())
	}
	return c, err
}

func (ox *OpenXM) recvCMOIndeterminate() (*Poly, error) {
	const fname = "recvCMOIndeterminate"
	cc, err := ox.recvCMO(nil)
	if err != nil {
		return nil, err
	}
	c := cc.(string)

	lv, ok := varstr2lv[c]
	if ok {
		return varlist[lv].p, nil
	}

	return nil, fmt.Errorf("unknown variable %s", c)
}

func (ox *OpenXM) toGObj(p interface{}) GObj {
	switch cc := p.(type) {
	case GObj:
		return cc
	case int32:
		return NewInt(int64(cc))
	case *big.Int:
		bb := newInt()
		bb.n.Set(cc)
		return bb
	case *big.Rat:
		bb := newRat()
		bb.n.Set(cc)
		return bb
	case string:
		return NewString(cc)
	case int, int64:
		panic(fmt.Sprintf("unsupported intx `%v`", p))
	default:
		panic(fmt.Sprintf("unsupported `%v`", p))
	}
}

func (ox *OpenXM) recvCMOPoly1Var(ringdef *List) (*Poly, error) {
	m, err := ox.dataReadInt32()
	if err != nil {
		return nil, err
	}
	lv, err := ox.dataReadInt32()
	if err != nil {
		return nil, err
	}
	p, _ := ringdef.Geti(int(lv))
	plv := p.(*Poly).lv

	// asir, openxm とは再帰表現での保持方法が逆.
	// LV の昇順→降順
	var ret RObj = zero
	for m > 0 {
		m--
		exp, _ := ox.dataReadInt32()
		coef, _ := ox.recvCMO(ringdef)
		c, ok := ox.toGObj(coef).(RObj)
		if ok {
			if exp == 0 {
				ret = Add(ret, c)
			} else {
				ret = Add(ret, Mul(c, newPolyVarn(plv, int(exp))))
			}
		}
	}
	return ret.(*Poly), err
}
func (ox *OpenXM) recvCMORPoly() (*Poly, error) {
	ringdef, err := ox.recvCMO(nil)
	if err != nil {
		return nil, err
	}
	ox.logger.Printf(" ..  ringdef=%v", ringdef)
	coef, err := ox.recvCMO(ringdef.(*List))
	if err != nil {
		return nil, err
	}
	return coef.(*Poly), nil
}

func (ox *OpenXM) recvCMOList() (*List, error) {
	const fname = "recvCMOList"
	var m int32
	err := ox.dataRead(&m)
	// ox.logger.Printf("recvCMOList() m=%d\n", m)
	if err != nil {
		ox.logger.Printf("%s(len) failed: %s", fname, err.Error())
		return nil, err
	}

	ret := NewList()
	for i := int(m); i > 0; i-- {
		o, err := ox.recvCMO(nil)
		if err != nil {
			ox.logger.Printf("%s(%d) failed: %s", fname, int(m)-i, err.Error())
			return nil, err
		}
		ret.Append(ox.toGObj(o))
	}

	return ret, nil
}

func (ox *OpenXM) recvCMOString() (string, error) {
	n, err := ox.recvCMOInt32()
	if err != nil {
		ox.logger.Printf("recvCMOString(len) failed: %s", err.Error())
		return "", err
	}
	b := make([]byte, n)
	err = ox.dataRead(b)
	if err != nil {
		ox.logger.Printf("recvCMOString(body) failed: %s", err.Error())
		return "", err
	}
	return string(b), err
}

func (ox *OpenXM) recvCMOQQ() (*big.Rat, error) {
	const fname = "recvCMOQQ"
	num, _ := ox.recvCMOZZ()
	den, _ := ox.recvCMOZZ()
	q := big.NewRat(0, 1)
	q.SetFrac(num, den)
	return q, nil
}

func (ox *OpenXM) recvCMOZZ() (*big.Int, error) {
	const fname = "recvCMOZZ"
	slen, err := ox.recvCMOInt32()
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
		if err != nil {
			ox.logger.Printf("%s(body:%d/%d) failed: %s", fname, i, m, err.Error())
		}
		uu := big.NewInt(int64(u))
		uu.Lsh(uu, uint(32*i))
		z.Add(z, uu)
	}
	if slen < 0 {
		z.Neg(z)
	}
	return z, nil
}

func (ox *OpenXM) cmoTagString(tag int32) string {
	switch tag {
	case CMO_ERROR2:
		return "CMO_ERROR2"
	case CMO_NULL:
		return "CMO_NULL"
	case CMO_INT32:
		return "CMO_INT32"
	case CMO_DATUM:
		return "CMO_DATUM"
	case CMO_STRING:
		return "CMO_STRING"
	case CMO_MATHCAP:
		return "CMO_MATHCAP"
	case CMO_LIST:
		return "CMO_LIST"

	case CMO_ATTRIBUTE_LIST:
		return "CMO_ATTRIBUTE_LIST"

	case CMO_MONOMIAL32:
		return "CMO_MONOMIAL32"
	case CMO_ZZ:
		return "CMO_ZZ"
	case CMO_QQ:
		return "CMO_QQ"
	case CMO_ZERO:
		return "CMO_ZERO"
	case CMO_DMS:
		return "CMO_DMS"
	case CMO_DMS_GENERIC:
		return "CMO_DMS_GENERIC"
	case CMO_DMS_OF_N_VARIABLES:
		return "CMO_DMS_OF_N_VARIABLES"
	case CMO_RING_BY_NAME:
		return "CMO_RING_BY_NAME"
	case CMO_RECURSIVE_POLYNOMIAL:
		return "CMO_RECURSIVE_POLYNOMIAL"
	case CMO_LIST_R:
		return "CMO_LIST_R"
	case CMO_INT32COEFF:
		return "CMO_INT32COEFF"
	case CMO_DISTRIBUTED_POLYNOMIAL:
		return "CMO_DISTRIBUTED_POLYNOMIAL"
	case CMO_POLYNOMIAL_IN_ONE_VARIABLE:
		return "CMO_POLYNOMIAL_IN_ONE_VARIABLE"
	case CMO_RATIONAL:
		return "CMO_RATIONAL"
	case CMO_COMPLEX:
		return "CMO_COMPLEX"

	case CMO_64BIT_MACHINE_DOUBLE:
		return "CMO_64BIT_MACHINE_DOUBLE"
	case CMO_ARRAY_OF_64BIT_MACHINE_DOUBLE:
		return "CMO_ARRAY_OF_64BIT_MACHINE_DOUBLE"
	case CMO_128BIT_MACHINE_DOUBLE:
		return "CMO_128BIT_MACHINE_DOUBLE"
	case CMO_ARRAY_OF_128BIT_MACHINE_DOUBLE:
		return "CMO_ARRAY_OF_128BIT_MACHINE_DOUBLE"

	case CMO_BIGFLOAT:
		return "CMO_BIGFLOAT"
	case CMO_IEEE_DOUBLE_FLOAT:
		return "CMO_IEEE_DOUBLE_FLOAT"
	case CMO_BIGFLOAT32:
		return "CMO_BIGFLOAT32"

	case CMO_INDETERMINATE:
		return "CMO_INDETERMINATE"
	case CMO_TREE:
		return "CMO_TREE"
	case CMO_LAMBDA:
		return "CMO_LAMBDA"
	default:
		return "unknown"
	}
}

func (ox *OpenXM) recvCMO(ringdef *List) (interface{}, error) {
	// Remarks: OX_TAG is already received.
	const fname = "recvCMO"
	var tag int32
	err := ox.dataRead(&tag)
	if err != nil {
		ox.logger.Printf("%s(tag) failed: %s", fname, err.Error())
		return nil, err
	}
	ox.logger.Printf("<--  %s() tag=%d:%s", fname, tag, ox.cmoTagString(tag))

	switch tag {
	case CMO_ZERO:
		return int32(0), nil
	case CMO_NULL:
		return int32(0), nil
	case CMO_INT32:
		return ox.recvCMOInt32()
	case CMO_STRING:
		return ox.recvCMOString()
	case CMO_ZZ:
		return ox.recvCMOZZ()
	case CMO_LIST: // 17
		return ox.recvCMOList()
	case CMO_RECURSIVE_POLYNOMIAL: // 27
		return ox.recvCMORPoly()
	case CMO_POLYNOMIAL_IN_ONE_VARIABLE: // 33
		return ox.recvCMOPoly1Var(ringdef)
	case CMO_INDETERMINATE: // 60
		return ox.recvCMOIndeterminate()
	case CMO_DATUM:
		m, _ := ox.dataReadInt32()
		ss := fmt.Sprintf("datam m=%d:", m)
		for i := int32(0); i < m; i++ {
			xx, _ := ox.dataReadInt32()
			ss += fmt.Sprintf(" %08x", xx)
		}
		ox.logger.Printf("%s() %s", fname, ss)
	case CMO_ERROR2:
		v, err := ox.recvCMO(nil)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("CMO_ERROR2 `%v`", v)

	}

	panic(fmt.Sprintf("unsupported: cmo=%d", tag))
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
	// ox.logger.Printf("PopCMO() start\n")
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
	tag, _, err := ox.PopOXTag()
	// ox.logger.Printf("PopCMO() receive: tag=%d, serial=%d\n", tag, serial)
	if tag != OX_DATA {
		return "", fmt.Errorf("PopCMO() unexpected tag=%d", tag)
	}
	v, err := ox.recvCMO(nil)
	if err != nil {
		ox.logger.Printf("PopCMO(get) failed: %s", err.Error())
		return "", err
	}
	return v, nil
}

func (ox *OpenXM) ExecString(val string) error {
	const fname = "ExecString"
	err := ox.PushOxCMO(val)
	if err != nil {
		ox.logger.Printf("%s() push-fname() failed", fname)
		return err
	}
	err = ox.PushOXCommand(SM_executeStringByLocalParser)
	if err != nil {
		ox.logger.Printf("%s(%d) push OxCommand() failed", fname, SM_executeStringByLocalParser)
		return err
	}
	err = ox.dw.Flush()
	return err
}

func (ox *OpenXM) ExecFunction(funcname string, argv ...interface{}) error {
	const fname = "ExecFunction"
	for i := len(argv) - 1; i >= 0; i-- {
		ox.logger.Printf("%s(arg-%d): %s() send %v", fname, i, funcname, argv[i])
		err := ox.PushOxCMO(argv[i])
		if err != nil {
			ox.logger.Printf("%s() push-arg(%d) %s() failed", fname, i, funcname)
			return err
		}
	}
	// ox.logger.Printf("%s(#arg-%d): send %s()", fname, len(argv), funcname)
	err := ox.PushOxCMO(int32(len(argv)))
	if err != nil {
		ox.logger.Printf("%s() push-arglen(%d) failed", fname, len(argv))
		return err
	}
	// ox.logger.Printf("%s(): send funcname %v", fname, funcname)
	err = ox.PushOxCMO(funcname)
	if err != nil {
		ox.logger.Printf("%s() push-fname() failed", fname)
		return err
	}
	// ox.logger.Printf("%s(): %s() send command %v", fname, funcname, SM_executeFunction)
	err = ox.PushOXCommand(SM_executeFunction)
	if err != nil {
		ox.logger.Printf("%s(%d) push OxCommand() failed", fname, SM_executeFunction)
		return err
	}
	err = ox.dw.Flush()
	// ox.logger.Printf("%s(): %s() finished", fname, funcname)
	return err
}

func (ox *OpenXM) PopString() (string, error) {
	// ox.logger.Printf("PopString() start\n")
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

	v, err := ox.recvCMO(nil)
	if err != nil {
		ox.logger.Printf("PopString(get) failed: %s", err.Error())
		return "", err
	}
	return v.(string), err
}
