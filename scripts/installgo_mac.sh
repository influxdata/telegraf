#!/bin/sh

set -eux

GO_ARCH="darwin-amd64"
GO_VERSION="1.17.2"
GO_VERSION_SHA="3c452046b1dfa27b70d3217c9fe6de266f9fd74d83aad81382fead70efcdffca" # from https://golang.org/dl

# This path is cachable. (Saving in /usr/local/ would cause issues restoring the cache.)
path="/usr/local/Cellar"

# Download Go and verify Go tarball. (Note: we aren't using brew because
# it is slow to update and we can't pull specific minor versions.)
setup_go () {
    echo "installing go"
    curl -L https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz --output go${GO_VERSION}.${GO_ARCH}.tar.gz
    if ! echo "${GO_VERSION_SHA}  go${GO_VERSION}.${GO_ARCH}.tar.gz" | shasum --algorithm 256 --check -; then
        echo "Checksum failed" >&2
        exit 1
    fi

    sudo rm -rf ${path}/go
    sudo tar -C $path -xzf go${GO_VERSION}.${GO_ARCH}.tar.gz
    ln -sf ${path}/go/bin/go /usr/local/bin/go
    ln -sf ${path}/go/bin/gofmt /usr/local/bin/gofmt
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
