#!/bin/sh

set -eux

ARCH=$(uname -m)
GO_VERSION="1.25.2"
GO_VERSION_SHA_arm64="d1ade1b480e51b6269b6e65856c602aed047e1f0d32fffef7eebbd7faa8d7687" # from https://go.dev/dl
GO_VERSION_SHA_amd64="95493abb01da81638ab5083ff3f97e8f923cb42a64c2e16728e3cf5b0cd3fc5a" # from https://go.dev/dl

if [ "$ARCH" = 'arm64' ]; then
    GO_ARCH="darwin-arm64"
    GO_VERSION_SHA=${GO_VERSION_SHA_arm64}
elif [ "$ARCH" = 'x86_64' ]; then
    GO_ARCH="darwin-amd64"
    GO_VERSION_SHA=${GO_VERSION_SHA_amd64}
fi

# This path is cacheable. (Saving in /usr/local/ would cause issues restoring the cache.)
path="/usr/local/Cellar"
sudo mkdir -p ${path}

# Download Go and verify Go tarball. (Note: we aren't using brew because
# it is slow to update and we can't pull specific minor versions.)
setup_go () {
    echo "installing go"
    curl -L "https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz" --output "go${GO_VERSION}.${GO_ARCH}.tar.gz"
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
