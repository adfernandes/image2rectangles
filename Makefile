.PHONY: test clean

prg=rect

test: _rect
	./_rect -i x.png -o _x.svg -n
	./_rect -i o.png -o _o.svg -n

clean:
	rm -fv _*

_rect: rect.go
	GOPATH="$(PWD)" go build rect.go
	mv -v rect _rect
