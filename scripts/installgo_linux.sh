#!/bin/sh

set -eux

GO_VERSION="1.19.1"
GO_ARCH="linux-amd64"
# from https://golang.org/dl
GO_VERSION_SHA="acc512fbab4f716a8f97a8b3fbaa9ddd39606a28be6c2515ef7c6c6311acffde"

# Download Go and verify Go tarball
setup_go () {
    echo "installing go"
    curl -L https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz --output go${GO_VERSION}.${GO_ARCH}.tar.gz
    if ! echo "${GO_VERSION_SHA}  go${GO_VERSION}.${GO_ARCH}.tar.gz" | shasum --algorithm 256 --check -; then
        echo "Checksum failed" >&2
        exit 1
    fi

    sudo rm -rfv /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.${GO_ARCH}.tar.gz

    echo "$PATH"
    which go
    go version
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
