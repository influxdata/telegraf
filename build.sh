#!/bin/bash

# Simple ExtremeNetworks script to facilitate building and
# uploading the telegraf utility.
#
# When creating the tar file name for uploading to artifactory, we concatenate the
# telegraf build_version.txt and our own extr_version.txt to create the tar file name.
# The extr_version is the version used when we make Extreme specific changes on
# a particular telegraf branch.
#
set -e

usage()
{
    echo "usage: $0 arch {build | upload}"
    echo " . arch  : valid architectures: arm64, x86_64, mips"
    echo " . build : build and tar utility for specified architecture"
    echo " . upload: upload specified architecture's tar to Artifactory"
}

build()
{
    make clean
    rm -f ${target}
    make CGO_ENABLED=0 GOOS=linux GOARCH=${bld_arch} GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org
    tar -cf ${target} telegraf
    rm -f telegraf
}

upload()
{
    if [ ! -f ${target} ]; then
        echo "info: ${target} not found; building first..."
        build
        if [ ! -f ${target} ]; then
            echo "error: could not find or build '${target}' tarball"
            exit 1
        fi
    fi

    # make sure jfrog config is set up
    if ! jfrog config show ${salem} &>/dev/null ; then
        echo "Could not find Salem Artifactory config.  Let's create one..."
        echo "Accept defaults and use your corporate password."; echo ""
        jfrog config add --url http://engartifacts1.extremenetworks.com:8081 --user $(whoami) ${salem}
        if ! jfrog config show ${salem} &>/dev/null ; then
            echo "error: failed to configure ${salem} Artifactory server access"
            exit 1
        fi
    fi
    jfrog rt upload ${target} xos-binaries-local-release/telegraf/${arch}/${target} --server-id ${salem}
}

#######################
# execution starts here
#######################
if [[ -z "$1" || "$1" == "--help" || "$1" == "-h" || "$1" == "?" ]]; then
    usage
    exit 0
fi

# force go to use alternate location of modules
go env -w GOMODCACHE=/opt/go/pkg/mod
go env -w GOCACHE=$(pwd -P)/.cache/go-build
# grab version strings
telegraf_version=$(cat build_version.txt)
extr_version=$(cat extr_version.txt)
salem=Salem

# set architecture name and build architecture as used by golang
arch=$1
if [ "$arch" = "x86_64" ]; then
    bld_arch=amd64
else
    bld_arch=${arch}
fi
target=telegraf_${arch}_${telegraf_version}.${extr_version}.tar

# check action argument
case $2 in
    build | upload)
        action=$2
        ;;
    *)
        echo "error: invalid action '$2'"
        usage
        exit 1
esac

# perform action
case $1 in
    arm64 | mips | x86_64)
        $action
        ;;
    *)
        echo "error: invalid architecture '$1'"
        usage
        exit 1
esac

