SUBDIRS := ast ifql promql

SOURCES := $(shell find . -name '*.go' -not -name '*_test.go')

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
	dep ensure -v

vendor/github.com/mna/pigeon/main.go: Gopkg.lock
	dep ensure -v

update:
	dep ensure -v -update

test: Gopkg.lock bin/ifql
	go test ./...

test-race: Gopkg.lock bin/ifql
	go test -race ./...

bench: Gopkg.lock bin/ifql
	go test -bench=. -run=^$$ ./...

bin/goreleaser:
	go build -i -o bin/goreleaser ./vendor/github.com/goreleaser/goreleaser

release: bin/goreleaser
	goreleaser --skip-publish  --snapshot --rm-dist

clean: $(SUBDIRS)
	rm -rf bin dist

.PHONY: all clean $(SUBDIRS) update test test-race bench release
