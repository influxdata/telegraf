#!/bin/bash

VERSION=${CIRCLE_TAG##*v}
BUILD_DIR=$HOME/telegraf-build
CIRCLE_RELEASE_REPO="telegraf-output-orangesys"
CIRCLE_RELEASE_USER="orangesys"

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
#exit_if_fail mkdir $BUILD_DIR
export GOPATH=$BUILD_DIR
# Turning off GOGC speeds up build times
export GOGC=off
export PATH=$GOPATH/bin:$PATH
#exit_if_fail mkdir -p $GOPATH/src/github.com/influxdata

# Dump some test config to the log.
echo "Test configuration"
echo "========================================"
echo "\$HOME: $HOME"
echo "\$GOPATH: $GOPATH"
echo "\$CIRCLE_BRANCH: $CIRCLE_BRANCH"
echo "\$CIRCLE_TAG: $CIRCLE_TAG"

sudo apt-get install -y rpm python-boto ruby ruby-dev autoconf libtool rpm
sudo gem instal fpm

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