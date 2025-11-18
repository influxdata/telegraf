FROM golang:1.25 AS builder

WORKDIR /src
COPY . .

RUN go mod download

RUN go build -o /telegraf ./cmd/telegraf


FROM debian:stable-slim

COPY --from=builder /telegraf /usr/bin/telegraf
COPY telegraf.conf /etc/telegraf/telegraf.conf

ENTRYPOINT ["/usr/bin/telegraf"]
CMD ["--config", "/etc/telegraf/telegraf.conf"]
