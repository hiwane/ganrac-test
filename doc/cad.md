# Cylindrical Algebraic Decomposition

## Projection Operator

| algorithm | implementation | citation |
| :-- | :--: | :--: |
| Collins' projection  | |
| [Hong's projection](../projhh.go) | ✔ |
| [McCallum's projection](../projmc.go) | ✔ |
| Lazard's projection |  |


```
> vars(a,b,s,t,u,x);
0
> F = ex([x], a*x+b != 0 && s*x^2+t*x+u <= 0);
ex([x], a*x+b!=0 && s*x^2+t*x+u<=0)
> cad(F); # McCallum Projection
go projalgo=0, lv=0
error: NOT well-oriented
> cad(F, 1); # Hong Projection
go projalgo=1, lv=0
(4*s*u-t^2<=0 && a*t-2*b*s!=0) || (a!=0 && s==0 && u==0) || (b!=0 && 4*s*u-t^2<0) || (b!=0 && u<=0) || (a!=0 && s<0) || a^2*u-a*b*t+b^2*s<0
```

## Lifting

| algorithm | implementation | citation |
| :-- | :--: | :--: |
| [Symbolic-numeric CAD](../lift.go) | ✔| [1](https://www.sciencedirect.com/science/article/pii/S0304397512009413) |
| [Dynamic evaluation](../cad_de.go) | ✔| [1](https://dl.acm.org/doi/10.1006/jsco.1994.1057), [2](https://www.semanticscholar.org/paper/About-a-New-Method-for-Computing-in-Algebraic-Dora-Dicrescenzo/2ebef9590ca6ce106a45f491b0b864aa5a2206c2), [3](https://www.sciencedirect.com/science/article/pii/S0304397512009413) |
| Local projection | | [1](https://dl.acm.org/doi/10.1145/2608628.2608633) |

## Soluation Formula Construction

- [Solution formula construction](https://dl.acm.org/doi/10.5555/929495)


## Demo

![cad](https://user-images.githubusercontent.com/7787544/199652778-84d4a90b-4906-4962-ac51-71b625bd9043.gif)

