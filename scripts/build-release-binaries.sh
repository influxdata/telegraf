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

go mod download
for OS in windows darwin linux; do
    echo "Building telegraf for ${OS}..."
    BINARY_PATH="telegraf-${GIT_TAG}_${OS}_amd64"
    if [[ "${OS}" == "windows" ]] ; then
        BINARY_PATH="${BINARY_PATH}.exe"
    fi

    GOOS=${OS} GOARCH=amd64 go build -o "${BINARY_PATH}" ./cmd/telegraf

    if [[ "${OS}" == "windows" ]] ; then
        ARCHIVE_PATH="$(basename "${BINARY_PATH}.zip")"
        zip -q "${ARCHIVE_PATH}" "${BINARY_PATH}"
    else
        ARCHIVE_PATH="$(basename "${BINARY_PATH}.tar.gz")"
        tar -czvf "${ARCHIVE_PATH}" "$(basename "${BINARY_PATH}")"
    fi

    if [[ -f "${DIR}/${ARCHIVE_PATH}" ]]; then
        rm "${DIR}/${ARCHIVE_PATH}"
    fi
    cp "${ARCHIVE_PATH}" "${DIR}"

    if [[ -f "${DIR}/${BINARY_PATH}" ]];then
        rm "${DIR}/${BINARY_PATH}"
    fi
    cp "${BINARY_PATH}" "${DIR}"

    echo "Successfully built ${BINARY_PATH} (compressed into ${ARCHIVE_PATH})"
done
