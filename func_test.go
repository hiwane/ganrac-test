package ganrac

import (
	"bufio"
	"strings"
	"testing"
)

func TestBuiltinFuncTable(t *testing.T) {
	g := NewGANRAC()
	name := "a"
	for i, bf := range g.builtin_func_table {
		if name > bf.name {
			t.Errorf("unsorted. i=%d,prev=%s,next=%s", i, name, bf.name)
			break
		}
		name = bf.name
	}
}

func TestHelpExamples(t *testing.T) {
	g := NewGANRAC()
	for _, bf := range g.builtin_func_table {
		if bf.ox && g.ox == nil {
			continue
		}

		scan := bufio.NewScanner(strings.NewReader(bf.help))

		pos := 0
		exp := ""
		for scan.Scan() {
			line := scan.Text()
			if line == "Examples" {
				pos++
			} else if strings.HasPrefix(line, "Examples") {
				t.Errorf("Examples...?")
				break
			} else if strings.HasPrefix(line, "========") && pos == 1 {
				pos++
			} else if strings.HasPrefix(line, "========") && pos == 2 {
				break
			} else if pos == 2 && strings.HasPrefix(line, "  > ") {
				exp = line[4:]
			} else if pos == 2 {
				example, err := g.Eval(strings.NewReader(exp))
				if err != nil {
					t.Errorf("invalid 1...`%s` %s", exp, err.Error())
					break
				}
				answer, err := g.Eval(strings.NewReader(line + ";"))
				if err != nil {
					t.Errorf("invalid 2...`%s` %s", line, err.Error())
					break
				}
				switch ee := example.(type) {
				case RObj:
					aa, ok := answer.(RObj)
					if !ok || !ee.Equals(aa) {
						t.Errorf("invalid r... `%s` `%s`", exp, line)
					}
				case Fof:
					aa, ok := answer.(Fof)
					if !ok || !ee.Equals(aa) {
						t.Errorf("invalid f... `%s` `%s`", exp, line)
					}
				case List:
					aa, ok := answer.(List)
					if !ok || !ee.Equals(aa) {
						t.Errorf("invalid f... `%s` `%s`", exp, line)
					}
				default:
					t.Errorf("invalid f... `%s` `%s`", exp, line)
				}
			}
		}
	}
}
