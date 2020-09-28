#!/bin/bash

docker run \
    --rm \
    --volume "$(pwd)":/app \
    --workdir /app \
    --env GIT_TAG \
    --entrypoint /app/scripts/build-release-binaries-docker-entrypoint.sh \
    golang:1.15
