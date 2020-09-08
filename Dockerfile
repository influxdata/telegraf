FROM golang:1.11.0 as builder
ENV DEP_VERSION 0.5.0
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 && chmod +x /usr/local/bin/dep
WORKDIR /go/src/github.com/influxdata/telegraf
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only
COPY . /go/src/github.com/influxdata/telegraf
RUN go install ./cmd/...

FROM buildpack-deps:stretch-curl
COPY --from=builder /go/bin/* /usr/bin/
COPY etc/telegraf.conf /etc/telegraf/telegraf.conf

EXPOSE 8125/udp 8092/udp 8094

COPY docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["telegraf"]
