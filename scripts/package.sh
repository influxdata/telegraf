#!/usr/bin/env bash

###########################################################################
# Packaging script which creates debian and RPM packages. It optionally
# tags the repo with the given version.
#
# Requirements: GOPATH must be set. 'fpm' must be on the path, and the AWS
# CLI tools must also be installed.
#
#    https://github.com/jordansissel/fpm
#    http://aws.amazon.com/cli/
#
# Packaging process: to package a build, simply execute:
#
#    package.sh
#
# The script will automatically determined the version number from git using
# `git describe --always --tags`
#
# AWS upload: the script will also offer to upload the packages to S3. If
# this option is selected, the credentials should be present in the file
# ~/aws.conf. The contents should be of the form:
#
#    [default]
#    aws_access_key_id=<access ID>
#    aws_secret_access_key=<secret key>
#    region = us-east-1
#
# Trim the leading spaces when creating the file. The script will exit if
# S3 upload is requested, but this file does not exist.

AWS_FILE=~/aws.conf

INSTALL_ROOT_DIR=/opt/telegraf
TELEGRAF_LOG_DIR=/var/log/telegraf
CONFIG_ROOT_DIR=/etc/opt/telegraf
LOGROTATE_DIR=/etc/logrotate.d

SAMPLE_CONFIGURATION=etc/config.sample.toml
LOGROTATE_CONFIGURATION=etc/logrotate.d/telegraf
INITD_SCRIPT=scripts/init.sh
SYSTEMD_SCRIPT=scripts/telegraf.service

TMP_WORK_DIR=`mktemp -d`
POST_INSTALL_PATH=`mktemp`
ARCH=`uname -i`
LICENSE=MIT
URL=influxdb.com
MAINTAINER=support@influxdb.com
VENDOR=InfluxDB
DESCRIPTION="InfluxDB Telegraf agent"
PKG_DEPS=(coreutils)
GO_VERSION="go1.5"
GOPATH_INSTALL=
BINS=(
    telegraf
    )

###########################################################################
# Helper functions.

# usage prints simple usage information.
usage() {
    echo -e "$0\n"
    cleanup_exit $1
}

# make_dir_tree creates the directory structure within the packages.
make_dir_tree() {
    work_dir=$1
    version=$2
    mkdir -p $work_dir/$INSTALL_ROOT_DIR/versions/$version/scripts
    if [ $? -ne 0 ]; then
        echo "Failed to create installation directory -- aborting."
        cleanup_exit 1
    fi
    mkdir -p $work_dir/$CONFIG_ROOT_DIR
    if [ $? -ne 0 ]; then
        echo "Failed to create configuration directory -- aborting."
        cleanup_exit 1
    fi
    mkdir -p $work_dir/$LOGROTATE_DIR
    if [ $? -ne 0 ]; then
        echo "Failed to create configuration directory -- aborting."
        cleanup_exit 1
    fi

}

# cleanup_exit removes all resources created during the process and exits with
# the supplied returned code.
cleanup_exit() {
    rm -r $TMP_WORK_DIR
    rm $POST_INSTALL_PATH
    exit $1
}

# check_gopath sanity checks the value of the GOPATH env variable, and determines
# the path where build artifacts are installed. GOPATH may be a colon-delimited
# list of directories.
check_gopath() {
    [ -z "$GOPATH" ] && echo "GOPATH is not set." && cleanup_exit 1
    GOPATH_INSTALL=`echo $GOPATH | cut -d ':' -f 1`
    [ ! -d "$GOPATH_INSTALL" ] && echo "GOPATH_INSTALL is not a directory." && cleanup_exit 1
    echo "GOPATH ($GOPATH) looks sane, using $GOPATH_INSTALL for installation."
}

check_gvm() {
    source $HOME/.gvm/scripts/gvm
    which gvm
    if [ $? -ne 0 ]; then
        echo "gvm not found -- aborting."
        cleanup_exit $1
    fi
    gvm use $GO_VERSION
    if [ $? -ne 0 ]; then
        echo "gvm cannot find Go version $GO_VERSION -- aborting."
        cleanup_exit $1
    fi
}

# check_clean_tree ensures that no source file is locally modified.
check_clean_tree() {
    modified=$(git ls-files --modified | wc -l)
    if [ $modified -ne 0 ]; then
        echo "The source tree is not clean -- aborting."
        cleanup_exit 1
    fi
    echo "Git tree is clean."
}

# do_build builds the code. The version and commit must be passed in.
do_build() {
    version=$1
    commit=`git rev-parse HEAD`
    if [ $? -ne 0 ]; then
        echo "Unable to retrieve current commit -- aborting"
        cleanup_exit 1
    fi

    for b in ${BINS[*]}; do
        rm -f $GOPATH_INSTALL/bin/$b
    done

    godep go install -a -ldflags="-X main.Version $version" ./...
    if [ $? -ne 0 ]; then
        echo "Build failed, unable to create package -- aborting"
        cleanup_exit 1
    fi
    echo "Build completed successfully."
}

# generate_postinstall_script creates the post-install script for the
# package. It must be passed the version.
generate_postinstall_script() {
    version=$1
    cat  <<EOF >$POST_INSTALL_PATH
rm -f $INSTALL_ROOT_DIR/telegraf
rm -f $INSTALL_ROOT_DIR/init.sh
ln -sfn $INSTALL_ROOT_DIR/versions/$version/telegraf $INSTALL_ROOT_DIR/telegraf

if ! id telegraf >/dev/null 2>&1; then
        useradd --system -U -M telegraf
fi

# Systemd
if which systemctl > /dev/null 2>&1 ; then
    cp $INSTALL_ROOT_DIR/versions/$version/scripts/telegraf.service \
        /lib/systemd/system/telegraf.service
    systemctl enable telegraf

# Sysv
else
    ln -sfn $INSTALL_ROOT_DIR/versions/$version/scripts/init.sh \
        $INSTALL_ROOT_DIR/init.sh
    rm -f /etc/init.d/telegraf
    ln -sfn $INSTALL_ROOT_DIR/init.sh /etc/init.d/telegraf
    chmod +x /etc/init.d/telegraf
    # update-rc.d sysv service:
    if which update-rc.d > /dev/null 2>&1 ; then
        update-rc.d -f telegraf remove
        update-rc.d telegraf defaults
    # CentOS-style sysv:
    else
        chkconfig --add telegraf
    fi

    mkdir -p $TELEGRAF_LOG_DIR
    chown -R -L telegraf:telegraf $TELEGRAF_LOG_DIR
fi

chown -R -L telegraf:telegraf $INSTALL_ROOT_DIR
chmod -R a+rX $INSTALL_ROOT_DIR

EOF
    echo "Post-install script created successfully at $POST_INSTALL_PATH"
}

###########################################################################
# Start the packaging process.

if [ "$1" == "-h" ]; then
    usage 0
fi

VERSION=`git describe --always --tags | tr -d v`
cd `git rev-parse --show-toplevel`
echo -e "\nStarting package process, version: $VERSION\n"

if [ "$CIRCLE_BRANCH" == "" ]; then
    check_gvm
fi
check_gopath
do_build $VERSION
make_dir_tree $TMP_WORK_DIR $VERSION

###########################################################################
# Copy the assets to the installation directories.

for b in ${BINS[*]}; do
    cp $GOPATH_INSTALL/bin/$b $TMP_WORK_DIR/$INSTALL_ROOT_DIR/versions/$VERSION
    if [ $? -ne 0 ]; then
        echo "Failed to copy binaries to packaging directory -- aborting."
        cleanup_exit 1
    fi
done

echo "${BINS[*]} copied to $TMP_WORK_DIR/$INSTALL_ROOT_DIR/versions/$VERSION"

cp $INITD_SCRIPT $TMP_WORK_DIR/$INSTALL_ROOT_DIR/versions/$VERSION/scripts
if [ $? -ne 0 ]; then
    echo "Failed to copy init.d script to packaging directory -- aborting."
    cleanup_exit 1
fi
echo "$INITD_SCRIPT copied to $TMP_WORK_DIR/$INSTALL_ROOT_DIR/versions/$VERSION/scripts"

cp $SYSTEMD_SCRIPT $TMP_WORK_DIR/$INSTALL_ROOT_DIR/versions/$VERSION/scripts
if [ $? -ne 0 ]; then
    echo "Failed to copy systemd file to packaging directory -- aborting."
    cleanup_exit 1
fi
echo "$SYSTEMD_SCRIPT copied to $TMP_WORK_DIR/$INSTALL_ROOT_DIR/versions/$VERSION/scripts"

cp $SAMPLE_CONFIGURATION $TMP_WORK_DIR/$CONFIG_ROOT_DIR/telegraf.conf
if [ $? -ne 0 ]; then
    echo "Failed to copy $SAMPLE_CONFIGURATION to packaging directory -- aborting."
    cleanup_exit 1
fi

cp $LOGROTATE_CONFIGURATION $TMP_WORK_DIR/$LOGROTATE_DIR/telegraf
if [ $? -ne 0 ]; then
    echo "Failed to copy $LOGROTATE_CONFIGURATION to packaging directory -- aborting."
    cleanup_exit 1
fi

generate_postinstall_script $VERSION

###########################################################################
# Create the actual packages.

if [ "$CIRCLE_BRANCH" == "" ]; then
    echo -n "Commence creation of $ARCH packages, version $VERSION? [Y/n] "
    read response
    response=`echo $response | tr 'A-Z' 'a-z'`
    if [ "x$response" == "xn" ]; then
        echo "Packaging aborted."
        cleanup_exit 1
    fi
fi

if [ $ARCH == "i386" ]; then
    rpm_package=telegraf-$VERSION-1.i686.rpm
    debian_package=telegraf_${VERSION}_i686.deb
    deb_args="-a i686"
    rpm_args="setarch i686"
elif [ $ARCH == "arm" ]; then
    rpm_package=telegraf-$VERSION-1.armel.rpm
    debian_package=telegraf_${VERSION}_armel.deb
else
    rpm_package=telegraf-$VERSION-1.x86_64.rpm
    debian_package=telegraf_${VERSION}_amd64.deb
fi

COMMON_FPM_ARGS="-C $TMP_WORK_DIR --vendor $VENDOR --url $URL --license $LICENSE \
                --maintainer $MAINTAINER --after-install $POST_INSTALL_PATH \
                --name telegraf --version $VERSION --config-files $CONFIG_ROOT_DIR ."
$rpm_args fpm -s dir -t rpm --description "$DESCRIPTION" $COMMON_FPM_ARGS
if [ $? -ne 0 ]; then
    echo "Failed to create RPM package -- aborting."
    cleanup_exit 1
fi
echo "RPM package created successfully."

fpm -s dir -t deb $deb_args --description "$DESCRIPTION" $COMMON_FPM_ARGS
if [ $? -ne 0 ]; then
    echo "Failed to create Debian package -- aborting."
    cleanup_exit 1
fi
echo "Debian package created successfully."

###########################################################################
# Offer to publish the packages.

if [ "$CIRCLE_BRANCH" == "" ]; then
    echo -n "Publish packages to S3? [y/N] "
    read response
    response=`echo $response | tr 'A-Z' 'a-z'`
    if [ "x$response" == "xy" ]; then
        echo "Publishing packages to S3."
        if [ ! -e "$AWS_FILE" ]; then
            echo "$AWS_FILE does not exist -- aborting."
            cleanup_exit 1
        fi

        # Upload .deb and .rpm packages
        for filepath in `ls *.{deb,rpm}`; do
            echo "Uploading $filepath to S3"
            filename=`basename $filepath`
            echo "Uploading $filename to s3://get.influxdb.org/telegraf/$filename"
            AWS_CONFIG_FILE=$AWS_FILE aws s3 cp $filepath \
                s3://get.influxdb.org/telegraf/$filename \
                --acl public-read --region us-east-1
            if [ $? -ne 0 ]; then
                echo "Upload failed -- aborting".
                cleanup_exit 1
            fi
            rm $filepath
        done

        # Make and upload linux amd64, 386, and arm
        make build-linux-bins
        for b in `ls telegraf_*`; do
            zippedbin=${b}_${VERSION}.tar.gz
            # Zip the binary
            tar -zcf $TMP_WORK_DIR/$zippedbin ./$b
            echo "Uploading binary: $zippedbin to S3"
            AWS_CONFIG_FILE=$AWS_FILE aws s3 cp $TMP_WORK_DIR/$zippedbin \
                s3://get.influxdb.org/telegraf/$zippedbin \
                --acl public-read --region us-east-1
            if [ $? -ne 0 ]; then
                echo "Binary upload failed -- aborting".
                cleanup_exit 1
            fi
        done
    else
        echo "Not publishing packages to S3."
    fi
fi

###########################################################################
# All done.

echo -e "\nPackaging process complete."
cleanup_exit 0
