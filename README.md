# Telegraf

<div align="center">
  <img width="600" src="assets/telegraf-logo.png" alt="Telegraf Logo">
</div>



<p align="center">
  <a href="https://github.com/influxdata/telegraf/blob/master/CONTRIBUTING.md">
  <img alt="Contribute" src="https://img.shields.io/badge/Contribute%20To%20Telegraf-orange.svg?logo=influx&style=for-the-badge" />
  </a>
  
  <a href="https://www.influxdata.com/slack">
  <img alt="Slack Status" src="https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=for-the-badge" />
  </a>
    
  <a href="https://circleci.com/gh/influxdata/telegraf">
  <img alt="Circle CI" src="https://circleci.com/gh/influxdata/telegraf.svg?style=svg" />
  </a>
  
  <a href="https://hub.docker.com/_/telegraf/">
  <img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/library/telegraf.svg" />
  </a>
    
  <a href="https://godoc.org/github.com/influxdata/telegraf">
  <img alt="GoDoc" src="https://godoc.org/github.com/influxdata/telegraf?status.svg" />
  </a>
  
  <a href="https://github.com/influxdata/telegraf/blob/master/LICENSE">
  <img alt="License" src="https://img.shields.io/github/license/influxdata/influxdb" />
  </a>
</p>
<h3 align="center">
    <a href="https://www.influxdata.com/">Website</a>
    •
    <a href="https://docs.influxdata.com/telegraf">Documentation</a>
    •
    <a href="https://university.influxdata.com/">InfluxDB University</a>
    •
    <a href="https://www.influxdata.com/blog/">Blog</a>
</h3>



---

<p align="center">
   <img src="assets/TelegrafTiger.png" "tiger"/>
</p>

Telegraf is an agent for collecting, processing, aggregating, and writing metrics. Based on a
plugin system to enable developers in the community to easily add support for additional
metric collection. There are four distinct types of plugins:

1. [Input Plugins](/docs/INPUTS.md) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](/docs/PROCESSORS.md) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](/docs/AGGREGATORS.md) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](/docs/OUTPUTS.md) write metrics to various destinations

New plugins are designed to be easy to contribute, pull requests are welcomed, and we work to
incorporate as many pull requests as possible. Consider looking at the
[list of external plugins](EXTERNAL_PLUGINS.md) as well.


## Minimum Requirements

Telegraf shares the same [minimum requirements][] as Go:

- Linux kernel version 2.6.23 or later
- Windows 7 or later
- FreeBSD 11.2 or later
- MacOS 10.11 El Capitan or later

[minimum requirements]: https://github.com/golang/go/wiki/MinimumRequirements#minimum-requirements

## Obtaining Telegraf

View the [changelog](/CHANGELOG.md) for the latest updates and changes by version.

### Binary Downloads

Binary downloads are available from the [InfluxData downloads](https://www.influxdata.com/downloads)
page or from each [GitHub Releases](https://github.com/influxdata/telegraf/releases) page.

### Package Repository

InfluxData also provides a package repo that contains both DEB and RPM downloads.

For deb-based platforms (e.g. Ubuntu and Debian) run the following to add the
repo key and setup a new sources.list entry:

```shell
# influxdb.key GPG Fingerprint: 05CE15085FC09D18E99EFB22684A14CF2582E0C5
wget -q https://repos.influxdata.com/influxdb.key
echo '23a1c8836f0afc5ed24e0486339d7cc8f6790b83886c4c96995b88a061c5bb5d influxdb.key' | sha256sum -c && cat influxdb.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdb.gpg > /dev/null
echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdb.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list
sudo apt-get update && sudo apt-get install telegraf
```

For RPM-based platforms (e.g. RHEL, CentOS) use the following to create a repo
file and install telegraf:

```shell
# influxdb.key GPG Fingerprint: 05CE15085FC09D18E99EFB22684A14CF2582E0C5
cat <<EOF | sudo tee /etc/yum.repos.d/influxdata.repo
[influxdata]
name = InfluxData Repository - Stable
baseurl = https://repos.influxdata.com/stable/\$basearch/main
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdb.key
EOF
sudo yum install telegraf
```

### Build From Source

Telegraf requires Go version 1.18 or newer, the Makefile requires GNU make.

On Windows, the makefile requires the use of a bash terminal to support all makefile targets.
An easy option to get bash for windows is using the version that comes with [git for windows](https://gitforwindows.org/).

1. [Install Go](https://golang.org/doc/install) >=1.18 (1.18.0 recommended)
2. Clone the Telegraf repository:

   ```shell
   git clone https://github.com/influxdata/telegraf.git
   ```

3. Run `make build` from the source directory

   ```shell
   cd telegraf
   make build
   ```

### Nightly Builds

[Nightly](/docs/NIGHTLIES.md) builds are available, generated from the master branch.

### 3rd Party Builds

Builds for other platforms or package formats are provided by members of theTelegraf community.
These packages are not built, tested, or supported by the Telegraf project or InfluxData. Please
get in touch with the package author if support is needed:

- [Ansible Role](https://github.com/rossmcdonald/telegraf)
- [Chocolatey](https://chocolatey.org/packages/telegraf) by [ripclawffb](https://chocolatey.org/profiles/ripclawffb)
- [Scoop](https://github.com/ScoopInstaller/Main/blob/master/bucket/telegraf.json)
- [Snap](https://snapcraft.io/telegraf) by Laurent Sesquès (sajoupa)

## Getting Started

<a href="https://university.influxdata.com/catalog/">
    <img src="assets/influxdbU-telegraf-banner.png" width="600"/>
</a>

See usage with:

```shell
telegraf --help
```

### Generate a telegraf config file

```shell
telegraf config > telegraf.conf
```

### Generate config with only cpu input & influxdb output plugins defined

```shell
telegraf config --section-filter agent:inputs:outputs --input-filter cpu --output-filter influxdb
```

### Run a single telegraf collection, outputting metrics to stdout

```shell
telegraf --config telegraf.conf --test
```

### Run telegraf with all plugins defined in config file

```shell
telegraf --config telegraf.conf
```

### Run telegraf, enabling the cpu & memory input, and influxdb output plugins

```shell
telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb
```

## Contribute to the Project

Telegraf is an MIT licensed open source project and we love our community. The fastest way to get something fixed is to open a PR. Check out our [contributing guide](CONTRIBUTING.md) if you're interested in helping out. Also, join us on our [Community Slack](https://influxdata.com/slack) or [Community Page](https://community.influxdata.com/) if you have questions or comments for our engineering teams.

If your completely new to Telegraf and InfluxDB, you can also enroll for free at [InfluxDB university](https://www.influxdata.com/university/) to take courses to learn more.

## Documentation

[Latest Release Documentation](https://docs.influxdata.com/telegraf/latest/)

For documentation on the latest development code see the [documentation index](/docs).

- [Input Plugins](/docs/INPUTS.md)
- [Output Plugins](/docs/OUTPUTS.md)
- [Processor Plugins](/docs/PROCESSORS.md)
- [Aggregator Plugins](/docs/AGGREGATORS.md)

## Additional Resources
- [Telegraf Tips and Tutorials](https://www.influxdata.com/blog/category/tech/telegraf/)
- [Telegraf Basics Course](https://university.influxdata.com/courses/telegraf-basics-tutorial/)
- [Data Collection with Telegraf Course](https://university.influxdata.com/courses/data-collection-with-telegraf-tutorial/)