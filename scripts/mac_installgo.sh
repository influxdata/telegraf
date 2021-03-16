#!/bin/sh

version="1.16.2"

# Download Go directly from tar, the reason we aren't using brew: it is slow to update and we can't pull specific minor versions
setup_go () {
    echo "installing go"
    curl -OL https://golang.org/dl/go1.16.2.darwin-amd64.tar.gz --output go1.16.2.darwin-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local/ -xzf go1.16.2.darwin-amd64.tar.gz
    sudo chown -R $USER:admin /usr/local/go
    sudo chmod 775 /usr/local/go
    ln -sf /usr/local/go/bin/go /usr/local/bin/go
    ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
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
