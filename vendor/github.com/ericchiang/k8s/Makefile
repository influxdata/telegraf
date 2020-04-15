KUBE_VERSION=1.10.1

build:
	go build -v ./...

test:
	go test -v ./...

test-examples:
	@for example in $(shell find examples/ -name '*.go'); do \
		go build -v $$example || exit 1; \
	done

.PHONY: generate
generate: _output/kubernetes _output/bin/protoc _output/bin/gomvpkg _output/bin/protoc-gen-gofast _output/src/github.com/golang/protobuf
	./scripts/generate.sh
	go run scripts/register.go
	cp scripts/json.go.partial apis/meta/v1/json.go

.PHONY: verify-generate
verify-generate: generate
	./scripts/git-diff.sh

_output/bin/protoc-gen-gofast:
	./scripts/go-install.sh \
		https://github.com/gogo/protobuf \
		github.com/gogo/protobuf \
		github.com/gogo/protobuf/protoc-gen-gofast \
		tags/v0.5

_output/bin/gomvpkg:
	./scripts/go-install.sh \
		https://github.com/golang/tools \
		golang.org/x/tools \
		golang.org/x/tools/cmd/gomvpkg \
		fbec762f837dc349b73d1eaa820552e2ad177942

_output/src/github.com/golang/protobuf:
	git clone https://github.com/golang/protobuf _output/src/github.com/golang/protobuf

_output/bin/protoc:
	./scripts/get-protoc.sh

_output/kubernetes:
	mkdir -p _output
	curl -o _output/kubernetes.zip -L https://github.com/kubernetes/kubernetes/archive/v$(KUBE_VERSION).zip
	unzip _output/kubernetes.zip -d _output > /dev/null
	mv _output/kubernetes-$(KUBE_VERSION) _output/kubernetes

.PHONY: clean
clean:
	rm -rf _output
