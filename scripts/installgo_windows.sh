#!/bin/sh

set -eux

GO_VERSION="1.21.2"

setup_go () {
    choco upgrade golang --allow-downgrade --version=${GO_VERSION} --verbose --installargs INSTALLDIR="C:\Go"
    git config --system core.longpaths true
}

if command -v C:\Go\go >/dev/null 2>&1; then
    echo "Go is already installed"
    v=$(C:\Go\go version | { read -r _ _ v _; echo "${v#go}"; })
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
        C:\Go\go version
    fi
else
    setup_go
fi
