UNAME := $(shell sh -c 'uname')

ifeq ($(UNAME), Darwin)
	export ADVERTISED_HOST := $(shell sh -c 'boot2docker ip')
endif
ifeq ($(UNAME), Linux)
	export ADVERTISED_HOST := localhost
endif

prepare:
	godep go install ./...

docker-compose:
	docker-compose up -d

test: prepare docker-compose
	godep go test -v ./...

test-short: prepare
	godep go test -v -short ./...

test-cleanup:
	docker-compose kill

.PHONY: test
