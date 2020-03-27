ifeq ($(OS), Windows_NT)
	VERSION := $(shell git describe --exact-match --tags 2>nul)
	HOME := $(HOMEPATH)
	CGO_ENABLED ?= 0
	export CGO_ENABLED
else
	VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
endif

PREFIX := /usr/local
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l -s $(filter-out plugins/parsers/influx/machine.go, $(GOFILES)))
BUILDFLAGS ?=

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(shell go env GOPATH))/bin:$(PATH)
endif

LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
	LDFLAGS += -X main.version=$(VERSION)
endif

.PHONY: all
all:
	@$(MAKE) --no-print-directory deps
	@$(MAKE) --no-print-directory telegraf

.PHONY: deps
deps:
	go mod download

.PHONY: telegraf
telegraf:
	go build -ldflags "$(LDFLAGS)" ./cmd/telegraf

.PHONY: go-install
go-install:
	go install -ldflags "-w -s $(LDFLAGS)" ./cmd/telegraf

.PHONY: install
install: telegraf
	mkdir -p $(DESTDIR)$(PREFIX)/bin/
	cp telegraf $(DESTDIR)$(PREFIX)/bin/

.PHONY: test
test:
	go test -short ./...

.PHONY: fmt
fmt:
	@gofmt -s -w $(filter-out plugins/parsers/influx/machine.go, $(GOFILES))

.PHONY: fmtcheck
fmtcheck:
	@if [ ! -z "$(GOFMT)" ]; then \
		echo "[ERROR] gofmt has found errors in the following files:"  ; \
		echo "$(GOFMT)" ; \
		echo "" ;\
		echo "Run make fmt to fix them." ; \
		exit 1 ;\
	fi

.PHONY: test-windows
test-windows:
	go test -short ./plugins/inputs/ping/...
	go test -short ./plugins/inputs/win_perf_counters/...
	go test -short ./plugins/inputs/win_services/...
	go test -short ./plugins/inputs/procstat/...
	go test -short ./plugins/inputs/ntpq/...

.PHONY: vet
vet:
	@echo 'go vet $$(go list ./... | grep -v ./plugins/parsers/influx)'
	@go vet $$(go list ./... | grep -v ./plugins/parsers/influx) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "go vet has found suspicious constructs. Please remediate any reported errors"; \
		echo "to fix them before submitting code for review."; \
		exit 1; \
	fi

.PHONY: tidy
tidy:
	go mod verify
	go mod tidy
	@if ! git diff --quiet go.mod go.sum; then \
		echo "please run go mod tidy and check in changes"; \
		exit 1; \
	fi

.PHONY: check
check: fmtcheck vet
	@$(MAKE) --no-print-directory tidy

.PHONY: test-all
test-all: fmtcheck vet
	go test ./...

.PHONY: check-deps
check-deps:
	./scripts/check-deps.sh

.PHONY: package
package:
	./scripts/build.py --package --platform=all --arch=all

.PHONY: package-release
package-release:
	./scripts/build.py --release --package --platform=all --arch=all \
		--upload --bucket=dl.influxdata.com/telegraf/releases

.PHONY: package-nightly
package-nightly:
	./scripts/build.py --nightly --package --platform=all --arch=all \
		--upload --bucket=dl.influxdata.com/telegraf/nightlies

.PHONY: clean
clean:
	rm -f telegraf
	rm -f telegraf.exe

.PHONY: docker-image
docker-image:
	docker build -f scripts/stretch.docker -t "telegraf:$(COMMIT)" .

plugins/parsers/influx/machine.go: plugins/parsers/influx/machine.go.rl
	ragel -Z -G2 $^ -o $@

.PHONY: static
static:
	@echo "Building static linux binary..."
	@CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" ./cmd/telegraf

.PHONY: plugin-%
plugin-%:
	@echo "Starting dev environment for $${$(@)} input plugin..."
	@docker-compose -f plugins/inputs/$${$(@)}/dev/docker-compose.yml up

.PHONY: ci-1.13
ci-1.13:
	docker build -t quay.io/influxdb/telegraf-ci:1.13.8 - < scripts/ci-1.13.docker
	docker push quay.io/influxdb/telegraf-ci:1.13.8

.PHONY: ci-1.12
ci-1.12:
	docker build -t quay.io/influxdb/telegraf-ci:1.12.17 - < scripts/ci-1.12.docker
	docker push quay.io/influxdb/telegraf-ci:1.12.17
