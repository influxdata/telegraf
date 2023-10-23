#!/bin/sh

set -eux

MINGW_VERSION="12.2.0.03042023"
GCC_VERSION="12.2.0"

setup_mingw () {
    choco upgrade mingw --allow-downgrade --force --version=${MINGW_VERSION}
}

export PATH="/c/ProgramData/chocolatey/lib/mingw/tools/install/mingw64/bin:$PATH"
echo "$PATH"

if command -v gcc >/dev/null 2>&1; then
    echo "MinGW is already installed"
    v=$(gcc -dumpversion)
    echo "$v is installed, required version is ${GCC_VERSION}"
    if [ "$v" != ${GCC_VERSION} ]; then
        setup_mingw
    fi
else
    setup_mingw
fi

echo "$PATH"
command -v gcc
gcc -dumpversion
