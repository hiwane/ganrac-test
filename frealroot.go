package ganrac

// Krawczyk method.
// real root isolation of a univariate polynomial with big.Float coeffient
import (
	"math/big"
	"sort"
)

type iKraw struct {
	f, g, m, y, k  *Interval
	d              []*big.Float // 入力の多項式
	prec           uint
	one, two, mone *big.Float
	c              *big.Float

	stack []*Interval
}

func newIKraw(prec uint, d []*big.Float) *iKraw {
	kraw := new(iKraw)
	kraw.prec = prec
	kraw.one = big.NewFloat(1)
	kraw.mone = big.NewFloat(-1)
	kraw.two = big.NewFloat(2)
	kraw.d = d
	kraw.stack = make([]*Interval, 0)
	kraw.c = kraw.newFloat()
	return kraw
}

func (kraw *iKraw) newFloat() *big.Float {
	f := new(big.Float)
	f.SetPrec(kraw.prec)
	return f
}

func (kraw *iKraw) pop() *Interval {
	m := len(kraw.stack) - 1
	x := kraw.stack[m]
	kraw.stack = kraw.stack[:m]
	return x
}

func (kraw *iKraw) push(x *Interval) {
	kraw.stack = append(kraw.stack, x)
}

func (kraw *iKraw) empty() bool {
	return len(kraw.stack) == 0
}

func (kraw *iKraw) mid(x *Interval, mid *big.Float) *big.Float {
	mid.Add(x.inf, x.sup)
	mid.Quo(mid, kraw.two)
	return mid
}

func (kraw *iKraw) calG(x *Interval) {
	// f(X): 区間 X での値域の範囲を計算し, kraw->G に設定する.
	deg := len(kraw.d) - 1
	kraw.g = NewIntervalFloat(kraw.d[deg], kraw.prec)

	for i := deg - 1; i >= 0; i-- {
		kraw.g = kraw.g.Mul(x).(*Interval)
		kraw.g = kraw.g.AddFloat(kraw.d[i])
	}
}

func (kraw *iKraw) calFd(x *Interval) {
	// f'(X): 区間 X での微分値の範囲を計算し, kraw->F に設定する.
	deg := len(kraw.d) - 1

	n := kraw.newFloat()
	n.SetInt64(int64(deg))
	kraw.f = NewIntervalInt64(int64(deg), kraw.prec)
	kraw.f = kraw.f.MulFloat(kraw.d[deg])

	wk := newInterval(kraw.prec)
	for i := deg - 1; i >= 1; i-- {
		kraw.f = kraw.f.Mul(x).(*Interval)
		// fmt.Printf("  f=%v\n", kraw.f)

		n.SetInt64(int64(i))
		wk.SetFloat(kraw.d[i])
		wk.sup.Mul(wk.sup, n)
		wk.inf.Mul(wk.inf, n)
		kraw.f = kraw.f.Add(wk).(*Interval)
	}
}

func (kraw *iKraw) calK(x *Interval, l *big.Float) {
	l.Quo(kraw.two, l)
	kraw.m = kraw.f.MulFloat(l)
	kraw.m = kraw.m.AddFloat(kraw.mone)

	kraw.g = kraw.g.MulFloat(l)
	kraw.g = kraw.g.SubFloat(kraw.c)

	kraw.y = x.SubFloat(kraw.c)
	kraw.y = kraw.y.Mul(kraw.m).Add(kraw.g).(*Interval)

	kraw.k = kraw.y.Neg().(*Interval)
}

func (kraw *iKraw) calS() *big.Float {
	var sum *big.Float
	if kraw.m.inf.Sign() >= 0 {
		sum = kraw.m.sup
	} else {
		sum = kraw.newFloat()
		kraw.m.inf.Neg(sum)
		if kraw.m.sup.Sign() >= 0 {
			if kraw.m.sup.Cmp(sum) > 0 {
				sum = kraw.m.sup
			}
		}
	}
	return sum
}

func (kraw *iKraw) check(x *Interval) int {
	// 区間 X での根の存在性チェック.
	// kraw.g, kraw.f の情報を参照.
	// return 0 if no root in x
	// return 1 otherwise

	l := kraw.newFloat()
	l.Add(kraw.f.sup, kraw.f.inf)
	if l.Sign() == 0 { // @TODO 最初の区間設定で排除できる.
		return 4 // 分割せよ
	}

	kraw.calK(x, l)

	if x.sup.Cmp(kraw.k.inf) < 0 || kraw.k.sup.Cmp(x.inf) < 0 {
		return 0 // 根がない
	}

	if kraw.k.inf.Cmp(x.inf) <= 0 || x.sup.Cmp(kraw.k.sup) <= 0 {
		return 3 // 根が k で改善できるかも
	}

	sum := kraw.calS()

	if sum.Cmp(kraw.one) >= 0 {
		return 2 // 根が k にあるかも
	}
	return 1 // 根は x にひとつ
}

func (kraw *iKraw) improve(x *Interval) *Interval {
	l := kraw.newFloat()
	for i := 0; i < 15; i++ {
		// fmt.Printf("  improve(%d) %v\n", i, x)

		kraw.calFd(x)
		l.Add(kraw.f.inf, kraw.f.sup)
		if l.Sign() == 0 {
			return x
		}

		kraw.mid(x, kraw.c)

		cc := newInterval(kraw.prec)
		cc.sup.Set(kraw.c)
		cc.inf.Set(kraw.c)

		kraw.calG(cc)
		kraw.calK(x, l)

		if x.sup.Cmp(kraw.k.inf) < 0 || kraw.k.sup.Cmp(x.inf) < 0 {
			return x
		}

		sum := kraw.calS()
		if sum.Cmp(kraw.one) >= 0 {
			return x
		}

		cnt := 0
		if kraw.k.inf.Cmp(x.inf) <= 0 {
			kraw.k.inf = x.inf
			cnt++
		}
		if kraw.k.sup.Cmp(x.sup) >= 0 {
			kraw.k.sup = x.sup
			cnt++
		}
		if cnt == 2 {
			return x
		}

		x = kraw.k
	}
	return x
}

func (kraw *iKraw) rootbound() *big.Float {
	var bound, c *big.Float

	t := kraw.newFloat()
	for i := len(kraw.d) - 2; i >= 0; i-- {
		if kraw.d[i].Sign() > 0 {
			c = kraw.d[i]
		} else {
			c = t
			t.Neg(kraw.d[i])
		}
		if bound == nil || bound.Cmp(c) < 0 {
			bound = c
		}
	}

	i := len(kraw.d) - 1
	if kraw.d[i].Sign() > 0 {
		c = kraw.d[i]
	} else {
		c = t
		t.Neg(kraw.d[i])
	}

	t.Quo(bound, c)
	t.Add(t, kraw.one)

	return t
}

func (kraw *iKraw) fRealRoot(x *Interval, lmax int) []*Interval {

	kraw.push(x)

	ans := make([]*Interval, 0, len(kraw.d)-1)

	for i := 0; !kraw.empty() && i < lmax; i++ {
		x = kraw.pop()
		// fmt.Printf("@ kraw.pop() x=%v\n", x)

		kraw.calG(x)
		if kraw.g.sup.Sign() < 0 || kraw.g.inf.Sign() > 0 {
			continue // 値域が 0 を含まない
		}
		kraw.mid(x, kraw.c)
		cc := newInterval(kraw.prec)
		cc.sup.Set(kraw.c)
		cc.inf.Set(kraw.c)

		kraw.calG(cc)
		kraw.calFd(x)

		chk := kraw.check(x)
		// fmt.Printf("# kraw() @@ [%2d,%d,%d] x=%v\n", len(kraw.stack), chk, len(ans), x)
		switch chk {
		case 1: // 根が一つ見つかった.
			kraw.k = kraw.improve(kraw.k)
			// fmt.Printf("! kraw() find-improved:: %v\n", kraw.k)
			ans = append(ans, kraw.k)
			continue
		case 2: // 根があるかも.
			x = kraw.k
		case 3: // 根があるかも.
			if kraw.k.inf.Cmp(x.inf) > 0 {
				x.inf.Set(kraw.k.inf)
			}
			if kraw.k.sup.Cmp(x.sup) < 0 {
				x.sup.Set(kraw.k.sup)
			}
		case 4:
			break
		case 0: // 根は存在しない
			continue
		default:
			panic("stop")
		}

		t := kraw.newFloat()
		kraw.mid(x, t)
		if x.sup.Cmp(x.inf) <= 0 || x.sup.Cmp(t) == 0 || x.inf.Cmp(t) == 0 {
			// prec が小さいため根を表現できない
			// fmt.Printf("  x=%v, t=%v\n", x, t)
			return nil
		}

		x2 := newInterval(kraw.prec)
		x2.inf.Set(t)
		x2.sup.Set(x.sup)
		kraw.push(x2)
		x2 = newInterval(kraw.prec)
		x2.sup.Set(t)
		x2.inf.Set(x.inf)
		kraw.push(x2)
	}

	if !kraw.empty() {
		return nil
	}

	sort.Slice(ans, func(i, j int) bool {
		return ans[i].inf.Cmp(ans[j].inf) <= 0
	})

	return ans
}

func FRealRoot(prec uint, lmax int, poly []*big.Float, rand float64) []*Interval {
	// poly: univariate polynomial
	// return: real root isolation
	kraw := newIKraw(prec, poly)
	bound := kraw.rootbound()
	bound.Add(bound, big.NewFloat(rand))

	x := newInterval(prec)
	x.inf.Neg(bound)
	x.sup.Set(bound)

	return kraw.fRealRoot(x, lmax)
}
