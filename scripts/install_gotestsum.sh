#!/bin/sh

set -eux

OS=$1
EXE=$2
VERSION="1.13.0"
ARCH=$(uname -m)

WINDOWS_SHA="fd5a6dc69e46a0970593e70d85a7e75f16714e9c61d6d72ccc324eb82df5bb8a"
DARWIN_ARM64_SHA="509cb27aef747f48faf9bce424f59dcf79572c905204b990ee935bbfcc7fa0e9"
LINUX_SHA="11ccddeaf708ef228889f9fe2f68291a75b27013ddfc3b18156e094f5f40e8ee"

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
