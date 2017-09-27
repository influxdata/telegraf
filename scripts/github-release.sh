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

# Check that go fmt has been run.
function check_go_fmt {
    fmtcount=`git ls-files | grep '.go$' | grep -v Godep | xargs gofmt -l 2>&1 | wc -l`
    if [ $fmtcount -gt 0 ]; then
        echo "run 'go fmt ./...' to format your source code."
        exit 1
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

# Move the checked-out source to a better location
#exit_if_fail mv $HOME/telegraf $GOPATH/src/github.com/influxdata
exit_if_fail cd $GOPATH/src/github.com/influxdata/telegraf

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