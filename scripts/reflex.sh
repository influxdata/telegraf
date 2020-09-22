#!/bin/bash

set -e

CMD=""

case $1 in
    build)
        CMD="./scripts/reflex-build.sh"
        ;;

    test)
        CMD="./scripts/reflex-test.sh"
        ;;

    *)
        echo "Usage: $0 {test|build}"
        exit 2
        ;;
esac

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

docker run --rm -ti \
    -v $(pwd):/telegraf \
    ${GOMODCACHE_MOUNT_OPTION} \
    --workdir /telegraf \
    --entrypoint reflex \
    telegraf-reflex:latest --decoration=fancy -r '(\.go$|go\.mod)' --start-service -- ${CMD}