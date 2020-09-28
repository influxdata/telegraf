#!/bin/bash

docker run \
    --rm \
    --volume "$(pwd)":/app \
    --workdir /app \
    --entrypoint /app/scripts/build-release-binaries.sh \
    golang:1.15
