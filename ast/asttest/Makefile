all: cmpopts.go

cmpopts.go: ../ast.go gen.go ../../bin/cmpgen
	PATH=../../bin:${PATH} go generate -x ./...

clean:
	rm -f cmpopts.go

.PHONY: all clean
