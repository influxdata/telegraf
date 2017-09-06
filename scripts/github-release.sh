#!/bin/bash

set -ex

VERSION=${CIRCLE_TAG##*v}
BUILD_DIR=$HOME/telegraf-build

# Executes the given statement, and exits if the command returns a non-zero code.
function exit_if_fail {
    command=$@
    echo "Executing '$command'"
    eval $command
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "'$command' returned $rc."
        exit $rc
    fi
}

# Set up the build directory, and then GOPATH.
exit_if_fail mkdir $BUILD_DIR
export GOPATH=$BUILD_DIR
# Turning off GOGC speeds up build times
export GOGC=off
export PATH=$GOPATH/bin:$PATH

gem instal fpm

sudo apt-get install -y rpm
unset GOGC
./scripts/build.py --release --package --platform=linux \
  --arch=amd64 --version=${VERSION}
rm build/telegraf
mv build $CIRCLE_ARTIFACTS

#intall github-release cmd
go get github.com/aktau/github-release

github-release release \
  --user $CIRCLE_RELEASE_USER \
  --repo $CIRCLE_RELEASE_REPO \
  --tag $VERSION \
  --name "orangesys-telegraf-${VERSION}" \
  --description "telegraf output orangesys"

upload_file() {
  _FILE=$1
  github-release upload \
    --user $CIRCLE_RELEASE_USER \
    --repo $CIRCLE_RELEASE_REPO \
    --tag $VERSION \
    --name "$_FILE" \
    --file $_FILE
}

cd ${CIRCLE_ARTIFACTS}/build

for i in `ls`; do
  upload_file $i
done
