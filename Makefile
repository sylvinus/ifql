SUBDIRS := ifql
TOPTARGETS := all clean

$(TOPTARGETS): $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

test:
	go test `go list ./... | grep -v /vendor/` ${args}

.PHONY: $(TOPTARGETS) $(SUBDIRS)
