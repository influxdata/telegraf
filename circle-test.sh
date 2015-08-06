#!/bin/bash
#
# This is the InfluxDB CircleCI test script. Using this script allows total control
# the environment in which the build and test is run, and matches the official
# build process for InfluxDB.

BUILD_DIR=$HOME/telegraf-build
GO_VERSION=go1.4.2

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

source $HOME/.gvm/scripts/gvm
exit_if_fail gvm use $GO_VERSION

# Set up the build directory, and then GOPATH.
exit_if_fail mkdir $BUILD_DIR
export GOPATH=$BUILD_DIR
exit_if_fail mkdir -p $GOPATH/src/github.com/influxdb

# Dump some test config to the log.
echo "Test configuration"
echo "========================================"
echo "\$HOME: $HOME"
echo "\$GOPATH: $GOPATH"
echo "\$CIRCLE_BRANCH: $CIRCLE_BRANCH"

# Move the checked-out source to a better location.
exit_if_fail mv $HOME/telegraf $GOPATH/src/github.com/influxdb
exit_if_fail cd $GOPATH/src/github.com/influxdb/telegraf
exit_if_fail git branch --set-upstream-to=origin/$CIRCLE_BRANCH $CIRCLE_BRANCH

# Install the code.
exit_if_fail cd $GOPATH/src/github.com/influxdb/telegraf
exit_if_fail go get -t -d -v ./...
exit_if_fail git checkout $CIRCLE_BRANCH # 'go get' switches to master. Who knew? Switch back.
exit_if_fail go build -v ./...

# Run the tests.
exit_if_fail [ `git ls-files | grep '.go$' | xargs gofmt -l 2>&1 | wc -l` -eq 0 ]
exit_if_fail go tool vet --composites=false .
exit_if_fail make test-short

exit $rc
