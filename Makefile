UNAME := $(shell sh -c 'uname')
VERSION := $(shell sh -c 'git describe --always --tags')

build: prepare
		CGO_ENABLED=0 $(GOPATH)/bin/godep go build -a -installsuffix cgo \
		-o telegraf \
		-ldflags \
		"-X main.Version $(VERSION)" \
		./cmd/telegraf/telegraf.go

prepare:
	go get github.com/tools/godep

docker-compose:
ifeq ($(UNAME), Darwin)
	ADVERTISED_HOST=$(shell sh -c 'boot2docker ip') docker-compose up -d
endif
ifeq ($(UNAME), Linux)
	ADVERTISED_HOST=localhost docker-compose up -d
endif

test: prepare docker-compose
	$(GOPATH)/bin/godep go test -v ./...

test-short: prepare
	$(GOPATH)/bin/godep go test -v -short ./...

test-cleanup:
	docker-compose kill

.PHONY: test
