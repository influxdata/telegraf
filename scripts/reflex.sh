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

if [[ -z "${GOPATH}" ]] ; then
    printf "GOPATH unset - sharing of pkg/mod/cache disable dwith docker container\n"
    docker run --rm -ti \
        -v $(pwd):/telegraf \
        --entrypoint reflex \
        telegraf-reflex:latest -r '(\.go$|go\.mod)' --start-service -- ${CMD}
else
    printf "GOPATH set (to %s) - sharing of pkg/mod/cache with docker container enabled\n" "${GOPATH}"
    docker run --rm -ti \
        -v $(pwd):/telegraf \
        -v $GOPATH/pkg/mod/cache:/go/pkg/mod/cache \
        --entrypoint reflex \
        telegraf-reflex:latest -r '(\.go$|go\.mod)' --start-service -- ${CMD}
fi

printf "'%s' finished successfully!\n" "${CMD}"