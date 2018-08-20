.PHONY: test clean

prg=rect

test: clean _rect
	./_rect < x.png -svg _x.svg -invert -negative _x~neg.png -animation-build _x~b.gif -animation-pixels _x~p.gif -report
	./_rect < o.png -svg _o.svg -invert -negative _o~neg.png -animation-build _o~b.gif -animation-pixels _o~p.gif -report
	./_rect < z.png -svg _z.svg -invert -negative _z~neg.png -animation-build _z~b.gif -animation-pixels _z~p.gif -report

clean:
	rm -rfv _*

_rect: rect.go
	GOPATH="$(PWD)" go build rect.go
	mv -v rect _rect
