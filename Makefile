UNAME := $(shell sh -c 'uname')
VERSION := $(shell sh -c 'git describe --always --tags')

build: prepare
	$(GOPATH)/bin/godep go build -o telegraf -ldflags \
		"-X main.Version $(VERSION)" \
		./cmd/telegraf/telegraf.go

prepare:
	go get github.com/tools/godep

docker-compose:
	docker-compose up -d

test:
ifeq ($(UNAME), Darwin)
	ADVERTISED_HOST=$(shell sh -c 'boot2docker ip') $(MAKE) test-full
endif
ifeq ($(UNAME), Linux)
	ADVERTISED_HOST=localhost $(MAKE) test-full
endif

test-full: prepare docker-compose
	$(GOPATH)/bin/godep go test -v ./...

test-short: prepare
	$(GOPATH)/bin/godep go test -v -short ./...

test-cleanup:
	docker-compose kill

.PHONY: test
