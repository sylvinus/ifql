all: promql.go

promql.go: promql.peg
	go generate -x ./...

.PHONY: all clean
