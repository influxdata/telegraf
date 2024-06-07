# Installation

Telegraf compiles to a single static binary, which makes it easy to install.
Both InfluxData and the community provide for a wide range of methods to install
Telegraf from. For details on each release, view the [changelog][] for the
latest updates and changes by version.

[changelog]: /CHANGELOG.md

There are many places to obtain Telegraf from:

* [Binary downloads](#binary-downloads)
* [Homebrew](#homebrew)
* [InfluxData Linux package repository](#influxdata-linux-package-repository)
* [Official Docker images](#official-docker-images)
* [Helm charts](#helm-charts)
* [Nightly builds](#nightly-builds)
* [Build from source](#build-from-source)
* [Custom builder](#custom-builder)

## Binary downloads

Binary downloads for a wide range of architectures and operating systems are
available from the [InfluxData downloads][] page or from the
[GitHub Releases][] page.

[InfluxData downloads]: https://www.influxdata.com/downloads
[GitHub Releases]: https://github.com/influxdata/telegraf/releases

## Homebrew

A [Homebrew Formula][] for Telegraf that updates after each release:

```shell
brew update
brew install telegraf
```

Note that the Homebrew organization builds Telegraf itself and does not use
binaries built by InfluxData. This is important as Homebrew builds with CGO,
which means there are some differences between the official binaries and those
found with Homebrew.

[Homebrew Formula]: https://formulae.brew.sh/formula/telegraf

## InfluxData Linux package repository

InfluxData provides a package repo that contains both DEB and RPM packages.

### DEB

For DEB-based platforms (e.g. Ubuntu and Debian) run the following to add the
repo GPG key and setup a new sources.list entry:

```shell
# influxdata-archive_compat.key GPG fingerprint:
#     9D53 9D90 D332 8DC7 D6C8 D3B9 D8FF 8E1F 7DF8 B07E
wget -q https://repos.influxdata.com/influxdata-archive_compat.key
echo '393e8779c89ac8d958f81f942f9ad7fb82a25e133faddaf92e15b16e6ac9ce4c influxdata-archive_compat.key' | sha256sum -c && cat influxdata-archive_compat.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg > /dev/null
echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list
sudo apt-get update && sudo apt-get install telegraf
```

### RPM

For RPM-based platforms (e.g. RHEL, CentOS) use the following to create a repo
file and install telegraf:

```shell
# influxdata-archive_compat.key GPG fingerprint:
#     9D53 9D90 D332 8DC7 D6C8 D3B9 D8FF 8E1F 7DF8 B07E
cat <<EOF | sudo tee /etc/yum.repos.d/influxdata.repo
[influxdata]
name = InfluxData Repository - Stable
baseurl = https://repos.influxdata.com/stable/\$basearch/main
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdata-archive_compat.key
EOF
sudo yum install telegraf
```

## Official Docker images

Telegraf is available as an [Official image][] on DockerHub. Official images
are a curated set of Docker Images that also automatically get security updates
from Docker, follow a set of best practices, and are available via a shortcut
syntax which omits the organization.

InfluxData maintains a Debian and Alpine based image across the last three
minor releases. To pull the latest Telegraf images:

```shell
# latest Debian-based image
docker pull telegraf
# latest Alpine-based image
docker pull telegraf:alpine
```

See the [Telegraf DockerHub][] page for complete details on available images,
versions, and tags.

[official image]: https://docs.docker.com/trusted-content/official-images/
[Telegraf DockerHub]: https://hub.docker.com/_/telegraf

## Helm charts

A community-supported [helm chart][] is also available:

```shell
helm repo add influxdata https://helm.influxdata.com/
helm search repo influxdata
```

[helm chart]: https://github.com/influxdata/helm-charts/tree/master/charts/telegraf

## Nightly builds

[Nightly builds][] are available and are generated from the master branch each
day at around midnight UTC. The artifacts include both binary packages, RPM &
DEB packages, as well as nightly Docker images that are hosted on [quay.io][].

[Nightly builds]: /docs/NIGHTLIES.md
[quay.io]: https://quay.io/repository/influxdb/telegraf-nightly?tab=tags&tag=latest

## Build from source

Telegraf generally follows the latest version of Go and requires GNU make to use
the Makefile for builds.

On Windows, the makefile requires the use of a bash terminal to support all
makefile targets. An easy option to get bash for windows is using the version
that comes with [git for windows](https://gitforwindows.org/).

1. [Install Go](https://golang.org/doc/install)
2. Clone the Telegraf repository:

   ```shell
   git clone https://github.com/influxdata/telegraf.git
   ```

3. Run `make build` from the source directory

   ```shell
   cd telegraf
   make build
   ```

## Custom builder

Telegraf also provides a way of building a custom minimized binary using the
[custom builder][]. This takes a user's configuration file(s), determines what
plugins are required, and builds a binary with only those plugins. This greatly
reduces the size of the Telegraf binary.

[custom builder]: /tools/custom_builder
