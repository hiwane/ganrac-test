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
	{"candj", exCandJ},
	{"easy7", exEasy7},
	{"makepdf", exMakePdf},
	{"makepd2", exMakePdf2},
	{"pl01", exPL01},
	{"quad", exQuad},
	{"wo1", exWO1},
	{"wo2", exWO2},
}

func GetExampleFof(name string) *QeExample {
	if name == "" {
		fmt.Printf("label\t# free\t# q\tdeg(f)\tdeg(q)\tatom\n")
		fmt.Printf("=====\t======\t===\t======\t======\t====\n")
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

			fmt.Printf("%s\t%2d\t%2d\t%2d\t%2d\t%2d\n", t.name, fnum, qnum, fdeg, qdeg, q.Input.numAtom())
		}
		return nil
	}

	for _, t := range qeExampleTable {
		if t.name == name {
			return t.f()
		}
	}
	return nil
}

func exAdam1() *QeExample {
	q := new(QeExample)
	q.Output = trueObj
	q.Input = NewQuantifier(true, []Level{0, 1}, newFmlOrs(
		NewAtom(NewPolyInts(0, 0, 1), GE),
		NewAtom(NewPolyCoef(1, NewPolyInts(0, -49719, 0, 50000), NewInt(0), NewInt(50000)), GE),
		NewAtom(NewPolyCoef(1,
			NewPolyInts(0, 0, 720000, 720000, 480000, 240000, 96000, 32200, 9200, 2225, 450, 75, 10, 1),
			NewInt(0),
			NewPolyInts(0, 0, 0, 0, 0, -3000, 1200, 2100, 1000, 275, 50, 6),
			NewInt(0),
			NewPolyInts(0, 0, 0, 3000, -6000, -2250, 300, 350, 100, 15),
			NewInt(0),
			NewPolyInts(0, -200, 2000, -1900, -600, 150, 100, 20),
			NewInt(0),
			NewPolyInts(0, 225, -350, -25, 50, 15),
			NewInt(0),
			NewPolyInts(0, -25, 10, 6),
			NewInt(0),
			NewInt(1)), LT)))
	q.Ref = "Adam W. Strzebonski. Cylindrical Algebraic Decomposition using validated numerics. 2006"
	q.DOI = "10.1016/j.jsc.2006.06.004"
	return q
}

func exCandJ() *QeExample {
	// ex([z], z>0 && z-1<0 && y>0 && 2*x >= 1 && (3*y^2+3*x^2-2*x)*z+-y^2-x^2<0 && (3*y^2+3*x^2-4*x+1)*z+-2*y^2+-2*x^2+2*x>0)

	q := new(QeExample)
	q.Output = NewAtom(NewPolyInts(0, 0, 1), LE)
	q.Input = NewQuantifier(false, []Level{2}, newFmlAnds(NewAtom(NewPolyInts(2, 0, 1), GT), NewAtom(NewPolyInts(2, -1, 1), LT), NewAtom(NewPolyInts(1, 0, 1), GT), NewAtom(NewPolyInts(0, -1, 2), GE), NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyInts(0, 0, 0, -1), NewInt(0), NewInt(-1)), NewPolyCoef(1, NewPolyInts(0, 0, -2, 3), NewInt(0), NewInt(3))), LT), NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyInts(0, 0, 2, -2), NewInt(0), NewInt(-2)), NewPolyCoef(1, NewPolyInts(0, 1, -4, 3), NewInt(0), NewInt(3))), GT)))

	return q
}

func exEasy7() *QeExample {
	// ex([x], a*x^2+b*x+c == 0)
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{2}, newFmlAnds(NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyInts(0, 0, -1), NewInt(-2)), NewInt(1)), EQ), NewAtom(NewPolyInts(2, -125, 0, 1), EQ), NewAtom(NewPolyInts(2, 0, 1), GT), NewAtom(NewPolyCoef(1, NewPolyInts(0, -25, 0, 1), NewInt(0), NewInt(1)), LE), NewAtom(NewPolyCoef(1, NewPolyInts(0, -50, -20, -2), NewPolyInts(0, -5, -1), NewInt(1)), LE)))

	q.Output = newFmlAnds(NewAtom(NewPolyInts(0, 0, 1), GT), NewAtom(NewPolyCoef(1, NewPolyInts(0, -25, 0, 1), NewInt(0), NewInt(1)), EQ), NewAtom(NewPolyCoef(1, NewPolyInts(0, -125, 0, 1), NewPolyInts(0, 0, 4), NewInt(4)), EQ))
	q.Ref = "syn_pdq error"

	return q
}
func exMakePdf() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{1}, newFmlAnds(NewAtom(NewPolyCoef(1, NewPolyInts(0, -1, 0, 1), NewInt(0), NewInt(1)), EQ), NewAtom(NewPolyCoef(1, NewPolyInts(0, 0, 1), NewInt(1)), LT)))
	q.Ref = "Christopher W. Brown. Solution formula construction for truth invariant CAD's. Thesis p65 1999"

	return q
}

func exMakePdf2() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{2}, newFmlAnds(NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyInts(0, -1, 0, 1), NewInt(0), NewInt(1)), NewInt(0), NewInt(1)), EQ), NewAtom(NewPolyCoef(2, NewPolyCoef(1, NewPolyInts(0, -1, 1), NewInt(2)), NewInt(1)), LT)))
	q.Ref = "Christopher W. Brown. Solution formula construction for truth invariant CAD's. Thesis p65 1999"

	return q
}

func exPL01() *QeExample {
	q := new(QeExample)
	q.Output = NewAtom(NewPolyInts(0, 0, 1), LE)
	q.Input = NewQuantifier(true, []Level{1, 2}, newFmlOrs(NewAtom(NewPolyInts(1, 1, 1), LT), NewAtom(NewPolyInts(1, -1, 1), GT), NewAtom(NewPolyInts(2, 1, 1), LT), NewAtom(NewPolyInts(2, -1, 1), GT), NewAtom(NewPolyCoef(2, NewPolyInts(0, 1, -1), NewInt(0), NewPolyInts(1, 0, 0, -3, 0, 1), NewInt(0), NewPolyInts(1, 0, 0, 1)), GE)))
	q.Ref = "P. Parrilo and S. Lall. Semidefinite Programming Relaxation and Algebraic Optimization in Control."
	q.DOI = ""
	return q
}

func exQuad() *QeExample {
	// ex([x], a*x^2+b*x+c == 0)
	q := new(QeExample)
	q.Input = NewQuantifier(false, []Level{3}, NewAtom(NewPolyCoef(3, NewPolyInts(2, 0, 1), NewPolyInts(1, 0, 1), NewPolyInts(0, 0, 1)), EQ))
	q.Output = newFmlOrs(newFmlAnds(NewAtom(NewPolyCoef(2, NewPolyInts(1, 0, 0, -1), NewPolyInts(0, 0, 4)), LE), NewAtom(NewPolyInts(0, 0, 1), NE)), newFmlAnds(NewAtom(NewPolyInts(0, 0, 1), EQ), NewAtom(NewPolyInts(1, 0, 1), NE)), NewAtom(NewPolyInts(2, 0, 1), EQ))

	return q
}

func exWO1() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{3}, NewAtom(NewPolyCoef(3, NewInt(1), NewPolyInts(2, 0, 1), NewPolyInts(1, 0, 1), NewInt(0), NewInt(0), NewInt(0), NewPolyInts(0, 0, 1)), GT))
	return q
}

func exWO2() *QeExample {
	q := new(QeExample)
	q.Input = NewQuantifier(true, []Level{3}, NewAtom(NewPolyCoef(3, NewPolyCoef(1, NewInt(0), NewPolyInts(0, 0, -1)), NewPolyCoef(2, NewInt(0), NewPolyInts(1, 0, 1)), NewInt(1)), GE))
	return q
}
