#!/bin/bash -e
# Run from directory above via ./scripts/cov.sh

rm -rf ./cov
mkdir cov
go test -v -covermode=atomic -coverprofile=./cov/conf.out ./conf
go test -v -covermode=atomic -coverprofile=./cov/log.out ./logger
go test -v -covermode=atomic -coverprofile=./cov/server.out ./server
go test -v -covermode=atomic -coverprofile=./cov/test.out -coverpkg=./server ./test
gocovmerge ./cov/*.out > acc.out
rm -rf ./cov

# If we have an arg, assume travis run and push to coveralls. Otherwise launch browser results
if [[ -n $1 ]]; then
    $HOME/gopath/bin/goveralls -coverprofile=acc.out -service travis-ci
    rm -rf ./acc.out
else
    go tool cover -html=acc.out
fi
