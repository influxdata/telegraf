# Makefile for the Telegraf classify processor plugin

# Where I have loaded Go temporarily for development work, because the OS copy is out of date.
# Set this as necessary to get this plugin built in your own development environment.
# GOPATH = /tmp/go
# The usual value, to access an OS-installed copy of Go:
GOPATH = /usr

default	: build

format	:
	env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" gofmt -w classify.go
	env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" gofmt -w classify_test.go

lint	:
	env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" staticcheck

# More aggressive linting.  Hopefully not enough to make a whole sweater.
# That said, not all recommendations made by the golangci-lint tool make
# sense, so we treat some of them like a yarn (fanciful story), and we
# disable a few specific checks in the code.
yarn	:
	env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" golangci-lint run

build	:
	env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" go build classify.go

test	:
	@# env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" go test -short
	env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" go test
	@# env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" go test -v
	@# env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" go test -v -failfast | more
	@# env PATH="${GOPATH}/bin:$$PATH" GOPATH="${GOPATH}" go test -v | more
