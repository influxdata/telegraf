#!/bin/bash

set -eo pipefail

echo "Installing zip"
apt-get update
apt-get install -y zip

echo "Installing dep"
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# create new directories inside of the container
mkdir -p /output
mkdir -p /go/src/github.com/influxdata

# This seems to be a lot faster than copying /src.
echo "Copying /src to /gopath/src/github.com/influxdata/telegraf"
git clone /src /go/src/github.com/influxdata/telegraf

echo "Applying diffs to tree..."
(cd /src && git diff HEAD | (cd /go/src/github.com/influxdata/telegraf && git apply))

# change to build directory
cd /go/src/github.com/influxdata/telegraf

echo "Cleaning Build Directory..."
make clean

echo "Linting Telegraf..."
make lint

#echo "Testing Telegraf..."
#make test

echo "Making Telegraf..."
make

echo "Archiving Telegraf..."
# remove existing builds
if [[ -f /output/Linux-x86_64.zip ]]; then
    echo "Removing existing Linux-x86_64.zip"
    rm /output/Linux-x86_64.zip
fi
mkdir -p Linux-x86_64
cp telegraf ./Linux-x86_64/
zip Linux-x86_64.zip Linux-x86_64/telegraf
cp /go/src/github.com/influxdata/telegraf/Linux-x86_64.zip /output/Linux-x86_64.zip
ls /output
echo "Done!"
