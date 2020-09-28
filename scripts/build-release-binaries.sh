#!/bin/bash

set -e

if [[ -z "${GIT_TAG}" ]]; then
    echo "Please provide GIT_TAG environment variable to point which version to build"
    exit 1
fi

DIR="$(pwd)"
TMP_PATH="$(mktemp -d)/telegraf"
REPO_URL="https://github.com/SumoLogic/telegraf.git"

FLAGS="--quiet"
if [[ -n "${CI}" ]] ; then
    FLAGS=""
fi

mkdir "${TMP_PATH}"

function cleanup() {
    rm -rf "${TMP_PATH}"
}
trap cleanup EXIT

echo "Cloning ${REPO_URL} to ${TMP_PATH}..."
git clone ${FLAGS} --depth 1 ${REPO_URL} "${TMP_PATH}" && cd "${TMP_PATH}"
git fetch ${FLAGS} --tags
git checkout ${FLAGS} "${GIT_TAG}"
echo "Checked out telegraf at ${GIT_TAG}"

# go mod download
for OS in windows darwin linux; do
    echo "Building telegraf for ${OS}..."
    if [[ "${OS}" == "windows" ]] ; then
        BINARY_PATH="${DIR}/telegraf-${GIT_TAG}_${OS}_amd64.exe"
    else
        BINARY_PATH="${DIR}/telegraf-${GIT_TAG}_${OS}_amd64"
    fi
    GOOS=${OS} GOARCH=amd64 go build -o "${BINARY_PATH}" ./cmd/telegraf
    echo "Successfully built ${BINARY_PATH}"
done
