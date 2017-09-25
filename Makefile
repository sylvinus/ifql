SUBDIRS := ifql promql
TOPTARGETS := all clean

$(TOPTARGETS): $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

test:
	go test ./... ${args}

generate:
	go install ./vendor/github.com/mna/pigeon
	go generate ./... ${args}

.PHONY: $(TOPTARGETS) $(SUBDIRS)
