package ganrac

import (
	"bufio"
	"net"
	"time"
)

func testConnectOx(g *Ganrac) (net.Conn, net.Conn) {
	cport := "localhost:1234"
	dport := "localhost:4321"
	connc, err := net.Dial("tcp", cport)
	if err != nil {
		return nil, nil
	}

	time.Sleep(time.Second / 20)

	connd, err := net.Dial("tcp", dport)
	if err != nil {
		connc.Close()
		return nil, nil
	}

	dw := bufio.NewWriter(connd)
	dr := bufio.NewReader(connd)
	cw := bufio.NewWriter(connc)
	cr := bufio.NewReader(connc)

	g.ConnectOX(cw, dw, cr, dr)
	return connc, connd
}
