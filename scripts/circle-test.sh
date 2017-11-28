#!/bin/bash
#
# This is the Telegraf CircleCI test script. Using this script allows total control
# the environment in which the build and test is run, and matches the official
# build process for InfluxDB.

BUILD_DIR=$HOME/telegraf-build
VERSION=`git describe --always --tags`

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
exit_if_fail mkdir $BUILD_DIR
export GOPATH=$BUILD_DIR
# Turning off GOGC speeds up build times
export GOGC=off
export PATH=$GOPATH/bin:$PATH
exit_if_fail mkdir -p $GOPATH/src/github.com/influxdata

# Dump some test config to the log.
echo "Test configuration"
echo "========================================"
echo "\$HOME: $HOME"
echo "\$GOPATH: $GOPATH"
echo "\$CIRCLE_BRANCH: $CIRCLE_BRANCH"

# Move the checked-out source to a better location
exit_if_fail mv $HOME/telegraf $GOPATH/src/github.com/influxdata
exit_if_fail cd $GOPATH/src/github.com/influxdata/telegraf

# Verify that go fmt has been run
check_go_fmt

# Build the code
exit_if_fail make

# Run the tests
exit_if_fail go vet ./...
exit_if_fail make docker-run-circle
# Sleep for OpenTSDB leadership election, aerospike cluster, etc.
exit_if_fail sleep 60
exit_if_fail go test -race ./...

# Simple Integration Tests
#   check that one test cpu & mem output work
tmpdir=$(mktemp -d)
./telegraf config > $tmpdir/config.toml
exit_if_fail ./telegraf -config $tmpdir/config.toml \
    -test -input-filter cpu:mem

gzip telegraf -c > "$CIRCLE_ARTIFACTS/telegraf.gz"

if git describe --exact-match HEAD 2>&1 >/dev/null; then
    # install fpm (packaging dependency)
    exit_if_fail gem install fpm
    # install boto & rpm (packaging & AWS dependencies)
    exit_if_fail sudo apt-get install -y rpm python-boto
    unset GOGC
    tag=$(git describe --exact-match HEAD)
    echo $tag
    exit_if_fail ./scripts/build.py --release --package --platform=all --arch=all --upload --bucket=dl.influxdata.com/telegraf/releases
    mv build $CIRCLE_ARTIFACTS
elif [ -n "${PACKAGE}" ]; then
    # install fpm (packaging dependency)
    exit_if_fail gem install fpm
    # install boto & rpm (packaging & AWS dependencies)
    exit_if_fail sudo apt-get install -y rpm python-boto
    unset GOGC
    if [ "$(git rev-parse --abbrev-ref HEAD)" = master ]
    then
        exit_if_fail ./scripts/build.py --nightly --package --platform=all --arch=all --upload --bucket=dl.influxdata.com/telegraf/nightlies
    else
        exit_if_fail ./scripts/build.py --package --platform=all --arch=all
    fi
    mv build $CIRCLE_ARTIFACTS
fi
