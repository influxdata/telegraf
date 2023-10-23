#!/bin/sh

set -ux

GO_VERSION="1.21.1"

setup_go () {
    choco upgrade golang --allow-downgrade --force --version=${GO_VERSION} --debug --verbose
    git config --system core.longpaths true
}

if command -v go >/dev/null 2>&1; then
    echo "Go is already installed"
    v=$(go version | { read -r _ _ v _; echo "${v#go}"; })
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
    fi
else
    setup_go
fi

command -v go
go version
