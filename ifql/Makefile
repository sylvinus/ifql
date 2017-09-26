all: ifql.go

ifql.go: ifql.peg gen.go ../bin/pigeon
	PATH=../bin:${PATH} go generate -x ./...

clean:
	rm -f ifql.go

.PHONY: all clean
