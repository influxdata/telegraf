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
    if [[ "${_target_arch}" == "arm" ]]; then
        # set GOARM only when building for 32 bit ARM
        _go_opts+="${_go_opts:+ }GOARM=${arm32_type}"
    elif [[ "${_target_arch}" == "mips" ]]; then
        # set GOARCH to build a 32 bit executable for MIPS
        _go_opts+="${_go_opts:+ }GOARCH=${_target_arch}"
    fi

    rm -f "${_target_tarball}"

    #
    # REF: https://docs.docker.com/build/building/multi-platform/#qemu
    #
    # We do not leverage the cross compilation feature of golang to build the executable;
    # in other words, GOOS and GOARCH are (typically) not passed to the go compiler.
    # Instead we build on a node (i.e. container) that has the same CPU as that target;
    # in other words:
    #       arch of container running golang compiler
    #           == target arch
    # This node can be one of the following:
    #   1. the host on which docker is running (typically amd64 or arm64) i.e.
    #       arch of host running the golang containers used to build the executable
    #           == arch of container running golang compiler
    #   2. a container running under qemu (qemu is provided by the host) i.e.
    #       arch of host running the golang containers used to build the executable
    #           != arch of container running golang compiler
    #
    # For #2 above, the container effectively runs under qemu. You might need to run
    # the following command to install qemu on the host:
    #   docker run --privileged --rm tonistiigi/binfmt --install linux/${_container_arch}
    #
    # NOTE: We are not building a multi-platform image as that requires some capabilities
    # in the image store that might/might not be present in the docker install.
    #

    # architecture of the container to use for building
    local _container_arch="${_target_arch}"
    if [[ "${_container_arch}" == "mips" ]]; then
        # use "mips64le" for the container arch when target arch is "mips"
        _container_arch="mips64le"
    fi

    local _dockerfile_stage="binary"
    local _target_image="telegraf/${_dockerfile_stage}/${_target_arch}:$(git describe --dirty)"
    docker buildx build --progress plain \
        --build-arg BUILD_GO_OPTS="${_go_opts}" \
        --platform "linux/${_container_arch}" \
        --tag "${_target_image}" \
        --target "${_dockerfile_stage}" \
        .

    local _copy_container="$(docker container create --quiet "${_target_image}")"
    docker container cp "${_copy_container}:/usr/bin/telegraf" telegraf
    docker container rm "${_copy_container}"

    docker image rm "${_target_image}"

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
