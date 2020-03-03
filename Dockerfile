FROM quay.io/influxdb/telegraf-ci:1.12.14 AS BUILD

WORKDIR /go/src/github.com/influxdata/telegraf

COPY Gopkg.lock .
COPY Gopkg.toml .
COPY Makefile .

RUN make deps

COPY Gopkg.lock .
COPY Gopkg.toml .
COPY Makefile .

COPY accumulator.go .
COPY input.go .
COPY metric.go .
COPY processor.go .
COPY aggregator.go .
COPY output.go .
COPY plugin.go .
COPY cmd cmd
COPY filter filter
COPY metric metric
COPY plugins plugins
COPY testutil testutil
COPY agent agent
COPY docs docs
COPY internal internal
COPY scripts scripts
COPY etc etc
COPY logger logger
COPY selfstat selfstat

COPY .git .git

RUN make check

RUN make test

RUN make package

FROM alpine:3.11.3

COPY --from=BUILD /go/src/github.com/influxdata/telegraf/build/linux/static_amd64/telegraf /usr/local/bin

WORKDIR /etc/telegraf

CMD ["telegraf", "--config", "/etc/telegraf/telegraf.conf"]