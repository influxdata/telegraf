# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
RUN apk add --no-cache git ca-certificates make bash
WORKDIR /src

COPY . .

ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BRANCH="unknown"

# Build a static-ish linux binary (common for Telegraf containers)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
  go build -trimpath \
    -ldflags "-s -w \
      -X main.version=${VERSION} \
      -X main.commit=${COMMIT} \
      -X main.branch=${BRANCH}" \
    -o /out/telegraf ./cmd/telegraf

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -s /sbin/nologin telegraf

COPY --from=build /out/telegraf /usr/bin/telegraf

# Most Telegraf images run telegraf directly; config is mounted at runtime
USER telegraf
ENTRYPOINT ["telegraf"]
