#!/bin/bash -x
cd /go/src/app
go get -u github.com/golang/lint/golint
go get github.com/sparrc/gdm
gdm restore || (echo 'Error getting dependencies, retrying...' && gdm restore)
ln -s /go/src/app /go/src/github.com/influxdata/telegraf
cd /go/src/github.com/influxdata/telegraf
export COMMIT="$(git log --pretty=format:'%h' -n 1)" 
export BRANCH="$(git rev-parse --abbrev-ref HEAD)" 
GOOS=linux GOARCH=amd64 go build -o ${PWD}/build/telegraf -ldflags="-w -s -X main.branch=${BRANCH} -X main.commit=${COMMIT}" -tags static ./cmd/telegraf