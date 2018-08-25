.PHONY: test clean

prg=image2rect

test: clean _$(prg)
	./_$(prg) < x.png -svg _x.svg -invert -negative _x~neg.png -animation-build _x~b.gif -animation-pixels _x~p.gif -report
	./_$(prg) < o.png -svg _o.svg -invert -negative _o~neg.png -animation-build _o~b.gif -animation-pixels _o~p.gif -report
	./_$(prg) < z.png -svg _z.svg -invert -negative _z~neg.png -animation-build _z~b.gif -animation-pixels _z~p.gif -report

clean:
	rm -rfv _*

_$(prg): $(prg).go
	GOPATH="$(PWD)" go build $(prg).go
	mv -v $(prg) _$(prg)
