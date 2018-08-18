.PHONY: test clean

prg=rect

test: clean _rect
	./_rect < x.png -svg _x.svg -invert -negative _x~neg.png -animation _x~anim.gif -report
	./_rect < o.png -svg _o.svg -invert -negative _o~neg.png -animation _o~anim.gif -report

clean:
	rm -rfv _*

_rect: rect.go
	GOPATH="$(PWD)" go build rect.go
	mv -v rect _rect
