package ganrac

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
	{"pl01", exPL01},
}

func GetExampleFof(name string) *QeExample {
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

func exPL01() *QeExample {
	q := new(QeExample)
	q.Output = NewAtom(NewPolyInts(0, 0, 1), LE)
	q.Input = NewQuantifier(true, []Level{1, 2}, newFmlOrs(NewAtom(NewPolyInts(1, 1, 1), LT), NewAtom(NewPolyInts(1, -1, 1), GT), NewAtom(NewPolyInts(2, 1, 1), LT), NewAtom(NewPolyInts(2, -1, 1), GT), NewAtom(NewPolyCoef(2, NewPolyInts(0, 1, -1), NewInt(0), NewPolyInts(1, 0, 0, -3, 0, 1), NewInt(0), NewPolyInts(1, 0, 0, 1)), GE)))
	q.Ref = "P. Parrilo and S. Lall. Semidefinite Programming Relaxation and Algebraic Optimization in Control."
	q.DOI = ""
	return q
}
