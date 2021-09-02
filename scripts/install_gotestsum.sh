#!/bin/sh

set -eux

OS=$1
EXE=$2
VERSION="1.7.0"

setup_gotestsum () {
    echo "installing gotestsum"
    curl -L https://github.com/gotestyourself/gotestsum/releases/download/v${VERSION}/gotestsum_${VERSION}_${OS}_amd64.tar.gz --output gotestsum.tar.gz
    tar --extract --file=gotestsum.tar.gz ${EXE}
}

if test -f ${EXE}; then
    echo "gotestsum is already installed"
    v=`./${EXE} --version`
    echo "$v is installed, required version is ${VERSION}"
    if [ "$v" != "gotestsum version ${VERSION}" ]; then
        setup_gotestsum
        ${EXE} --version
    fi
else
    setup_gotestsum
fi
