.PHONY: test clean

prg=rect

test: clean _rect
	./_rect < x.png -svg _x.svg -invert -negative _x~neg.png

clean:
	rm -rfv _*

_rect: rect.go
	GOPATH="$(PWD)" go build rect.go
	mv -v rect _rect
