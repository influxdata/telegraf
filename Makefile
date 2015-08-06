UNAME := $(shell sh -c 'uname')

ifeq ($(UNAME), Darwin)
	export ADVERTISED_HOST := $(shell sh -c 'boot2docker ip')
endif
ifeq ($(UNAME), Linux)
	export ADVERTISED_HOST := localhost
endif

prepare:
	go get -d -v -t ./...

docker-compose:
	docker-compose up -d

test: prepare docker-compose
	go test -v ./...

test-short: prepare
	go test -v -short ./...

test-cleanup:
	docker-compose kill

update:
	go get -u -v -d -t ./...

.PHONY: test
