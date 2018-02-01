#!/bin/bash
set -u

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

run make
run gzip telegraf -c > "$ARTIFACT_DIR/telegraf.gz"
# RPM is used to build packages for Enterprise Linux hosts.
# Boto is used to upload packages to S3.
#
run sudo apt-get install -y rpm python-boto ruby ruby-dev
run sudo gem install fpm

# If a release tag is found, perform a full release, else run nightlies.
if git describe --exact-match HEAD 2>&1 >/dev/null; then
    run ./scripts/build.py --release --package --platform=all --arch=all --upload --bucket=dl.influxdata.com/telegraf/releases
elif [ -n "${PACKAGE}" ]; then
    if [ "$(git rev-parse --abbrev-ref HEAD)" = master ]
    then
        run ./scripts/build.py --nightly --package --platform=all --arch=all --upload --bucket=dl.influxdata.com/telegraf/nightlies
    else
        run ./scripts/build.py --package --platform=all --arch=all
    fi
fi

run mkdir -p ${ARTIFACT_DIR}
run mv build $ARTIFACT_DIR
