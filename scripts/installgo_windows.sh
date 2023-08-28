#!/bin/sh

set -eux

GO_VERSION="1.21.0"

setup_go () {
    choco upgrade golang --allow-downgrade --version=${GO_VERSION}
    choco install make
    git config --system core.longpaths true
}

if command -v go >/dev/null 2>&1; then
    echo "Go is already installed"
    v=$(go version | { read -r _ _ v _; echo "${v#go}"; })
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
        go version
    fi
else
    setup_go
fi
