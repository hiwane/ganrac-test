package ganrac

import (
	"testing"
)

func TestPolyRootBound(t *testing.T) {
	for _, s := range []struct {
		a     []int64
		bound NObj
	}{
		{[]int64{2, 1}, NewInt(2)},
		{[]int64{17, 0, 3}, NewRatInt64(238, 100)},
		{[]int64{39, -40, 0, 1}, NewInt(3)},
		{[]int64{-6, 25, -27, 4}, NewInt(3)},
		{[]int64{4, 12, 21, 20, 9, -3, -7, -3, 0, 1}, NewInt(2)}, // qebook 115
		{[]int64{3, 2, -1, -3, -2, 1}, NewInt(3)},                // qebook 125
		{[]int64{1, 3, 2}, one},                                  // CA p39
		{[]int64{-2, 1, -2, 1}, two},                             // CA p182
	} {
		a := NewPolyInts(0, s.a...)
		b, err := a.RootBound()
		if err != nil {
			t.Errorf("?? %s", err.Error())
			continue
		}

		if b.Cmp(s.bound) < 0 {
			t.Errorf("invalid bound expect=%v, actual=%v", s.bound, b)
			continue
		}

		c := a.rootBoundBinInt()
		if c.Cmp(b) < 0 {
			t.Errorf("invalid binint expect=%v, actual=%v -> %v", s.bound, b, c)
			continue
		}
	}
}
