#
# Dockerfile to build telegraf in a container.
#

#
# Global build args that can be passed from the builder:
#
ARG BASE_GOLANG_VER="1.21"
ARG FIPS_GOLANG_VER="${BASE_GOLANG_VER}"

ARG GOLANG_FIPS_BUILD_ROOT="/root/go/golang-fips"
ARG TELEGRAF_BUILD_ROOT="/root/go/telegraf"

# When BUILD_GO_FIPS=1, build a dynamically linked executable using go patched
# with golang-fips patch and cgo (CGO_ENABLED=1). Otherwise, build a statically
# linked executable with the standard go.
ARG BUILD_GO_FIPS=

# Typically set to any GO specific environment variables; see build.sh for how
# this is used.
ARG BUILD_GO_OPTS=


# Start with the base golang container and get a custom golang image.
FROM "golang:${BASE_GOLANG_VER}-bookworm" AS golang
    ARG \
        BASE_GOLANG_VER \
        FIPS_GOLANG_VER \
        GOLANG_FIPS_BUILD_ROOT \
        BUILD_GO_FIPS
    ARG GO_SRC_BRANCH="release-branch.go${FIPS_GOLANG_VER}"
    ARG GOLANG_FIPS_BRANCH="go${FIPS_GOLANG_VER}-fips-release"

    ADD --keep-git-dir=true "https://github.com/golang-fips/go.git#${GOLANG_FIPS_BRANCH}" "${GOLANG_FIPS_BUILD_ROOT}"

    RUN ln -sf /bin/bash /bin/sh
    RUN \
        if [[ "${BUILD_GO_FIPS}" == "1" ]]; then \
            git config --global user.email "dev@extremenetworks.com" && \
            git config --global user.name "Dev Extreme" && \
            cd "${GOLANG_FIPS_BUILD_ROOT}" && \
            ./scripts/full-initialize-repo.sh "${GO_SRC_BRANCH}" && \
            : ; \
        fi \
    && \
        :


# Copy telegraf sources into the golang build environment.
FROM golang AS source
    ARG TELEGRAF_BUILD_ROOT

    WORKDIR "${TELEGRAF_BUILD_ROOT}"

    COPY .git               .git
    COPY agent              agent
    COPY cmd                cmd
    COPY config             config
    COPY filter             filter
    COPY internal           internal
    COPY logger             logger
    COPY metric             metric
    COPY models             models
    COPY plugins            plugins
    COPY selfstat           selfstat
    COPY *.go go.*          ./
    COPY build_version.txt  ./
    COPY Makefile           ./


# Build telegraf using the custom golang container.
FROM source AS build
    ARG \
        GOLANG_FIPS_BUILD_ROOT \
        BUILD_GO_FIPS \
        BUILD_GO_OPTS

    RUN \
        export LDFLAGS="-w -s" ${BUILD_GO_OPTS} \
    && \
        git config --global user.email "dev@extremenetworks.com" \
    && \
        git config --global user.name "Dev Extreme" \
    && \
        export \
    && \
        if [[ "${BUILD_GO_FIPS}" == "1" ]]; then \
            apt-get update --yes && \
            apt-get install --yes --no-install-recommends --no-install-suggests libssl-dev && \
            if [[ "${GOARCH}" == "mips" ]]; then \
                apt-get install --yes --no-install-recommends --no-install-suggests gcc-multilib && \
                : ; \
            fi && \
            rm -rf /var/lib/apt/lists/* && \
            export PATH="${GOLANG_FIPS_BUILD_ROOT}/go/bin:${PATH}" && \
            export CGO_ENABLED=1 && \
            : ; \
        else \
            export CGO_ENABLED=0 && \
            : ; \
        fi \
    && \
        which go \
    && \
        go env \
    && \
        make telegraf \
    && \
        :


# Copy the binary to an empty container.
FROM scratch AS binary
    ARG TELEGRAF_BUILD_ROOT

    WORKDIR /

    COPY --from=build "${TELEGRAF_BUILD_ROOT}/telegraf" /usr/bin/

    ENTRYPOINT ["/usr/bin/telegraf"]
    CMD ["--help"]
