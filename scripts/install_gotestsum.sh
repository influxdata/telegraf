#!/bin/sh

set -eux

OS=$1
EXE=$2
VERSION="1.10.1"
ARCH=$(uname -m)

WINDOWS_SHA="3a409d05e6d0b89b7860b3a1d66bd855831a276ac25a05d33700f330f554d315"
DARWIN_ARM64_SHA="01be1b28f7c2558af6191050671a97e783eab5ceb813ea8bfac739d5759de596"
LINUX_SHA="44be2c02d4cf99cdd61edcb27851ef98ef8724a2ae3355b438bd108e9abb9056"

if [ "$ARCH" = 'arm64' ]; then
    GO_ARCH="arm64"
elif [ "$ARCH" = 'x86_64' ]; then
    GO_ARCH="amd64"
fi

setup_gotestsum () {
    echo "installing gotestsum"
    curl -L "https://github.com/gotestyourself/gotestsum/releases/download/v${VERSION}/gotestsum_${VERSION}_${OS}_${GO_ARCH}.tar.gz" --output gotestsum.tar.gz

    if [ "$OS" = "windows" ]; then
        SHA=$WINDOWS_SHA
        SHATOOL="sha256sum"
    elif [ "$OS" = "darwin" ]; then
        SHA=$DARWIN_ARM64_SHA
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
        ./"${EXE}" --version
    fi
else
    setup_gotestsum
fi
