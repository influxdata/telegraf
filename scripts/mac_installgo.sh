#!/bin/sh

set -eux

GO_ARCH="darwin-amd64"
GO_VERSION="1.16.2"
GO_VERSION_SHA="c98cde81517c5daf427f3071412f39d5bc58f6120e90a0d94cc51480fa04dbc1" # from https://golang.org/dl

# This path is cachable. (Saving in /usr/local/ would cause issues restoring the cache.)
path="/usr/local/Cellar"

# Download Go and verify Go tarball. (Note: we aren't using brew because
# it is slow to update and we can't pull specific minor versions.)
setup_go () {

    brew install coreutils # Install sha256sum util

    echo "installing go"
    curl -OL https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz --output go${GO_VERSION}.${GO_ARCH}.tar.gz
    echo "${GO_VERSION_SHA} go${GO_VERSION}.${GO_ARCH}.tar.gz" | sha256sum --check
    sudo rm -rf ${path}/go
    sudo tar -C $path -xzf go${GO_VERSION}.${GO_ARCH}.tar.gz
    ln -sf ${path}/go/bin/go /usr/local/bin/go
    ln -sf ${path}/go/bin/gofmt /usr/local/bin/gofmt
}

if command -v go &> /dev/null; then
    echo "Go is already installed"
    v=`go version | { read _ _ v _; echo ${v#go}; }`
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
        go version
    fi
else
    setup_go
fi
