#!/bin/bash
#
# This is the Telegraf CircleCI test script. Using this script allows total control
# the environment in which the build and test is run, and matches the official
# build process for InfluxDB.

BUILD_DIR=$HOME/telegraf-build
# GO_VERSION=go1.4.2

# Executes the given statement, and exits if the command returns a non-zero code.
function exit_if_fail {
    command=$@
    echo "Executing '$command'"
    $command
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "'$command' returned $rc."
        exit $rc
    fi
}

# source $HOME/.gvm/scripts/gvm
# exit_if_fail gvm use $GO_VERSION

# Set up the build directory, and then GOPATH.
exit_if_fail mkdir $BUILD_DIR
export GOPATH=$BUILD_DIR
export PATH=$GOPATH/bin:$PATH
exit_if_fail mkdir -p $GOPATH/src/github.com/influxdb

# Get golint
go get github.com/golang/lint/golint
# Get gox (cross-compiler)
go get github.com/mitchellh/gox
# Get godep tool
go get github.com/tools/godep

# Dump some test config to the log.
echo "Test configuration"
echo "========================================"
echo "\$HOME: $HOME"
echo "\$GOPATH: $GOPATH"
echo "\$CIRCLE_BRANCH: $CIRCLE_BRANCH"

# Move the checked-out source to a better location
exit_if_fail mv $HOME/telegraf $GOPATH/src/github.com/influxdb
exit_if_fail cd $GOPATH/src/github.com/influxdb/telegraf

# Install the code
exit_if_fail godep go build -v ./...
exit_if_fail godep go install -v ./...

# Run the tests
exit_if_fail godep go vet ./...
exit_if_fail godep go test -v -short ./...

# Build binaries
GOPATH=`godep path`:$GOPATH gox -os="linux" -arch="386 amd64" ./...
# Check return code of gox command
exit_if_fail return $?
# Artifact binaries
mv telegraf* $CIRCLE_ARTIFACTS

exit $rc
