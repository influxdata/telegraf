#!/bin/sh

set -eux

GO_VERSION="1.27.1"
GO_ARCH="linux-amd64"
# from https://go.dev/dl
GO_VERSION_SHA="63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445"

# Download Go and verify Go tarball
setup_go () {
    echo "installing go"
    curl -L https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz --output go${GO_VERSION}.${GO_ARCH}.tar.gz
    if ! echo "${GO_VERSION_SHA}  go${GO_VERSION}.${GO_ARCH}.tar.gz" | shasum --algorithm 256 --check -; then
        echo "Checksum failed" >&2
        exit 1
    fi

    sudo rm -rfv /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.${GO_ARCH}.tar.gz
}

if command -v go >/dev/null 2>&1; then
    echo "Go is already installed"
    cd
    v=$(go version | { read -r _ _ v _; echo "${v#go}"; })
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
    fi
else
    setup_go
fi

echo "$PATH"
command -v go
go version
