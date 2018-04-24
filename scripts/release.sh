#!/bin/bash

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
run mkdir -p ${ARTIFACT_DIR}
run gzip telegraf -c > "$ARTIFACT_DIR/telegraf.gz"

# RPM is used to build packages for Enterprise Linux hosts.
# Boto is used to upload packages to S3.
run sudo apt-get install -y rpm python-boto ruby ruby-dev autoconf libtool
run sudo gem install fpm

if git describe --exact-match HEAD 2>&1 >/dev/null; then
    run ./scripts/build.py --release --package --platform=all --arch=all --upload --bucket=dl.influxdata.com/telegraf/releases
elif [ "${CIRCLE_STAGE}" = nightly ]; then
	run ./scripts/build.py --nightly --package --platform=all --arch=all --upload --bucket=dl.influxdata.com/telegraf/nightlies
else
	run ./scripts/build.py --package --platform=all --arch=all
fi

run mv build $ARTIFACT_DIR
