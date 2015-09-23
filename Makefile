UNAME := $(shell sh -c 'uname')
VERSION := $(shell sh -c 'git describe --always --tags')
ifndef GOBIN
	GOBIN = $(GOPATH)/bin
endif

build: prepare
	$(GOBIN)/godep go build -o telegraf -ldflags \
		"-X main.Version=$(VERSION)" \
		./cmd/telegraf/telegraf.go

build-linux-bins: prepare
	GOARCH=amd64 GOOS=linux $(GOBIN)/godep go build -o telegraf_linux_amd64 \
                     -ldflags "-X main.Version=$(VERSION)" \
                     ./cmd/telegraf/telegraf.go
	GOARCH=386 GOOS=linux $(GOBIN)/godep go build -o telegraf_linux_386 \
                     -ldflags "-X main.Version=$(VERSION)" \
                     ./cmd/telegraf/telegraf.go
	GOARCH=arm GOOS=linux $(GOBIN)/godep go build -o telegraf_linux_arm \
                     -ldflags "-X main.Version=$(VERSION)" \
                     ./cmd/telegraf/telegraf.go

prepare:
	go get github.com/tools/godep

docker-compose:
ifeq ($(UNAME), Darwin)
	ADVERTISED_HOST=$(shell sh -c 'boot2docker ip || docker-machine ip default') \
		docker-compose --file scripts/docker-compose.yml up -d
endif
ifeq ($(UNAME), Linux)
	ADVERTISED_HOST=localhost docker-compose --file scripts/docker-compose.yml up -d
endif

test: prepare docker-compose
	$(GOBIN)/godep go test ./...

test-short: prepare
	$(GOBIN)/godep go test -short ./...

test-cleanup:
	docker-compose --file scripts/docker-compose.yml kill

.PHONY: test
