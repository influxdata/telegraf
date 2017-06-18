#!/bin/bash

BUILD_DIR=$HOME/telegraf-build
VERSION=`git describe --always --tags`

export GOPATH=$BUILD_DIR
export PATH=$GOPATH/bin:$PATH

#intall github-release cmd
go get github.com/aktau/github-release
cd $CIRCLE_ARTIFACTS/build && rm -f telegraf

#
# Create a release page
#
github-release release \
  --user $CIRCLE_PROJECT_USERNAME \
  --repo $CIRCLE_RELEASE_URL \
  --tag $VERSION \
  --name "Orangesys-telegraf-${VERSION}"
  --description "telegraf output orangesys"

#
# Upload package files and build a release note
#
github-release upload \
  --user $CIRCLE_PROJECT_USERNAME \
  --repo $CIRCLE_RELEASE_URL \
  --tag $VERSION \
  --name "telegraf-output-orangesys"
  --file telegraf*
