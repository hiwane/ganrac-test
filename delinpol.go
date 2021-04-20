package ganrac

import (
	"fmt"
	"os"
)

func (cad *CAD) projmc_vanish(cell *Cell, pf *ProjFactor) bool {
	cell.Print(os.Stdout, "cellp")
	fmt.Printf("%v\n", pf.p)
	for c := cell; c.lv >= 0; c = c.parent {
		if c.index%2 == 0 {
		}
	}
	return true
}
