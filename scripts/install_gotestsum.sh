#!/bin/sh

set -eux

OS=$1
EXE=$2
VERSION="1.7.0"

WINDOWS_SHA="7ae12ddb171375f0c14d6a09dd27a5c1d1fc72edeea674e3d6e7489a533b40c1"
DARWIN_SHA="a8e2351604882af1a67601cbeeacdcfa9b17fc2f6fbac291cf5d434efdf2d85b"
LINUX_SHA="b5c98cc408c75e76a097354d9487dca114996e821b3af29a0442aa6c9159bd40"

setup_gotestsum () {
    echo "installing gotestsum"
    curl -L "https://github.com/gotestyourself/gotestsum/releases/download/v${VERSION}/gotestsum_${VERSION}_${OS}_amd64.tar.gz" --output gotestsum.tar.gz

    if [ "$OS" = "windows" ]; then
        SHA=$WINDOWS_SHA
        SHATOOL="sha256sum"
    elif [ "$OS" = "darwin" ]; then
        SHA=$DARWIN_SHA
        SHATOOL="shasum --algorithm 256"
    elif [ "$OS" = "linux" ]; then
        SHA=$LINUX_SHA
        SHATOOL="sha256sum"
    fi

    if ! echo "${SHA}  gotestsum.tar.gz" | ${SHATOOL} --check -; then
        echo "Checksum failed" >&2
        exit 1
    fi

    tar --extract --file=gotestsum.tar.gz "${EXE}"
}

if test -f "${EXE}"; then
    echo "gotestsum is already installed"
    v=$(./"${EXE}" --version)
    echo "$v is installed, required version is ${VERSION}"
    if [ "$v" != "gotestsum version ${VERSION}" ]; then
        setup_gotestsum
        ${EXE} --version
    fi
else
    setup_gotestsum
fi
