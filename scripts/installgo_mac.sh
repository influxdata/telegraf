#!/bin/sh

set -eux

ARCH=$(uname -m)
GO_VERSION="1.22.0"
GO_VERSION_SHA_arm64="bf8e388b09134164717cd52d3285a4ab3b68691b80515212da0e9f56f518fb1e" # from https://golang.org/dl
GO_VERSION_SHA_amd64="ebca81df938d2d1047cc992be6c6c759543cf309d401b86af38a6aed3d4090f4" # from https://golang.org/dl

if [ "$ARCH" = 'arm64' ]; then
    GO_ARCH="darwin-arm64"
    GO_VERSION_SHA=${GO_VERSION_SHA_arm64}
elif [ "$ARCH" = 'x86_64' ]; then
    GO_ARCH="darwin-amd64"
    GO_VERSION_SHA=${GO_VERSION_SHA_amd64}
fi

# This path is cachable. (Saving in /usr/local/ would cause issues restoring the cache.)
path="/usr/local/Cellar"
sudo mkdir -p ${path}

# Download Go and verify Go tarball. (Note: we aren't using brew because
# it is slow to update and we can't pull specific minor versions.)
setup_go () {
    echo "installing go"
    curl -L "https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz" --output "go${GO_VERSION}.${GO_ARCH}.tar.gz"
    if ! echo "${GO_VERSION_SHA}  go${GO_VERSION}.${GO_ARCH}.tar.gz" | shasum --algorithm 256 --check -; then
        echo "Checksum failed" >&2
        exit 1
    fi

    sudo rm -rf "${path}/go"
    sudo tar -C "$path" -xzf "go${GO_VERSION}.${GO_ARCH}.tar.gz"
    sudo mkdir -p /usr/local/bin
    sudo ln -sf "${path}/go/bin/go" /usr/local/bin/go
    sudo ln -sf "${path}/go/bin/gofmt" /usr/local/bin/gofmt
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
