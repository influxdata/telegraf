#!/bin/sh

set -eux

GO_VERSION="1.18.1"

setup_go () {
    choco upgrade golang --allow-downgrade --version=${GO_VERSION}
    choco install make
    git config --system core.longpaths true
    rm -rf /c/Go
    cp -r /c/Program\ Files/Go /c/
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
