#!/bin/bash
# Run goversioninfo to generate the resource.syso to embed version info.
set -eux

NAME="Telegraf"
VERSION=$(cd ../../ && make version)
FLAGS=()

# Check that an argument is passed
if [ -n "${1+x}" ]; then
  # If arm64, set both arm and 64 flags
  if [ "$1" = "arm64" ]; then
    FLAGS=(-arm -64)
  # If amd64, set only the 64 flag
  elif [ "$1" = "amd64" ]; then
    FLAGS=(-64)
  fi
fi

goversioninfo "${FLAGS[@]}" \
    -product-name "$NAME" \
    -product-version "$VERSION" \
    -skip-versioninfo \
    -icon=../../assets/windows/tiger.ico \
    -o resource.syso
