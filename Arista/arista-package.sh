#!/usr/bin/env bash

. ./tools2-functions

#redefine cleanup_exit
cleanup_exit() {
    OK=${1:-1}
    rm -rf $TMP_BIN_DIR > /dev/null 2>&1
    rm -rf $TMP_CONFIG_DIR > /dev/null 2>&1
    rm -rf ./*.rpm > /dev/null 2>&1
    exit $OK
}

check_linux
check_gopath
check_fpm

ARCH=x86_64
GOBIN=../
TMP_BIN_DIR=./rpm_bin
TMP_CONFIG_DIR=./rpm_config
CONFIG_FILES_DIR=./ConfigFiles

LICENSE=MIT
URL=github.com/aristanetworks/telegraf
TELEGRAF_UPSTREAM_VERSION="v1.2.0-823-g89c596cf"
DESCRIPTION="InfluxDB Telegraf agent version: $TELEGRAF_UPSTREAM_VERSION"
VENDOR=Influxdata

set -e

# It's common practice to use a 'v' prefix on tags, but the prefix should be
# removed when making the RPM version string.
#
# Use "git describe" as the basic RPM version data.  If there are no tags
# yet, simulate a v0 tag on the initial/empty repo and a "git describe"-like
# tag (eg v0-12-gdeadbee) so there's a legitimate, upgradeable RPM version.
#
# Include "-dirty" on the end if there are any uncommitted changes.
#
# Replace hyphens with underscores; RPM uses them to separate version/release.
git_ver=$(git describe --dirty --match "v[0-9]*-ar" 2>/dev/null || echo "v0-`git rev-list --count HEAD`-g`git describe --dirty --always`")
version=$(echo "$git_ver" | sed -e "s/^v//" -e "s/-/_/g")
echo "Version, $version"

# Build and install the latest code
echo "Building and Installing telegraf"
make -C ../
#make -C ../ test-short

echo "Creating RPMS"

# Cleanup old RPMS
mkdir ./RPMS > /dev/null 2>&1 || rm -rf ./RPMS/*
rm ./*.rpm > /dev/null 2>&1  || true

COMMON_FPM_ARGS="\
--log error \
--vendor $VENDOR \
--url $URL \
--license $LICENSE"

# Create Binary RPMS
BINARY_FPM_ARGS="\
 -C $TMP_BIN_DIR \
--prefix / \
-a $ARCH \
-v $version \
$COMMON_FPM_ARGS"

# Make a copy of the generated binaries into a tmp directory bin
echo "Seting up temporary bin directory"
mkdir $TMP_BIN_DIR > /dev/null 2>&1 || rm -rf $TMP_BIN_DIR/*
mkdir -p $TMP_BIN_DIR/usr/bin/
for binary in "telegraf"
do
    cp $GOBIN/$binary $TMP_BIN_DIR/usr/bin/
done

# Add a default telegraf configig
mkdir -p $TMP_BIN_DIR/etc/telegraf/
cp $CONFIG_FILES_DIR/telegraf-default.conf $TMP_BIN_DIR/etc/telegraf/telegraf.conf

fpm -s dir -t rpm $BINARY_FPM_ARGS --description "$DESCRIPTION" -n "telegraf" . || cleanup_exit 1

mv ./*.rpm RPMS

# Create Config RPMS
CONFIG_FPM_ARGS="\
-C $TMP_CONFIG_DIR \
--prefix / \
-a noarch \
-d telegraf \
--config-files /etc/telegraf/ \
--after-install ./post_install_config.sh \
--after-remove ./post_uninstall_config.sh \
-v $version \
$COMMON_FPM_ARGS"

# Create directory structure for config files
echo "Setting up temporary config file tree"
mkdir $TMP_CONFIG_DIR > /dev/null 2>&1 || rm -rf $TMP_CONFIG_DIR/*
mkdir -p $TMP_CONFIG_DIR/etc/default
cp $CONFIG_FILES_DIR/telegraf.default $TMP_CONFIG_DIR/etc/default/telegraf
mkdir -p $TMP_CONFIG_DIR/etc/logrotate.d
cp $CONFIG_FILES_DIR/telegraf.logrotate $TMP_CONFIG_DIR/etc/logrotate.d/telegraf
mkdir -p $TMP_CONFIG_DIR/lib/systemd/system
cp $CONFIG_FILES_DIR/telegraf-dhclient.service $TMP_CONFIG_DIR/lib/systemd/system/
cp $CONFIG_FILES_DIR/telegraf-networkd.service $TMP_CONFIG_DIR/lib/systemd/system/
# To ensure telegraf.service is removed when the rpm itself is removed/uninstalled.
cp $CONFIG_FILES_DIR/telegraf-networkd.service $TMP_CONFIG_DIR/lib/systemd/system/telegraf.service
mkdir -p $TMP_CONFIG_DIR/etc/telegraf
mkdir -p $TMP_CONFIG_DIR/etc/telegraf/telegraf.d

# Linux-Config
rm -f $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-linux.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-Linux" etc lib || cleanup_exit 1

# Redis-Config
rm -f $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-redis.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-Redis" etc lib || cleanup_exit 1

# Docker-Config
rm -f $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-docker.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-Docker" etc lib || cleanup_exit 1

# Perforce-Config
rm -rf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-perforce.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-Perforce" etc lib || cleanup_exit 1

# Apache-Config
rm -f $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-apache.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-Apache" etc lib || cleanup_exit 1

# Swift-Config
rm -rf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-swift.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-Swift" etc lib || cleanup_exit 1

# QUBIT Scylla config
rm -rf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-qubit-scylla.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-qubit-scylla" etc lib || cleanup_exit 1

# QUBIT Scylla dev config
rm -rf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-qubit-scylla-dev.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-qubit-scylla-dev" etc lib || cleanup_exit 1


# QUBIT Worker config
rm -rf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-qubit-worker.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-qubit-worker" etc lib || cleanup_exit 1

# QUBIT Spin config
rm -rf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/*
cp $CONFIG_FILES_DIR/telegraf-qubit-spin.conf $TMP_CONFIG_DIR/etc/telegraf/telegraf.d/
fpm -s dir -t rpm $CONFIG_FPM_ARGS --description "$DESCRIPTION" -n "telegraf-qubit-spin" etc lib || cleanup_exit 1


mv ./*.rpm RPMS

echo "Created RPMS"
ls -l RPMS | awk '{print($9);}'
cleanup_exit 0
