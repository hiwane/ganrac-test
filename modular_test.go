package ganrac

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func TestModularInvUint(t *testing.T) {
	for _, s := range []struct {
		a Uint
		p Uint
	}{
		{1, 5},
		{2, 5},
		{3, 5},
		{4, 5},
		{1, 11},
		{2, 11},
		{3, 11},
		{5, 11},
		{6, 11},
		{7, 11},
	} {
		b := s.a.inv_mod(nil, s.p).(Uint)
		if b == 0 || b >= s.p {
			t.Errorf("invalid input=%v, p=%v, inv=%v", s.a, s.p, b)
			continue
		}
		u := (uint64(b) * uint64(s.a)) % uint64(s.p)
		if u != 1 {
			t.Errorf("invalid input=%v, p=%v, inv=%v, mul=%v", s.a, s.p, b, u)
			continue
		}
	}
}

func TestModularInvPoly(t *testing.T) {
	cad := new(CAD)
	cad.root = NewCell(cad, nil, 0)
	cell0 := NewCell(cad, nil, 1)
	cell0.lv = 0
	cell0.parent = cad.root
	cell0.defpoly = NewPolyCoef(0, -1, 0, 151)
	cell1 := NewCell(cad, nil, 1)
	cell1.lv = 1
	cell1.parent = cell0
	cell1.defpoly = NewPolyCoef(1, -1, 0, NewPolyCoef(0, 0, 0, 1))

	cell := cell1

	for _, s := range []struct {
		a *Poly
		p []Uint
	}{
		{
			NewPolyCoef(0, 3, 5),
			[]Uint{2, 3, 151, 17, lprime_table[10]},
		},
	} {
		for _, p := range s.p {
			a := s.a.mod(p)
			cellp, ok := cell.mod(cad, p)
			if !ok {
				if p == 151 {
					continue
				}
				for c := cellp; c != nil; c = c.parent {
					fmt.Printf("[%d] `%v` `%v` `%v`\n", c.lv, c.defpoly, c.factor1, c.factor2)
				}
				panic("!")
			}

			fmt.Printf("go invmod: a=%v mod %v\n", a, p)
			b := a.inv_mod(cellp, p)
			if b == nil {
				// 共通根をもった... 定義多項式が分解された
				continue
			}
			if err := b.valid_mod(cellp, p); err != nil {
				t.Errorf("a=%v, p=%d, b=%v, err=%w", a, p, b, err)
				return
			}
			if b == Uint(0) {
				t.Errorf("a=%v, p=%d, b=%v", a, p, b)
				return
			}

			fmt.Printf("go invmod: b=%v, a=%v\n", b, a)
			ab := a.mul_mod(b, p).simpl_mod(cellp, p)
			if !ab.IsOne() {
				t.Errorf("a=%v, p=%d, b=%v, ab=%v", a, p, b, ab)
				return
			}
		}
	}
}

func TestModularInterpol(t *testing.T) {
	seed := time.Now().UnixNano()
	fmt.Printf("TestModularInterpol() seed=%v\n", seed)
	r := rand.NewSource(seed)

	for i := 0; i < 20; i++ {
		f := randPoly(r, 3, 5, 100, 20)

		var g *Poly
		var p *Int
		ps := make([]Uint, 0, 10)
		primes := []Uint{2, 3, 7, 23, 37, 71, 11, 13, 17}
		for len(ps) < 3 {
			idx := r.Int63() % int64(len(primes))
			q := primes[idx]
			primes = append(primes[:idx], primes[idx+1:]...)

			_fp := f.mod(q)
			fp, ok := _fp.(*Poly)
			if !ok {
				i--
				continue
			}
			ps = append(ps, q)
			if len(ps)+len(primes) != 9 {
				panic("why?")
			}

			if len(ps) <= 1 {
				g = fp
				p = NewInt(int64(q))
			} else {
				g, p, _ = g.crt_interpol(fp, p, q)
				if g == nil {
					t.Errorf("seed=%d, i=%d\nin =%v\nout=%v\nq =%v\nfp =%v\n", seed, i, f, g, q, fp)
					return
				}
				if err := g.valid(); err != nil {
					t.Errorf("invalid... i=%d, seed=%d, %v", i, seed, err)
					return
				}
			}
		}
		if err := g.valid(); err != nil {
			t.Errorf("invalid... i=%d, seed=%d, %v", i, seed, err)
			return
		}

		for _, q := range ps {
			gp := g.mod(q)
			fp := f.mod(q)
			if !fp.Equals(gp) {
				t.Errorf("i=%d, seed=%d\nin =%v\nout=%v\nq =%v\nfp =%v\ngp =%v\n", i, seed, f, g, q, fp, gp)
				break
			}
		}
	}
}

func TestModularInterpolZ(t *testing.T) {
	seed := time.Now().UnixNano()
	fmt.Printf("TestModularInterpolZ() seed=%v\n", seed)
	r := rand.New(rand.NewSource(seed))

	max := big.NewInt(1)
	max.Lsh(max, 150)

	primes := []Uint{5, 7, 3, lprime_table[0], lprime_table[10], 13, 17, 23}

	type boo struct {
		p   Uint
		px  *Int
		mui Uint
		mx  *Int
	}

	for i := 0; i < 10; i++ {
		m := new(big.Int)
		m.Rand(r, max)

		p := make([]*boo, 0, 3)
		for len(p) < 3 {
			q := primes[r.Int()%len(primes)]
			found := false
			for _, pp := range p {
				if pp.p == q {
					found = true
				}
			}
			if !found {
				b := new(boo)
				b.p = q
				b.px = NewInt(int64(q))
				b.mx = newInt()
				b.mx.n.Mod(m, b.px.n)
				b.mui = Uint(b.mx.Int64())
				p = append(p, b)
			}
		}

		xx := new(big.Int)
		pqinf := p[0].px._crt_init(p[1].p)

		pqpinv := p[0].px.Mul(NewInt(int64(pqinf.pinv))).(*Int)
		xx.Mod(pqpinv.n, p[1].px.n)
		if xx.Cmp(one.n) != 0 {
			t.Errorf("invalid p=%v q=%v pq=%v pinv=%v\n", p[0].p, p[1].p, pqinf.pq, pqinf.pinv)
			return
		}

		m2 := p[0].mx.interpol_ui(p[1].mui, pqinf)
		if m2.Sign() < 0 || m2.Cmp(pqinf.pq) >= 0 {
			t.Errorf("invalid m=%v, m2=%v, pq=%v\n", m, m2, pqinf.pq)
			return
		}

		pqinf2 := pqinf.pq._crt_init(p[2].p)
		m3 := m2.interpol_ui(p[2].mui, pqinf2)
		if m3.Sign() < 0 || m3.Cmp(pqinf2.pq) >= 0 {
			t.Errorf("invalid m=%v, m3=%v, pq=%v, pinv=%v\n", m, m3, pqinf2.pq, pqinf2.pinv)
			return
		}

		for _, b := range p {
			xx.Mod(m3.n, b.px.n)
			if xx.Cmp(b.mx.n) != 0 {
				t.Errorf("invalid b=%v, m3=%v, p=%v xx=%v\n", b, m3, b.p, xx)
				return
			}
		}
	}
}
