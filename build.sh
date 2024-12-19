#!/usr/bin/env bash

# Simple ExtremeNetworks script to facilitate building and
# uploading the telegraf utility.
#
# When creating the tar file name for uploading to artifactory, we concatenate the
# telegraf build_version.txt and our own extr_version.txt to create the tar file name.
# The extr_version is the version used when we make Extreme specific changes on
# a particular telegraf branch.
#

show_usage()
{
    echo "usage: ${0} <arch> {build | upload}"
    echo " . arch   : valid architectures: arm, arm64, x86_64, mips"
    echo " . build  : build and tar utility for specified architecture"
    echo " . upload : upload specified architecture's tar to Artifactory"
}

do_build()
{
    local _extr_arch="${1}"
    local _target_tarball="${2}"

    # set the target architecture for the executable
    local _target_arch="${_extr_arch}"
    if [[ "${_target_arch}" = "x86_64" ]]; then
        _target_arch="amd64"
    fi

    # options passed to GO
    local _go_opts=""
    _go_opts+="${_go_opts:+ }GOOS=linux"
    _go_opts+="${_go_opts:+ }GOARCH=${_target_arch}"
    if [[ "${_target_arch}" == "arm" ]]; then
        # set GOARM only when building for 32 bit ARM
        _go_opts+="${_go_opts:+ }GOARM=${arm32_type}"
    fi

    # force GO to use alternate location of cache and modules
    local _gocache="$(pwd -P)/.tmp/go-cache"
    local _gomodcache="$(pwd -P)/.tmp/go-mod"
    _go_opts+="${_go_opts:+ }GOCACHE=${_gocache@Q}"
    _go_opts+="${_go_opts:+ }GOMODCACHE=${_gomodcache@Q}"

    make clean
    rm -f "${_target_tarball}"
    env LDFLAGS="-w -s" CGO_ENABLED=0 ${_go_opts} make telegraf
    tar -cf "${_target_tarball}" telegraf MIT generic_MIT
    rm -f telegraf
}

do_upload()
{
    local _extr_arch="${1}"
    local _target_tarball="${2}"

    if [[ ! -f "${_target_tarball}" ]]; then
        echo "info: ${_target_tarball} not found; building first..."
        do_build "${_extr_arch}" "${_target_tarball}"
        if [[ ! -f "${_target_tarball}" ]]; then
            echo "error: could not find or build '${_target_tarball}' tarball"
            exit 1
        fi
    fi

    # make sure jfrog config is set up
    if ! jfrog config show "${afy_server_name}" &>/dev/null ; then
        echo "Could not find ${afy_server_name} Artifactory config.  Let's create one..."
        echo "Accept defaults and use your corporate password."; echo
        jfrog config add --url "${afy_server_url}" --user "${USER}" "${afy_server_name}"
        if ! jfrog config show "${afy_server_name}" &>/dev/null ; then
            echo "error: failed to configure ${afy_server_name} Artifactory server access"
            exit 1
        fi
    fi

    jfrog rt upload "${_target_tarball}" "${aft_repo}/${_extr_arch}/${_target_tarball}" --server-id "${afy_server_name}"
}

#######################
# execution starts here
#######################

main()
{
    # all hard-coded globals go here
    arm32_type="5"              # default ARM type for "arm" architecture
    afy_server_name="Salem"
    afy_server_url="http://engartifacts1.extremenetworks.com:8081"
    aft_repo="xos-binaries-local-release/telegraf"

    local _extr_arch="${1}"
    local _script_action="${2}"

    # verify inputs - count of args
    if [[ ${#} -ne 2 ]]; then
        echo "error: incorrect number of arguments '${#}'"
        show_usage
        exit 1
    fi

    # verify inputs - architecture
    case "${_extr_arch}" in
        arm64 | mips | x86_64 | arm)
            :
            ;;
        --help | -h | -?)
            show_usage
            exit 0
            ;;
        *)
            echo "error: invalid architecture '${1}'"
            show_usage
            exit 1
    esac

    # verify inputs - action
    case "${_script_action}" in
        build | upload)
            :
            ;;
        --help | -h | -?)
            show_usage
            exit 0
            ;;
        *)
            echo "error: invalid action '$2'"
            show_usage
            exit 1
    esac

    # grab version strings and set the name of the target tarball
    local _telegraf_version="$(< build_version.txt)"
    local _extr_version="$(< extr_version.txt)"
    local _target_tarball="telegraf_${_extr_arch}_${_telegraf_version}.${_extr_version}.tar"

    "do_${_script_action}" "${_extr_arch}" "${_target_tarball}"
}

set -o errexit
set -o pipefail

main "${@}"
