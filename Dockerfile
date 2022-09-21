FROM quay.io/influxdb/telegraf-ci:1.19.1 AS BUILD

WORKDIR /go/src/github.com/influxdata/telegraf

COPY go.mod .
COPY go.sum .
COPY Makefile .

RUN make deps

COPY go.mod .
COPY go.sum .
COPY Makefile .
COPY build_version.txt .

COPY accumulator.go .
COPY aggregator.go .
COPY input.go .
COPY metric.go .
COPY output.go .
COPY parser.go .
COPY plugin.go .
COPY processor.go .
COPY agent agent
COPY assets assets
COPY cmd cmd
COPY config config
COPY docs docs
COPY etc etc
COPY filter filter
COPY hooks hooks
COPY internal internal
COPY logger logger
COPY metric metric
COPY models models
COPY plugins plugins
COPY scripts scripts
COPY selfstat selfstat
COPY testutil testutil
COPY tools tools


COPY .git .git

ENV CGO_ENABLED=1

# RUN make check

# RUN make test

RUN make

FROM alpine:3

RUN apk add libc6-compat

COPY --from=BUILD /go/src/github.com/influxdata/telegraf/telegraf /usr/local/bin

WORKDIR /etc/telegraf

CMD ["telegraf", "--config", "/etc/telegraf/telegraf.conf"]
