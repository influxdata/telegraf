all: test integration

test:
	go test ./...

integration:
	go test -v -tags=integration ./uatest/...

install-py-opcua:
	pip3 install opcua

gen:
	GOMODULES111=on go get -u golang.org/x/tools/cmd/stringer
	go generate ./...

release:
	GITHUB_TOKEN=$$(security find-generic-password -gs GITHUB_TOKEN -w) goreleaser --rm-dist
