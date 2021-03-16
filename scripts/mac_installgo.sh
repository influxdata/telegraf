#!/bin/sh

version="1.16.2"
path="/Users/distiller"

# Download Go directly from tar, the reason we aren't using brew: it is slow to update and we can't pull specific minor versions
download_go () {
    echo "installing go"
    curl -OL https://golang.org/dl/go${version}.darwin-amd64.tar.gz --output go${version}.darwin-amd64.tar.gz
    rm -rf ${path}/go
    tar -C $path -xzf go${version}.darwin-amd64.tar.gz
}

check_go () {
    if [ -d $path/go ]l; then
        echo "Go is already downloaded"
        v=`${path}/go/bin/go version | { read _ _ v _; echo ${v#go}; }`
        echo "$v is downloaded, required version is $version"
        if [ "$v" != $version ]; then
            download_go
            $path/go/bin/go version
        fi
    else
        download_go
    fi
}

setup_go () {
    if [ -d $path/go ]; then
        sudo cp ${path}/go/bin/go /usr/local/bin/
        sudo cp ${path}/go/bin/gofmt /usr/local/bin/
        sudo cp -R ${path}/go /usr/local/
    else
        echo "Missing go from macdeps, ${path} doesn't exist"
    fi
}

for arg in "$@"
do
    case $arg in
        --download)
        check_go
        shift
        ;;
        --setup)
        setup_go
        shift
        ;;
    esac
done
