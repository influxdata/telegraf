#!/bin/bash
# Run goversioninfo to generate the resource.syso to embed version info.
set -eux

NAME="Telegraf"
VERSION=$(cd ../../ && make version)
FLAGS=()

# If building for arm64, then incude the extra flags required.
if [ -n "${1+x}" ] && [ "$1" = "arm64" ]; then
    FLAGS=(-arm -64)
fi

goversioninfo "${FLAGS[@]}" \
    -product-name "$NAME" \
    -product-version "$VERSION" \
    -skip-versioninfo \
    -icon=../../assets/windows/tiger.ico \
    -o resource.syso
