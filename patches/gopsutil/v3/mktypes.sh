#!/bin/sh

PKGS="cpu disk docker host load mem net process"

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
GOARCH=$(go env GOARCH)

for DIR in . v3
do
        (cd "$DIR" || exit
        for PKG in $PKGS
        do
                if [ -e "${PKG}/types_${GOOS}.go" ]; then
                        (echo "// +build $GOOS"
                        echo "// +build $GOARCH"
                        go tool cgo -godefs "${PKG}/types_${GOOS}.go") | gofmt > "${PKG}/${PKG}_${GOOS}_${GOARCH}.go"
                fi
        done)
done
