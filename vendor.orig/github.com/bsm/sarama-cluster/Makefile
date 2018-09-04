SCALA_VERSION?= 2.12
KAFKA_VERSION?= 1.0.0
KAFKA_DIR= kafka_$(SCALA_VERSION)-$(KAFKA_VERSION)
KAFKA_SRC= https://archive.apache.org/dist/kafka/$(KAFKA_VERSION)/$(KAFKA_DIR).tgz
KAFKA_ROOT= testdata/$(KAFKA_DIR)
PKG=$(shell go list ./... | grep -v vendor)

default: vet test

vet:
	go vet $(PKG)

test: testdeps
	KAFKA_DIR=$(KAFKA_DIR) go test $(PKG) -ginkgo.slowSpecThreshold=60

test-verbose: testdeps
	KAFKA_DIR=$(KAFKA_DIR) go test $(PKG) -ginkgo.slowSpecThreshold=60 -v

test-race: testdeps
	KAFKA_DIR=$(KAFKA_DIR) go test $(PKG) -ginkgo.slowSpecThreshold=60 -v -race

testdeps: $(KAFKA_ROOT)

doc: README.md

.PHONY: test testdeps vet doc

# ---------------------------------------------------------------------

$(KAFKA_ROOT):
	@mkdir -p $(dir $@)
	cd $(dir $@) && curl -sSL $(KAFKA_SRC) | tar xz

README.md: README.md.tpl $(wildcard *.go)
	becca -package $(subst $(GOPATH)/src/,,$(PWD))
