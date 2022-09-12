#!/bin/sh

set -eux

ARCH=$(uname -m)
GO_VERSION="1.19.1"
if [ "$ARCH" = 'arm64' ]; then
    GO_ARCH="darwin-arm64"
    GO_VERSION_SHA="e46aecce83a9289be16ce4ba9b8478a5b89b8aa0230171d5c6adbc0c66640548" # from https://golang.org/dl
elif [ "$ARCH" = 'x86_64' ]; then
    GO_ARCH="darwin-amd64"
    GO_VERSION_SHA="b2828a2b05f0d2169afc74c11ed010775bf7cf0061822b275697b2f470495fb7" # from https://golang.org/dl
fi

# This path is cachable. (Saving in /usr/local/ would cause issues restoring the cache.)
path="/usr/local/Cellar"
sudo mkdir -p ${path}

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
    sudo mkdir -p /usr/local/bin
    sudo ln -sf ${path}/go/bin/go /usr/local/bin/go
    sudo ln -sf ${path}/go/bin/gofmt /usr/local/bin/gofmt
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
