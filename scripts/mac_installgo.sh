install_go() {
    echo "installing go"
    curl https://golang.org/dl/go1.16.2.darwin-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.16.2.linux-amd64.tar.gz
    brew link --force --overwrite go@16
}

if [-d /usr/local/go]; then
    echo "Go is already installed"
    v=`go version | { read _ _ v _; echo ${v#go}; }`
    echo "$v is installed"
    if $v != "1.6.2"; then
        install_go
    fi
else
    install_go
fi