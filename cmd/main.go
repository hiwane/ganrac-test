package main

import (
	"bufio"
	"fmt"
	"github.com/hiwane/ganrac"
	"io"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	fmt.Println("GANRAC. see help();")

	ganrac.InitVarList([]string{
		"x", "y", "z", "w", "a", "b", "c", "e", "f", "g", "h",
	})

	for {
		if _, err := os.Stdout.WriteString("> "); err != nil {
			fmt.Fprintf(os.Stderr, "WriteString: %s", err)
			break
		}
		line, err := in.ReadBytes(';')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReadBytes: %s", err)
			continue
		}

		p, err := ganrac.Eval(strings.NewReader(string(line)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			continue
		}
		if p != nil {
			fmt.Println(p)
		}
	}
}
