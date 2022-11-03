GaNRAC
======

## Getting started

### Requirements

- [ox-asir](http://www.math.sci.kobe-u.ac.jp/OpenXM/)

CentOS
```sh
# yum install libX11-devel libXt-devel libXaw-devel
> curl -O "http://www.math.sci.kobe-u.ac.jp/pub/OpenXM/Head/openxm-head.tar.gz"
> tar xf openxm-head.tar.gz
> (cd OpenXM/src; make install)
> (cd OpenXM/rc; make install)
> while :; do ox -ox ox_asir -control 1234 -data 4321; done
```

Ubuntu
```sh
# apt install build-essential m4 bison
# apt install libx11-dev libxt-dev libxaw7-dev
> curl -O "http://www.math.sci.kobe-u.ac.jp/pub/OpenXM/Head/openxm-head.tar.gz"
> tar xf openxm-head.tar.gz
> (cd OpenXM/src; make install)
> (cd OpenXM/rc; make install)
> while :; do ox -ox ox_asir -control 1234 -data 4321; done
```

## Installation

```sh
> go get github.com/hiwane/ganrac/cmd/ganrac
```

## Demo

![ganrac9](https://user-images.githubusercontent.com/7787544/123178824-fc812c80-d4c2-11eb-8c5a-3cb209b83478.gif)

### real Quantifier Elimination

See [qe.md](doc/qe.md) for details.

![ganrac7](https://user-images.githubusercontent.com/7787544/122847029-0891b080-d342-11eb-84ab-f085f5bbaad6.gif)

```
F = ex([x], a*x^2+b*x+c==0);
qe(F);
G = example("quad");
H = time(qe(G[0]));
qe(all([x,y,z], equiv(G[1], H)));
```

### real QE by Cylindrical Algebraic Decomposition

See [cad.md](doc/cad.md) for details.

![ganrac8](https://user-images.githubusercontent.com/7787544/122847006-fdd71b80-d341-11eb-8156-8a0e5f49b535.gif)

```
vars(c,b,a,x);
F = ex([x], a*x^2+b*x+c <= 0);
F;
C = cadinit(F);
cadproj(C);
print(C, "proj");
print(C, "proj", 3, 0);
cadlift(C);
print(C, "sig");
print(C, "sig", 0);
print(C, "cell", 0, 1);
cadsfc(C);
print(C, "stat");
```

GaGAGAGA
