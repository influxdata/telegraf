# ---- telegraf build environment... ----
FROM golang:latest as build

WORKDIR /go/src/github.com/influxdata/telegraf
COPY . .

RUN mkdir -p bin
RUN go build -o bin/builder ./builder

# ---- break here and ship this as its own image? ----

# build telegraf binary with target OS and ARCH and specified plugins
# TODO: how to pass these in?
# for GOOS and GOARCH values see:
#   https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
RUN bin/builder -GOOS="linux" -GOARCH="amd64" -plugins "inputs.cpu,processors.starlark,outputs.influxdb_v2,outputs.file,inputs.tail"
# outputs binary ./tmp/telegraf

# ---- now build a custom telegraf package with that binary? ----

# TODO: what to do about alpine here, have another build?
FROM telegraf:latest

# overwrite the existing binary
COPY --from=build /go/src/github.com/influxdata/telegraf/tmp/telegraf /usr/bin/telegraf