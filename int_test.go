package ganrac

import (
	"testing"
)

func TestInts(t *testing.T) {
	v := NewInt(0)
	if v.Sign() != 0 || v.String() != "0" || v.IsOne() || v.IsMinusOne() || !v.IsZero() {
		t.Errorf("invalid int v=%v, sign=%d, str=%s\n", v, v.Sign(), v.String())
	}

	v = NewInt(1)
	if v.Sign() <= 0 || v.String() != "1" || !v.IsOne() || v.IsMinusOne() || v.IsZero() {
		t.Errorf("invalid int v=%v, sign=%d, str=%s\n", v, v.Sign(), v.String())
	}

	v = NewInt(-1)
	if v.Sign() >= 0 || v.String() != "-1" || v.IsOne() || !v.IsMinusOne() || v.IsZero() {
		t.Errorf("invalid int v=%v, sign=%d, str=%s\n", v, v.Sign(), v.String())
	}
}
