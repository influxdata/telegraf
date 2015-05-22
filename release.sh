#!/bin/bash

VERSION="0.9.b1"

echo "Building Telegraf version $VERSION"

mkdir -p pkg

build() {
  echo -n "=> $1-$2: "
  GOOS=$1 GOARCH=$2 go build -o pkg/telegraf-$1-$2 -ldflags "-X main.Version $VERSION" ./cmd/telegraf/telegraf.go
  du -h pkg/telegraf-$1-$2
}

build "darwin" "amd64"
build "linux" "amd64"
build "linux" "386"

