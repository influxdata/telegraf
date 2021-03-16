#!/bin/sh

version="1.16.2"

# Install Go directly from tar, the reason we aren't using brew: it is slow to update and we can't pull specific minor versions
install_go () {
    echo "installing go"
    curl -OL https://golang.org/dl/go${version}.darwin-amd64.tar.gz --output go${version}.darwin-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /Users/distiller/ -xzf go${version}.darwin-amd64.tar.gz
}

if command -v go &> /dev/null; then
    echo "Go is already installed"
    v=`go version | { read _ _ v _; echo ${v#go}; }`
    echo "$v is installed, required version is $version"
    if [ "$v" != $version ]; then
        install_go
        go version
    fi
else
    install_go
fi
