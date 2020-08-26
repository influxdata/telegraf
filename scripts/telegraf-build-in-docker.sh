#!/bin/bash

set -e

GO_IMAGE=golang:1.15
printf "Building for GOOS=%s and GOARCH=%s\n" "${GOOS}" "${GOARCH}"
printf "After the build is done you should have 'telegraf' binary built in repository root\n"

if [[ $(which go) ]] ; then
    GOMODCACHE="$(go env GOMODCACHE)"
    if [[ ! -z "${GOMODCACHE}" ]] ; then
        printf "GOMODCACHE set (to %s) - sharing pkg/mod/ with docker container enabled\n" "${GOMODCACHE}"
        GOMODCACHE_MOUNT_OPTION="-v ${GOMODCACHE}:/go/pkg/mod/"
    elif [[ ! -z "${GOPATH}" ]] ; then
        printf "GOPATH set (to %s) - sharing of pkg/mod/ with docker container enabled\n" "${GOPATH}"
        GOMODCACHE_MOUNT_OPTION="-v ${GOPATH}/pkg/mod/:/go/pkg/mod/"
    fi
fi

docker run --rm -it \
    -v $(pwd):/telegraf \
    --env GOOS=${GOOS} \
    --env GOARCH=${GOARCH} \
    ${GOMODCACHE_MOUNT_OPTION} \
    --workdir /telegraf \
    "${GO_IMAGE}" make telegraf
