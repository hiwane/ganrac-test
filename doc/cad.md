# Cylindrical Algebraic Decomposition

## Projection Operator

| algorithm | implementation | citation |
| :-- | :--: | :--: |
| Hong Projection | ✔ |
| McCallum Projection | ✔ |
| Lazard Projection |  |


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

## ...

- [x] [Symbolic-numeric CAD](https://www.sciencedirect.com/science/article/pii/S0304397512009413)
- [ ] [Local projection](https://dl.acm.org/doi/10.1145/2608628.2608633)
- [x] [Solution formula construction](https://dl.acm.org/doi/10.5555/929495)
