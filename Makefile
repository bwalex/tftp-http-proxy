PROG=tftp-http-proxy
PACKAGE=bwalex/$(PROG)
SOURCEDIR=.

GO?=go
GOPATH = $(CURDIR)/.gopath
BASE = $(GOPATH)/src/$(PACKAGE)

SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

dist/$(PROG): $(SOURCES) | $(BASE)
	cd $(BASE) && GOPATH=$(GOPATH) $(GO) build -o $@

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

.PHONY: clean
clean:
	rm -f $(PROG)
