#!/bin/bash

set -e

apt-get update && apt-get install --yes zip
GIT_TAG="${GIT_TAG}" ./scripts/build-release-binaries.sh
