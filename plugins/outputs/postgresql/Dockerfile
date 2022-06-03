# This Dockerfile can be used to build an image including the pguint extension.
#
#   docker build -t postgres:pguint .
#   docker run -d --name postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust postgres:pguint
#   docker logs -f postgres 2>&1 | grep -q 'listening on IPv4 address "0.0.0.0", port 5432'
#   go test

# Tag from https://hub.docker.com/_/postgres?tab=tags
ARG POSTGRES_TAG=latest

ARG PGUINT_REPO
ARG PGUINT_RELEASE

FROM postgres:${POSTGRES_TAG}

RUN apt-get update && apt-get install -y build-essential curl postgresql-server-dev-${PG_MAJOR}=${PG_VERSION}

ENV PGUINT_REPO=${PGUINT_REPO:-phemmer/pguint}
ENV PGUINT_REF=${PGUINT_REF:-fix-getmsgint64}
RUN mkdir /pguint && cd /pguint && \
    curl -L https://github.com/${PGUINT_REPO}/tarball/${PGUINT_REF} | tar -zx --strip-components=1 && \
    make && make install && \
    echo 'CREATE EXTENSION uint;' > /docker-entrypoint-initdb.d/uint.sql && \
    echo '\\c template1' >> /docker-entrypoint-initdb.d/uint.sql && \
    echo 'CREATE EXTENSION uint;' >> /docker-entrypoint-initdb.d/uint.sql

