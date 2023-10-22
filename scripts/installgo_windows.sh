#!/bin/sh

set -eux

GO_VERSION="1.21.3"

setup_go () {
    choco upgrade golang --allow-downgrade --version=${GO_VERSION} --verbose
    git config --system core.longpaths true
}

echo "PATH before: $PATH"
echo "go location before: `command -v go`"

if command -v go >/dev/null 2>&1; then
    echo "Go is already installed"
    v=$(go version | { read -r _ _ v _; echo "${v#go}"; })
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
        go version
    fi
else
    echo "Setup go"
    setup_go
fi

echo "PATH after: $PATH"
echo "go location after: `command -v go`"
go version