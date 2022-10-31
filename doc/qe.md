# Quantifier Elimination

## QE algorithm

| algorithm | implementation | citation |
| :-- | :--: | :--: |
| [Cylindrical algebraic decomposition](../cad.go) | ✔ | [cad](cad.md) |
| [Linear virtual substitution](../vs.go) | ✔ | [Weispfenning88](https://www.sciencedirect.com/science/article/pii/S0747717188800038) |
| Quadratic virtual substitution | | [Weispfenning97](https://link.springer.com/article/10.1007/s002000050055) |
| Cubic virtual substitution |  | [Weispfenning94](https://dl.acm.org/doi/10.1145/190347.190425) |
| [Linear equational constraints](../quadeq.go) `ex([x], a*x+b==0 && phi)` | ✔ | [Hong93](https://dl.acm.org/doi/10.1145/164081.164140) |
| [Quadratic equational constraints](../quadeq.go) `ex([x], a*x^2+b*x+c==0 && phi)` | ✔ | [Hong93](https://dl.acm.org/doi/10.1145/164081.164140) |
| Root `all([x], f(x) > 0)` | | [GonzalezVega98](https://link.springer.com/chapter/10.1007/978-3-7091-9459-1_19) |
| Sign definite condition `all([x], x >= 0 && f(x) > 0)` | | [Anai07](https://www.tandfonline.com/doi/abs/10.1080/00207170600726550?journalCode=tcon20), [Iwane13](https://link.springer.com/chapter/10.1007/978-3-319-02297-0_17) |
| [Inequational constraints](../neq.go) `ex([x], f1 != 0 && f2 != 0 && ...)` | ✔ | [Iwane15](https://repository.kulib.kyoto-u.ac.jp/dspace/bitstream/2433/224375/1/1976-06.pdf) |
| Comprehensive Groebner systems || [Weispfenning98](https://link.springer.com/chapter/10.1007/978-3-7091-9459-1_20), [Fukasaku15](https://dl.acm.org/doi/10.1145/2755996.2756646) |


## Simplification

| algorithm | implementation | citation |
| :-- | :--: | :--: |
| [Basic](../simpl_basic.go) |✔| [Dolzmann97](https://www.sciencedirect.com/science/article/pii/S0747717197901231) |
| [Factorization](../simpl_fctr.go) |✔| [Dolzmann97](https://www.sciencedirect.com/science/article/pii/S0747717197901231) |
| [Equotional constraints](../simpl_reduce.go) |✔|
| [Even formula](../even.go) `phi(x^2) <=> ex([y], x^2 = y /\ phi(y) /\ y >= 0)` ||
| [Scale invaiant formula](../simpl_homo.go) |✔| [Iwane17](https://dl.acm.org/doi/abs/10.1145/3087604.3087627) |
| Translation invariant formula || [Iwane17](https://dl.acm.org/doi/abs/10.1145/3087604.3087627) |
| Rotation invariant formula || [Iwane17](https://dl.acm.org/doi/abs/10.1145/3087604.3087627) |
| [Symbolic-numeric](../simpl_num.go) |✔| [Iwane18](http://www.jssac.org/Editor/Suushiki/V24/V242.html) |
