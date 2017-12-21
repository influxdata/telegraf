#!/bin/bash

set -e

echo "Installing zip"
apt-get update
apt-get install -y zip

# create new directories inside of the container
mkdir -p /output
mkdir -p /usr/local/go/src/github.com/influxdata

echo "Copying /src to /gopath/src/github.com/influxdata/telegraf"
cp -r /src /usr/local/go/src/github.com/influxdata/telegraf

# change to build directory
cd /usr/local/go/src/github.com/influxdata/telegraf

echo "Cleaning Build Directory..."
make clean

echo "Restoring Dependencies..."
make deps

echo "Linting Telegraf..."
make lint

echo "Testing Telegraf..."
#make test

echo "Making Telegraf..."
make
ls

echo "Archiving Telegraf..."
# remove existing builds
if [[ -f /output/Linux-x86_64.zip ]]; then
    echo "Removing existing Linux-x86_64.zip"
    rm /output/Linux-x86_64.zip
fi
mkdir -p Linux-x86_64
cp telegraf ./Linux-x86_64/
zip Linux-x86_64.zip Linux-x86_64/telegraf
cp /usr/local/go/src/github.com/influxdata/telegraf/Linux-x86_64.zip /output/Linux-x86_64.zip
ls /output
echo "Done!"
