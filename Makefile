next_version :=  $(shell cat build_version.txt)
tag := $(shell git describe --exact-match --tags 2>git_describe_error.tmp; rm -f git_describe_error.tmp)
branch := $(shell git rev-parse --abbrev-ref HEAD)
commit := $(shell git rev-parse --short=8 HEAD)
glibc_version := 2.17

MAKEFLAGS += --no-print-directory
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
HOSTGO := env -u GOOS -u GOARCH -u GOARM -- go

LDFLAGS := $(LDFLAGS) -X main.commit=$(commit) -X main.branch=$(branch) -X main.goos=$(GOOS) -X main.goarch=$(GOARCH)
ifneq ($(tag),)
	LDFLAGS += -X main.version=$(version)
endif

# Go built-in race detector works only for 64 bits architectures.
ifneq ($(GOARCH), 386)
	race_detector := -race
endif


GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l -s $(filter-out plugins/parsers/influx/machine.go, $(GOFILES)))

prefix ?= /usr/local
bindir ?= $(prefix)/bin
sysconfdir ?= $(prefix)/etc
localstatedir ?= $(prefix)/var
pkgdir ?= build/dist

.PHONY: all
all:
	@$(MAKE) deps
	@$(MAKE) telegraf

.PHONY: help
help:
	@echo 'Targets:'
	@echo '  all        - download dependencies and compile telegraf binary'
	@echo '  deps       - download dependencies'
	@echo '  telegraf   - compile telegraf binary'
	@echo '  test       - run short unit tests'
	@echo '  fmt        - format source files'
	@echo '  tidy       - tidy go modules'
	@echo '  check-deps - check docs/LICENSE_OF_DEPENDENCIES.md'
	@echo '  clean      - delete build artifacts'
	@echo ''
	@echo 'Package Targets:'
	@$(foreach dist,$(dists),echo "  $(dist)";)

.PHONY: deps
deps:
	go mod download -x

.PHONY: telegraf
telegraf:
	go build -ldflags "$(LDFLAGS)" ./cmd/telegraf

# Used by dockerfile builds
.PHONY: go-install
go-install:
	go install -mod=mod -ldflags "-w -s $(LDFLAGS)" ./cmd/telegraf

.PHONY: test
test:
	go test -short $(race_detector) ./...

.PHONY: test-integration
test-integration:
	go test -run Integration $(race_detector) ./...

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
	go test -short ./...

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
	go test $(race_detector) ./...

.PHONY: check-deps
check-deps:
	./scripts/check-deps.sh

.PHONY: clean
clean:
	rm -f telegraf
	rm -f telegraf.exe
	rm -rf build

.PHONY: docker-image
docker-image:
	docker build -f scripts/buster.docker -t "telegraf:$(commit)" .

plugins/parsers/influx/machine.go: plugins/parsers/influx/machine.go.rl
	ragel -Z -G2 $^ -o $@

.PHONY: plugin-%
plugin-%:
	@echo "Starting dev environment for $${$(@)} input plugin..."
	@docker-compose -f plugins/inputs/$${$(@)}/dev/docker-compose.yml up

.PHONY: ci-1.15
ci-1.15:
	docker build -t quay.io/influxdb/telegraf-ci:1.15.5 - < scripts/ci-1.15.docker
	docker push quay.io/influxdb/telegraf-ci:1.15.5

.PHONY: ci-1.14
ci-1.14:
	docker build -t quay.io/influxdb/telegraf-ci:1.14.9 - < scripts/ci-1.14.docker
	docker push quay.io/influxdb/telegraf-ci:1.14.9

.PHONY: install
install: $(buildbin)
	@mkdir -pv $(DESTDIR)$(bindir)
	@mkdir -pv $(DESTDIR)$(sysconfdir)
	@mkdir -pv $(DESTDIR)$(localstatedir)
	@if [ $(GOOS) != "windows" ]; then mkdir -pv $(DESTDIR)$(sysconfdir)/logrotate.d; fi
	@if [ $(GOOS) != "windows" ]; then mkdir -pv $(DESTDIR)$(localstatedir)/log/telegraf; fi
	@if [ $(GOOS) != "windows" ]; then mkdir -pv $(DESTDIR)$(sysconfdir)/telegraf/telegraf.d; fi
	@cp -fv $(buildbin) $(DESTDIR)$(bindir)
	@if [ $(GOOS) != "windows" ]; then cp -fv etc/telegraf.conf $(DESTDIR)$(sysconfdir)/telegraf/telegraf.conf$(conf_suffix); fi
	@if [ $(GOOS) != "windows" ]; then cp -fv etc/logrotate.d/telegraf $(DESTDIR)$(sysconfdir)/logrotate.d; fi
	@if [ $(GOOS) = "windows" ]; then cp -fv etc/telegraf_windows.conf $(DESTDIR)/telegraf.conf; fi
	@if [ $(GOOS) = "linux" ]; then scripts/check-dynamic-glibc-versions.sh $(buildbin) $(glibc_version); fi
	@if [ $(GOOS) = "linux" ]; then mkdir -pv $(DESTDIR)$(prefix)/lib/telegraf/scripts; fi
	@if [ $(GOOS) = "linux" ]; then cp -fv scripts/telegraf.service $(DESTDIR)$(prefix)/lib/telegraf/scripts; fi
	@if [ $(GOOS) = "linux" ]; then cp -fv scripts/init.sh $(DESTDIR)$(prefix)/lib/telegraf/scripts; fi

# Telegraf build per platform.  This improves package performance by sharing
# the bin between deb/rpm/tar packages over building directly into the package
# directory.
$(buildbin):
	@mkdir -pv $(dir $@)
	go build -o $(dir $@) -ldflags "$(LDFLAGS)" ./cmd/telegraf

tars := telegraf-$(tar_version)_darwin_amd64.tar.gz

dists := $(tars)

.PHONY: package
package: $(dists)

rpm_amd64 := amd64
rpm_386 := i386
rpm_s390x := s390x
rpm_ppc64le := ppc64le
rpm_arm5 := armel
rpm_arm6 := armv6hl
rpm_arm647 := aarch64
rpm_arch = $(rpm_$(GOARCH)$(GOARM))

.PHONY: $(rpms)
$(rpms):
	@$(MAKE) install
	@mkdir -p $(pkgdir)
	fpm --force \
		--log info \
		--architecture $(rpm_arch) \
		--input-type dir \
		--output-type rpm \
		--vendor InfluxData \
		--url https://github.com/influxdata/telegraf \
		--license MIT \
		--maintainer support@influxdb.com \
		--config-files /etc/telegraf/telegraf.conf \
		--config-files /etc/logrotate.d/telegraf \
		--after-install scripts/rpm/post-install.sh \
		--before-install scripts/rpm/pre-install.sh \
		--after-remove scripts/rpm/post-remove.sh \
		--description "Plugin-driven server agent for reporting metrics into InfluxDB." \
		--depends coreutils \
		--depends shadow-utils \
		--rpm-posttrans scripts/rpm/post-install.sh \
		--name telegraf \
		--version $(version) \
		--iteration $(rpm_iteration) \
        --chdir $(DESTDIR) \
		--package $(pkgdir)/$@

deb_amd64 := amd64
deb_386 := i386
deb_s390x := s390x
deb_ppc64le := ppc64el
deb_arm5 := armel
deb_arm6 := armhf
deb_arm647 := arm64
deb_mips := mips
deb_mipsle := mipsel
deb_arch = $(deb_$(GOARCH)$(GOARM))

.PHONY: $(debs)
$(debs):
	@$(MAKE) install
	@mkdir -pv $(pkgdir)
	fpm --force \
		--log info \
		--architecture $(deb_arch) \
		--input-type dir \
		--output-type deb \
		--vendor InfluxData \
		--url https://github.com/influxdata/telegraf \
		--license MIT \
		--maintainer support@influxdb.com \
		--config-files /etc/telegraf/telegraf.conf.sample \
		--config-files /etc/logrotate.d/telegraf \
		--after-install scripts/deb/post-install.sh \
		--before-install scripts/deb/pre-install.sh \
		--after-remove scripts/deb/post-remove.sh \
		--before-remove scripts/deb/pre-remove.sh \
		--description "Plugin-driven server agent for reporting metrics into InfluxDB." \
		--name telegraf \
		--version $(version) \
		--iteration $(deb_iteration) \
		--chdir $(DESTDIR) \
		--package $(pkgdir)/$@

.PHONY: $(tars)
$(tars):
	@$(MAKE) install
	@mkdir -p $(pkgdir)
	tar --owner 0 --group 0 -czvf $(pkgdir)/$@ -C $(dir $(DESTDIR)) .

%darwin_amd64.tar.gz: export GOOS := darwin
%darwin_amd64.tar.gz: export GOARCH := amd64

%.tar.gz: export pkg := tar
%.tar.gz: export prefix := /usr
%.tar.gz: export sysconfdir := /etc
%.tar.gz: export localstatedir := /var

%.deb %.rpm %.tar.gz %.zip: export DESTDIR = build/$(GOOS)-$(GOARCH)$(GOARM)$(cgo)-$(pkg)/telegraf-$(version)
%.deb %.rpm %.tar.gz %.zip: export buildbin = build/$(GOOS)-$(GOARCH)$(GOARM)$(cgo)/telegraf$(EXEEXT)
%.deb %.rpm %.tar.gz %.zip: export LDFLAGS = -w -s
