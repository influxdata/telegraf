#
# Dockerfile to build telegraf in a container.
#

#
# Global build args that can be passed from the builder:
#
ARG BASE_GOLANG_VER="1.21"

ARG TELEGRAF_BUILD_ROOT="/root/go/telegraf"

# Typically set to any GO specific environment variables; see build.sh for how
# this is used.
ARG BUILD_GO_OPTS=


# Start with the base golang container and get a custom golang image.
FROM "golang:${BASE_GOLANG_VER}-bookworm" AS golang
    RUN ln -sf /bin/bash /bin/sh


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
    ARG BUILD_GO_OPTS

    RUN \
        export LDFLAGS="-w -s" CGO_ENABLED=0 ${BUILD_GO_OPTS} \
    && \
        git config --global user.email "dev@extremenetworks.com" \
    && \
        git config --global user.name "Dev Extreme" \
    && \
        export \
    && \
        make telegraf && \
        :


# Copy the binary to an empty container.
FROM scratch AS binary
    ARG TELEGRAF_BUILD_ROOT

    WORKDIR /

    COPY --from=build "${TELEGRAF_BUILD_ROOT}/telegraf" /usr/bin/

    ENTRYPOINT ["/usr/bin/telegraf"]
    CMD ["--help"]
