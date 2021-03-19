package ganrac

var varlist = []string{
	"x", "y", "z", "w", "a", "b", "c", "e", "f", "g", "h",
}

func Add(x, y Coef) Coef {
	if x.Tag() >= y.Tag() {
		return x.Add(y)
	} else {
		return y.Add(x)
	}
}
