SUBDIRS := ast ifql promql

SOURCES := $(shell find . -name '*.go') 

all: Gopkg.lock $(SUBDIRS) bin/ifql

$(SUBDIRS): bin/pigeon bin/cmpgen
	$(MAKE) -C $@ $(MAKECMDGOALS)

bin/ifql: $(SOURCES) bin/pigeon bin/cmpgen
	go build -i -o bin/ifql ./cmd/ifql

bin/pigeon: ./vendor/github.com/mna/pigeon/main.go
	go build -i -o bin/pigeon  ./vendor/github.com/mna/pigeon

bin/cmpgen: ./ast/asttest/cmpgen/main.go
	go build -i -o bin/cmpgen ./ast/asttest/cmpgen

Gopkg.lock: Gopkg.toml
	dep ensure

update:
	dep ensure -update

test: Gopkg.lock bin/ifql
	go test ./...

test-race: Gopkg.lock bin/ifql
	go test -race ./...

bench: Gopkg.lock bin/ifql
	go test -bench=. -run=^$$ ./...

clean: $(SUBDIRS)
	rm -rf bin

.PHONY: all clean $(SUBDIRS) update
