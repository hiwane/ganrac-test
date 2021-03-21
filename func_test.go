package ganrac

import (
	"testing"
)

func TestBuiltinFuncTable(t *testing.T) {
	name := "a"
	for i, bf := range builtin_func_table {
		if name > bf.name {
			t.Errorf("unsorted. i=%d,prev=%s,next=%s", i, name, bf.name)
			break
		}
		name = bf.name
	}
}
