PREFIX := /usr/local
VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l $(filter-out plugins/parsers/influx/machine.go, $(GOFILES)))
BUILDFLAGS ?=

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
	LDFLAGS += -X main.version=$(VERSION)
endif

all:
	$(MAKE) deps
	$(MAKE) telegraf

deps:
	go get -u github.com/golang/lint/golint
	go get github.com/sparrc/gdm
	gdm restore

telegraf:
	go build -i -ldflags "$(LDFLAGS)" ./cmd/telegraf

go-install:
	go install -ldflags "-w -s $(LDFLAGS)" ./cmd/telegraf

install: telegraf
	mkdir -p $(DESTDIR)$(PREFIX)/bin/
	cp $(TELEGRAF) $(DESTDIR)$(PREFIX)/bin/

test:
	go test -short ./...

fmt:
	@gofmt -w $(filter-out plugins/parsers/influx/machine.go, $(GOFILES))

fmtcheck:
	@echo '[INFO] running gofmt to identify incorrectly formatted code...'
	@if [ ! -z "$(GOFMT)" ]; then \
		echo "[ERROR] gofmt has found errors in the following files:"  ; \
		echo "$(GOFMT)" ; \
		echo "" ;\
		echo "Run make fmt to fix them." ; \
		exit 1 ;\
	fi
	@echo '[INFO] done.'

test-windows:
	go test ./plugins/inputs/ping/...
	go test ./plugins/inputs/win_perf_counters/...
	go test ./plugins/inputs/win_services/...
	go test ./plugins/inputs/procstat/...
	go test ./plugins/inputs/ntpq/...

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@echo 'go vet $$(go list ./... | grep -v ./plugins/parsers/influx)'
	@go vet $$(go list ./... | grep -v ./plugins/parsers/influx) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "go vet has found suspicious constructs. Please remediate any reported errors"; \
		echo "to fix them before submitting code for review."; \
		exit 1; \
	fi

test-ci: fmtcheck vet
	go test -short ./...

test-all: fmtcheck vet
	go test ./...

package:
	./scripts/build.py --package --platform=all --arch=all

clean:
	rm -f telegraf
	rm -f telegraf.exe

docker-image:
	./scripts/build.py --package --platform=linux --arch=amd64
	cp build/telegraf*$(COMMIT)*.deb .
	docker build -f scripts/dev.docker --build-arg "package=telegraf*$(COMMIT)*.deb" -t "telegraf-dev:$(COMMIT)" .

plugins/parsers/influx/machine.go: plugins/parsers/influx/machine.go.rl
	ragel -Z -G2 $^ -o $@

.PHONY: deps telegraf install test test-windows lint vet test-all package clean docker-image fmtcheck uint64
