#!/bin/sh

version="1.16.2"
# This path is cachable, while saving directly in /usr/local/ will cause issues restoring the cache
path="/usr/local/Cellar"

# Download Go directly from tar, the reason we aren't using brew: it is slow to update and we can't pull specific minor versions
setup_go () {
    echo "installing go"
    curl -OL https://golang.org/dl/go${version}.darwin-amd64.tar.gz --output go${version}.darwin-amd64.tar.gz
    sudo rm -rf ${path}/go
    sudo tar -C $path -xzf go${version}.darwin-amd64.tar.gz
    ln -sf ${path}/go/bin/go /usr/local/bin/go
    ln -sf ${path}/go/bin/gofmt /usr/local/bin/gofmt
}

if command -v go &> /dev/null; then
    echo "Go is already installed"
    v=`go version | { read _ _ v _; echo ${v#go}; }`
    echo "$v is installed, required version is $version"
    if [ "$v" != $version ]; then
        setup_go
        go version
    fi
else
    setup_go
fi
