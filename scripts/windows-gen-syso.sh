#!/bin/bash
# Run goversioninfo to generate the resource.syso to embed version info.
set -eux

NAME="Telegraf"
VERSION=$(cd ../../ && make version)
FLAGS=()

# If building for arm64, then include the extra required flag.
if [ -n "${1+x}" ] && [ "$1" = "arm64" ]; then
    FLAGS=(-arm)
fi

goversioninfo "${FLAGS[@]}" \
    -64 \
    -product-name "$NAME" \
    -product-version "$VERSION" \
    -skip-versioninfo \
    -icon=../../assets/windows/tiger.ico \
    -o resource.syso
