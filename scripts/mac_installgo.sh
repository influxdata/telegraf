#!/bin/sh

set -eux

GO_ARCH="darwin-amd64"
GO_VERSION="1.16.6"
GO_VERSION_SHA="0b49b6cbe50b30aa0a5bb9f8ccdbb43f9cd3d9a3c36a769b8e46777d694539b5" # from https://golang.org/dl

# This path is cachable. (Saving in /usr/local/ would cause issues restoring the cache.)
path="/usr/local/Cellar"

# Download Go and verify Go tarball. (Note: we aren't using brew because
# it is slow to update and we can't pull specific minor versions.)
setup_go () {
    echo "installing go"
    curl -L https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz --output go${GO_VERSION}.${GO_ARCH}.tar.gz
    echo "${GO_VERSION_SHA}  go${GO_VERSION}.${GO_ARCH}.tar.gz" | shasum -a 256 --check
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
