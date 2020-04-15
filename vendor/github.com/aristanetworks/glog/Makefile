GO := go
DEFAULT_GOPATH := $${GOPATH%%:*}
GOPATH_BIN := $(DEFAULT_GOPATH)/bin
GOLINT := $(GOPATH_BIN)/golint

all: check

lint:
	$(GOLINT) ./ &>lint && \
	if test -s lint; then cat lint; rm lint; exit 1; else rm lint; fi
# The above is ugly, but unfortunately golint doesn't exit 1 when it finds
# lint.  See https://github.com/golang/lint/issues/65

fmtcheck:
	if ! gofmt -l  . ; then echo Check the above file for coding style; exit 1; fi

test:
	PATH=$(GOPATH_BIN):$$PATH $(GO) test ./...

check: fmtcheck lint test


doc:
	godoc -http="localhost:6060" -play=true

.PHONY: all lint fmtcheck test check doc
