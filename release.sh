#!/bin/bash

VERSION="0.9.b1"

echo "Building Tivan version $VERSION"

mkdir -p pkg

build() {
  echo -n "=> $1-$2: "
  GOOS=$1 GOARCH=$2 go build -o pkg/tivan-$1-$2 -ldflags "-X main.Version $VERSION" ./cmd/tivan/tivan.go
  du -h pkg/tivan-$1-$2
}

build "darwin" "amd64"
build "linux" "amd64"
build "linux" "386"

