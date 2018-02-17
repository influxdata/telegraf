#!/bin/bash
set -ex

VERSION=${CIRCLE_TAG##*v}
CIRCLE_RELEASE_REPO="telegraf-output-orangesys"
CIRCLE_RELEASE_USER="orangesys"
ARTIFACT_DIR='artifacts'

run()
{
    "$@"
    ret=$?
    if [[ $ret -eq 0 ]]
    then
        echo "[INFO]  [ $@ ]"
    else
        echo "[ERROR] [ $@ ] returned $ret"
        exit $ret
    fi
}

# Turning off GOGC speeds up build times
export PATH=$GOPATH/bin:$PATH
run mkdir -p ${ARTIFACT_DIR}

# Dump some test config to the log.
echo "Test configuration"
echo "========================================"
echo "\$HOME: $HOME"
echo "\$GOPATH: $GOPATH"
echo "\$CIRCLE_BRANCH: $CIRCLE_BRANCH"
echo "\$CIRCLE_TAG: $CIRCLE_TAG"

run sudo apt-get install -y rpm python-boto ruby ruby-dev autoconf libtool rpm
run sudo gem instal fpm

run ./scripts/build.py --release --package --platform=linux \
  --arch=amd64 --version=${VERSION}
run rm -f build/telegraf
run mv build $ARTIFACT_DIR

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

cd $ARTIFACT_DIR/build

for i in `ls`; do
  run upload_file $i
done