package ganrac

import (
	"fmt"
)

type QeExample struct {
	Input  Fof
	Output Fof
	Ref    string
	DOI    string
}

type qeExTable struct {
	name string
	f    func() *QeExample
}

var qeExampleTable []qeExTable = []qeExTable{
	{"adam1", exAdam1},
	{"adam2-1", exAdam2_1},
	{"adam2-2", exAdam2_2},
	{"adam3", exAdam3},
	{"candj", exCandJ},
	{"easy7", exEasy7},
	{"makepdf", exMakePdf},
	{"makepd2", exMakePdf2},
	{"pl01", exPL01},
	{"quad", exQuad},
	{"wo1", exWO1},
	{"wo2", exWO2},
	{"wo3", exWO3},
	{"wo4", exWO4},
	{"xaxis", exXAxisEllipse},
}

func GetExampleFof(name string) *QeExample {
	if name == "" {
		fmt.Printf("label\t# free\t# q \tdeg(f)\tdeg(q)\tatom\n")
		fmt.Printf("=====\t======\t====\t======\t======\t====\n")
		for _, t := range qeExampleTable {
			q := t.f()
			v := q.Input.maxVar()
			fdeg := 0
			qdeg := 0
			fnum := 0
			qnum := 0
			for i := Level(0); i <= Level(v); i++ {
				d := q.Input.Deg(i)
				if q.Input.hasFreeVar(i) {
					fnum++
					if d > fdeg {
						fdeg = d
					}
				} else if q.Input.hasVar(i) {
					qnum++
					if d > qdeg {
						qdeg = d
					}
				}
			}

			fmt.Printf("%s\t%4d\t%4d\t%4d\t%4d\t%4d\n", t.name, fnum, qnum, fdeg, qdeg, q.Input.numAtom())
		}
		return nil
	}

	for _, t := range qeExampleTable {
		if t.name == name {
			// fmt.Printf("%S\n", t.f().Input)
			return t.f()
		}
	}
	return nil
}

func exAdam1() *QeExample {
	q := new(QeExample)
	q.Output = trueObj
	q.Input = NewQuantifier(true, []Level{0, 1}, newFmlOrs(
		NewAtom(NewPolyCoef(0, 0, 1), GE),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -49719, 0, 50000), 0, 50000), GE),
		NewAtom(NewPolyCoef(1,
			NewPolyCoef(0, 0, 720000, 720000, 480000, 240000, 96000, 32200, 9200, 2225, 450, 75, 10, 1),
			0,
			NewPolyCoef(0, 0, 0, 0, 0, -3000, 1200, 2100, 1000, 275, 50, 6),
			0,
			NewPolyCoef(0, 0, 0, 3000, -6000, -2250, 300, 350, 100, 15),
			0,
			NewPolyCoef(0, -200, 2000, -1900, -600, 150, 100, 20),
			0,
			NewPolyCoef(0, 225, -350, -25, 50, 15),
			0,
			NewPolyCoef(0, -25, 10, 6), 0, 1), LT)))
	q.Ref = "Adam W. Strzebonski. Cylindrical Algebraic Decomposition using validated numerics. 2006"
	q.DOI = "10.1016/j.jsc.2006.06.004"
	return q
}

func exAdam2_1() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{0, 1, 2}, newFmlOrs(
		NewAtom(NewPolyCoef(0, 0, 1), LT),
		NewAtom(NewPolyCoef(1, 0, 1), LT),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -1, 0, 4), 0, 4), GE),
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 4, -3, -4, 4), NewPolyCoef(0, -4, 2, 8, -8), NewPolyCoef(0, 5, -12, 8)), NewPolyCoef(1, NewPolyCoef(0, 0, 0, 2, -4), NewPolyCoef(0, 0, 4, -4, 8), NewPolyCoef(0, 2, -4, -8), NewPolyCoef(0, -4, 8)), NewPolyCoef(1, NewPolyCoef(0, 0, -4, 5), NewPolyCoef(0, 4, 2, -12), NewPolyCoef(0, -3, 8, 8), NewPolyCoef(0, -4, -8), 4)), LE),
		NewAtom(NewPolyCoef(2, 0, 0, NewPolyCoef(1, NewPolyCoef(0, 0, 0, -2, 0, 2), NewPolyCoef(0, 0, 2, 0, -4), NewPolyCoef(0, 0, -4, 6)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, -4, 4), NewPolyCoef(0, 0, -4), NewPolyCoef(0, 0, 4)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 2, -4), NewPolyCoef(0, -2, 0, 6), NewPolyCoef(0, 0, -4), 2)), LE)))

	q.Output = trueObj
	q.Ref = "Adam W. Strzebonski. Cylindrical Algebraic Decomposition using validated numerics. 2006"
	q.DOI = "10.1016/j.jsc.2006.06.004"
	return q
}
func exAdam2_2() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{0, 1, 2}, newFmlOrs(
		NewAtom(NewPolyCoef(0, 0, 1), LT),
		NewAtom(NewPolyCoef(0, -1, 1), GT),
		NewAtom(NewPolyCoef(1, 0, 1), LT),
		NewAtom(NewPolyCoef(1, -1, 1), GT), newFmlAnds(
			NewAtom(NewPolyCoef(2, 0, 0, 0, 0, NewPolyCoef(1, NewPolyCoef(0, 0, 0, -1, 0, 1), NewPolyCoef(0, 0, 2, 0, -4), NewPolyCoef(0, -1, 0, 6), NewPolyCoef(0, 0, -4), 1)), LE),
			NewAtom(NewPolyCoef(2, NewPolyCoef(0, 0, 0, -1, 0, 1), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, -4, 4)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 2, -4), NewPolyCoef(0, 0, -4, 6)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, -4), NewPolyCoef(0, 0, 4)), NewPolyCoef(1, 0, 0, -1, 0, 1)), LE), newFmlOrs(
				NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 4, -3, -4, 4), NewPolyCoef(0, -4, 2, 8, -8), NewPolyCoef(0, 5, -12, 8)), NewPolyCoef(1, NewPolyCoef(0, 0, 0, 2, -4), NewPolyCoef(0, 0, 4, -4, 8), NewPolyCoef(0, 2, -4, -8), NewPolyCoef(0, -4, 8)), NewPolyCoef(1, NewPolyCoef(0, 0, -4, 5), NewPolyCoef(0, 4, 2, -12), NewPolyCoef(0, -3, 8, 8), NewPolyCoef(0, -4, -8), 4)), LE),
				NewAtom(NewPolyCoef(2, 0, 0, NewPolyCoef(1, NewPolyCoef(0, 0, 0, -2, 0, 2), NewPolyCoef(0, 0, 2, 0, -4), NewPolyCoef(0, 0, -4, 6)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 0, -4, 4), NewPolyCoef(0, 0, -4), NewPolyCoef(0, 0, 4)), NewPolyCoef(1, 0, NewPolyCoef(0, 0, 2, -4), NewPolyCoef(0, -2, 0, 6), NewPolyCoef(0, 0, -4), 2)), LE)))))

	q.Output = trueObj
	q.Ref = "Adam W. Strzebonski. Cylindrical Algebraic Decomposition using validated numerics. 2006"
	q.DOI = "10.1016/j.jsc.2006.06.004"
	return q
}
func exAdam3() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{2, 3}, NewQuantifier(true, []Level{1}, newFmlAnds(
		NewAtom(NewPolyCoef(2, -1, 1), GT),
		NewAtom(NewPolyCoef(3, 0, 1), GT),
		NewAtom(NewPolyCoef(0, 0, 1), GT),
		NewAtom(NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, -1), 0, NewPolyCoef(0, 0, -1)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, -2)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 1, -1), 0, 1)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, 2)), NewPolyCoef(1, NewPolyCoef(0, 1, -1), 0, 1)), LE),
		NewAtom(NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, -1), 0, NewPolyCoef(0, 0, -1)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, 2)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 1, -1), 0, 1)), NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, 2)), NewPolyCoef(1, NewPolyCoef(0, 1, -1), 0, 1)), LE))))
	q.Input = NewQuantifier(false, []Level{1, 2}, NewQuantifier(true, []Level{3}, newFmlAnds(
		NewAtom(NewPolyCoef(1, -1, 1), GT),
		NewAtom(NewPolyCoef(2, 0, 1), GT),
		NewAtom(NewPolyCoef(0, 0, 1), GT),
		NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, 0, NewPolyCoef(0, 1, -1)), 0, NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, -1), NewPolyCoef(0, 0, -2), NewPolyCoef(0, 1, -1)), NewPolyCoef(0, 0, 2), 1), 0, NewPolyCoef(1, NewPolyCoef(0, 0, -1), 0, 1)), LE),
		NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, 0, NewPolyCoef(0, 1, -1)), 0, NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, -1), NewPolyCoef(0, 0, 2), NewPolyCoef(0, 1, -1)), NewPolyCoef(0, 0, 2), 1), 0, NewPolyCoef(1, NewPolyCoef(0, 0, -1), 0, 1)), LE))))

	q.Output = NewAtom(NewPolyCoef(0, -4, 1), GT)
	q.Ref = "Adam W. Strzebonski. Cylindrical Algebraic Decomposition using validated numerics. 2006"
	q.DOI = "10.1016/j.jsc.2006.06.004"
	return q
}

func exCandJ() *QeExample {
	// ex([z], z>0 && z-1<0 && y>0 && 2*x >= 1 && (3*y^2+3*x^2-2*x)*z+-y^2-x^2<0 && (3*y^2+3*x^2-4*x+1)*z+-2*y^2+-2*x^2+2*x>0)

	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{2}, newFmlAnds(
		NewAtom(NewPolyCoef(2, 0, 1), GT),
		NewAtom(NewPolyCoef(2, -1, 1), LT),
		NewAtom(NewPolyCoef(1, 0, 1), GT),
		NewAtom(NewPolyCoef(0, -1, 2), GE),
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 0, -1), 0, -1), NewPolyCoef(1, NewPolyCoef(0, 0, -2, 3), 0, 3)), LT),
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 2, -2), 0, -2), NewPolyCoef(1, NewPolyCoef(0, 1, -4, 3), 0, 3)), GT)))

	q.Output = q.Input

	return q
}

func exEasy7() *QeExample {
	// ex([x], a*x^2+b*x+c == 0)
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{2}, newFmlAnds(
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, -1), -2), 1), EQ),
		NewAtom(NewPolyCoef(2, -125, 0, 1), EQ),
		NewAtom(NewPolyCoef(2, 0, 1), GT),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -25, 0, 1), 0, 1), LE),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -50, -20, -2), NewPolyCoef(0, -5, -1), 1), LE)))
	q.Output = newFmlAnds(
		NewAtom(NewPolyCoef(0, 0, 1), GT),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -25, 0, 1), 0, 1), EQ),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -125, 0, 1),
			NewPolyCoef(0, 0, 4), 4), EQ))
	q.Ref = "syn_pdq error"

	return q
}
func exMakePdf() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{1}, newFmlAnds(
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, -1, 0, 1), 0, 1), EQ),
		NewAtom(NewPolyCoef(1, NewPolyCoef(0, 0, 1), one), LT)))
	q.Output = q.Input
	q.Ref = "Christopher W. Brown. Solution formula construction for truth invariant CAD's. Thesis p65 1999"

	return q
}

func exMakePdf2() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{2}, newFmlAnds(
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, -1, 0, 1), 0, 1), 0, 1), EQ),
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, -1, 1), 2), 1), LT)))
	q.Output = q.Input
	q.Ref = "Christopher W. Brown. Solution formula construction for truth invariant CAD's. Thesis p65 1999"

	return q
}

func exPL01() *QeExample {
	q := new(QeExample)
	q.Output = NewAtom(NewPolyCoef(0, 0, 1), LE)
	q.Input = NewQuantifier(true, []Level{1, 2}, newFmlOrs(
		NewAtom(NewPolyCoef(1, 1, 1), LT),
		NewAtom(NewPolyCoef(1, -1, 1), GT),
		NewAtom(NewPolyCoef(2, 1, 1), LT),
		NewAtom(NewPolyCoef(2, -1, 1), GT),
		NewAtom(NewPolyCoef(2, NewPolyCoef(0, 1, -1), 0, NewPolyCoef(1, 0, 0, -3, 0, 1), 0, NewPolyCoef(1, 0, 0, 1)), GE)))
	q.Ref = "P. Parrilo and S. Lall. Semidefinite Programming Relaxation and Algebraic Optimization in Control."
	q.DOI = ""
	return q
}

func exQuad() *QeExample {
	// ex([x], a*x^2+b*x+c == 0)
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{3}, NewAtom(NewPolyCoef(3, NewPolyCoef(2, 0, 1), NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 1)), EQ))
	q.Output = newFmlOrs(newFmlAnds(
		NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, 0, -1), NewPolyCoef(0, 0, 4)), LE),
		NewAtom(NewPolyCoef(0, 0, 1), NE)), newFmlAnds(
		NewAtom(NewPolyCoef(0, 0, 1), EQ),
		NewAtom(NewPolyCoef(1, 0, 1), NE)),
		NewAtom(NewPolyCoef(2, 0, 1), EQ))

	return q
}

func exWO1() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{3}, NewAtom(NewPolyCoef(3, 1, NewPolyCoef(2, 0, 1), NewPolyCoef(1, 0, 1), 0, 0, 0, NewPolyCoef(0, 0, 1)), GT))
	q.Output = q.Input
	q.Ref = "original: NOT well-oriented"
	return q
}

func exWO2() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{3}, NewAtom(
		NewPolyCoef(3,
			NewPolyCoef(1, 0, NewPolyCoef(0, 0, -1)),
			NewPolyCoef(2, 0, NewPolyCoef(1, 0, 1)),
			1), GE))
	q.Output = newFmlOrs(
		newFmlAnds(NewAtom(NewPolyCoef(1, 0, 1), GE),
			NewAtom(NewPolyCoef(2, NewPolyCoef(0, 0, 4), 0, NewPolyCoef(1, 0, 1)), LE)),
		newFmlAnds(NewAtom(NewPolyCoef(2, NewPolyCoef(0, 0, 4), 0, NewPolyCoef(1, 0, 1)), GE),
			NewAtom(NewPolyCoef(1, 0, 1), LE)))
	q.Ref = "original: NOT well-oriented, but no delineating polynomial is needed!"
	return q
}

func exWO3() *QeExample {
	// (c,d,b,x)
	// 3
	// (E x) [ x >= 0 /\ x^3 + b x^2 + c x + d < 0 ].
	// Error! Delineating polynomial should be added over cell(2,2)!
	// d-cell (2,2) -> (x=0, y=0)
	// Degrees after substitution  : (-1)
	// x=y=0
	// P_3,1  = fac(J_3,1) = fac(dis(A_4,1))
	//        = 4 c^3 - b^2 c^2 - 18 b d c + 27 d^2 + 4 b^3 d
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{3}, newFmlAnds(
		NewAtom(NewPolyCoef(3, 0, 1), GE),
		NewAtom(NewPolyCoef(3, NewPolyCoef(1, 0, 1), NewPolyCoef(0, 0, 1), NewPolyCoef(2, 0, 1), 1), LT)))
	q.Output = newFmlOrs(
		NewAtom(NewPolyCoef(1, 0, 1), LT),
		newFmlAnds(
			NewAtom(NewPolyCoef(2, 0, 1), LT),
			NewAtom(NewPolyCoef(2, NewPolyCoef(0, 0, -4), 0, 1), GT),
			NewAtom(NewPolyCoef(1, 0, 1), EQ)),
		newFmlAnds(
			NewAtom(NewPolyCoef(2, 0, 1), LT),
			NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 0, 0, 4), 0, 27), NewPolyCoef(1, 0, NewPolyCoef(0, 0, -18)), NewPolyCoef(0, 0, 0, -1), NewPolyCoef(1, 0, 4)), LT)),
		newFmlAnds(
			NewAtom(NewPolyCoef(0, 0, 1), LT),
			NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyCoef(0, 0, 0, 0, 4), 0, 27), NewPolyCoef(1, 0, NewPolyCoef(0, 0, -18)), NewPolyCoef(0, 0, -1), NewPolyCoef(1, 0, 4)), LT)),
	)
	q.Ref = "original(SDC): NOT well-oriented"

	return q
}

func exWO4() *QeExample {
	// (x,y,z,w)
	// 3
	// (E w) [ w^2 < x /\ z w + y <= 0 ].
	q := new(QeExample)
	q.Ref = "original: NOT well-oriented, but no delineating polynomial is needed!"
	q.Input = NewQuantifier(false, []Level{3}, newFmlAnds(
		NewAtom(NewPolyCoef(3, NewPolyCoef(0, 0, -1), 0, 1), LT),
		NewAtom(NewPolyCoef(3, NewPolyCoef(1, 0, 1), NewPolyCoef(2, 0, 1)), LE)))
	q.Output = newFmlAnds(
		NewAtom(NewPolyCoef(0, 0, 1), GT),
		newFmlOrs(
			NewAtom(NewPolyCoef(1, 0, 1), LE),
			NewAtom(NewPolyCoef(2, NewPolyCoef(1, 0, 0, -1), 0, NewPolyCoef(0, 0, 1)), GT)))
	return q
}

func exXAxisEllipse() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{3, 4}, newFmlAnds(
		NewAtom(NewPolyCoef(0, 0, 1), GT),
		NewAtom(NewPolyCoef(1, 0, 1), GT), newFmlOrs(
			NewAtom(NewPolyCoef(4, NewPolyCoef(3, NewPolyCoef(2, NewPolyCoef(1, 0, 0, NewPolyCoef(0, 0, 0, -1)), 0, NewPolyCoef(1, 0, 0, 1)), NewPolyCoef(2, 0, NewPolyCoef(1, 0, 0, -2)), NewPolyCoef(1, 0, 0, 1)), 0, NewPolyCoef(0, 0, 0, 1)), NE),
			NewAtom(NewPolyCoef(4, NewPolyCoef(3, -1, 0, 1), 0, 1), LE))))
	q.Output = q.Input
	q.Ref = "The x-axix ellipse problem: W. Kahan. Problem no. 9: An ellipse problem."
	return q
}
