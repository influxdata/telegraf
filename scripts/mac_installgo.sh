#!/bin/sh

# Install Go directly from tar, the reason we aren't using brew: it is slow to update and we can't pull specific minor versions
install_go () {
    echo "installing go"
    curl -OL https://golang.org/dl/go1.16.2.darwin-amd64.tar.gz --output go1.16.2.darwin-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.16.2.darwin-amd64.tar.gz
    ln -s /usr/local/go/bin/go /usr/local/bin
}

if command -v go &> /dev/null; then
    echo "Go is already installed"
    v=`go version | { read _ _ v _; echo ${v#go}; }`
    echo "$v is installed"
    if $v != "1.6.2"; then
        install_go
    fi
else
    install_go
fi
