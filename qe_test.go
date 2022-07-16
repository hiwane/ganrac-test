package ganrac

import (
	"fmt"
	"testing"
)

func benchmarkQE(b *testing.B, name string) {
	input := GetExampleFof(name).Input
	g := NewGANRAC()
	connc, connd := testConnectOx(g)
	if g.ox == nil {
		fmt.Printf("skip TestNeqQE... (no ox)\n")
		return
	}
	defer connc.Close()
	defer connd.Close()
	for i := 0; i < b.N; i++ {
		opt := NewQEopt()
		g.QE(input, opt)
	}
}

func BenchmarkAdam1(b *testing.B) {
	benchmarkQE(b, "adam1")
}

func TestBench(t *testing.T) {
}
