# Packaging

## Package using Docker

This packaging method uses the CI images, and is very similar to how the
official packages are created on release.  This is the recommended method for
building the rpm/deb as it is less system dependent.

Pull the CI images from quay, the version corresponds to the version of Go
that is used to build the binary:
```
docker pull quay.io/influxdb/telegraf-ci:1.9.7
```

Start a shell in the container:
```
docker run -ti quay.io/influxdb/telegraf-ci:1.9.7 /bin/bash
```

From within the container:
```
go get -d github.com/influxdata/telegraf
cd /go/src/github.com/influxdata/telegraf

# Use tag of Telegraf version you would like to build
git checkout release-1.10
git reset --hard 1.10.2
make deps

# This builds _all_ platforms and architectures; will take a long time
./scripts/build.py --release --package
```

If you would like to only build a subset of the packages run this:

```
# Use the platform and arch arguments to skip unwanted packages:
./scripts/build.py --release --package --platform=linux --arch=amd64
```

From the host system, copy the build artifacts out of the container:
```
docker cp romantic_ptolemy:/go/src/github.com/influxdata/telegraf/build/telegraf-1.10.2-1.x86_64.rpm .
```
