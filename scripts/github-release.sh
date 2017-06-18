#!/bin/bash

VERSION="1.3.2"
BUILD_DIR=$HOME/telegraf-build
export GOPATH=$BUILD_DIR
export PATH=$GOPATH/bin:$PATH

# Dump some test config to the log.
echo "Test configuration"
echo "========================================"
echo "\$HOME: $HOME"
echo "\$GOPATH: $GOPATH"
echo "\$CIRCLE_BRANCH: $CIRCLE_BRANCH"

gem instal fpm

sudo apt-get install -y rpm
unset GOGC
./scripts/build.py --release --package --platform=linux \
  --arch=amd64 --version=${VERSION}
mv build $CIRCLE_ARTIFACTS

#intall github-release cmd
go get github.com/aktau/github-release

upload_file() {
  _FILE=$1
  github-release upload \
    --user $CIRCLE_RELEASE_USER \
    --repo $CIRCLE_RELEASE_REPO \
    --tag $VERSION \
    --name "$_FILE" \
    --file $_FILE
}
cd ${CIRCLE_ARTIFACTS}/build && rm -fr telegraf

for i in `ls`; do
  upload_file $i
done
