#!/bin/sh

set -ux

GO_VERSION="1.21.3"

setup_go () {
    choco upgrade golang --allow-downgrade --version=${GO_VERSION} --verbose --debug
    git config --system core.longpaths true
}

#export PATH="/c/Go/bin:$PATH"
#refreshenv

echo "PATH before: $PATH"
echo "go location before: $(command -v go)"

'/c/Users/circleci/go/bin/go' version
'/c/Program Files/Go/bin/go' version
'/c/Go/bin/go' version



if command -v go >/dev/null 2>&1; then
    echo "Go is already installed"
    v=$(go version | { read -r _ _ v _; echo "${v#go}"; })
    echo "$v is installed, required version is ${GO_VERSION}"
    if [ "$v" != ${GO_VERSION} ]; then
        setup_go
        command -v go
        go version
    fi
else
    setup_go
    command -v go
    go version
fi

echo "PATH after: $PATH"

'/c/Users/circleci/go/bin/go' version
'/c/Program Files/Go/bin/go' version
'/c/Go/bin/go' version