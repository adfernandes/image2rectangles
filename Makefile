.PHONY: test clean

prg=rect

test: clean _rect
	./_rect -i x.png -o _x.svg -n
	convert -loop 0 -delay 20 _zzz/_x.svg~*.png _x.gif
	./_rect -i o.png -o _o.svg -n
	convert -loop 0 -delay 20 _zzz/_o.svg~*.png _o.gif

clean:
	rm -rfv _*
	mkdir -p _zzz

_rect: rect.go
	GOPATH="$(PWD)" go build rect.go
	mv -v rect _rect
