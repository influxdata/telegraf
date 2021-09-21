
# Telegraf

![tiger](TelegrafTiger.png "tiger")

[![Circle CI](https://circleci.com/gh/influxdata/telegraf.svg?style=svg)](https://circleci.com/gh/influxdata/telegraf) [![Docker pulls](https://img.shields.io/docker/pulls/library/telegraf.svg)](https://hub.docker.com/_/telegraf/) [![Total alerts](https://img.shields.io/lgtm/alerts/g/influxdata/telegraf.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/influxdata/telegraf/alerts/)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://www.influxdata.com/slack)

Telegraf is an agent for collecting, processing, aggregating, and writing metrics.

Design goal:
- Have a minimal memory footprint with a plugin system so that developers in the community can easily add support for collecting metrics.

Telegraf is plugin-driven and has the concept of 4 distinct plugin types:

1. [Input Plugins](#input-plugins) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](#processor-plugins) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](#aggregator-plugins) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](#output-plugins) write metrics to various destinations

New plugins are designed to be easy to contribute, pull requests are welcomed and we work to incorporate as many pull requests as possible. If none of the internal plugins fit your needs, you could have a look at the
[list of external plugins](EXTERNAL_PLUGINS.md).

## Minimum Requirements

Telegraf shares the same [minimum requirements][] as Go:
- Linux kernel version 2.6.23 or later
- Windows 7 or later
- FreeBSD 11.2 or later
- MacOS 10.11 El Capitan or later

[minimum requirements]: https://github.com/golang/go/wiki/MinimumRequirements#minimum-requirements

## Installation:

You can download the binaries directly from the [downloads](https://www.influxdata.com/downloads) page
or from the [releases](https://github.com/influxdata/telegraf/releases) section.

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telegraf

### From Source:

Telegraf requires Go version 1.14 or newer, the Makefile requires GNU make.

1. [Install Go](https://golang.org/doc/install) >=1.14 (1.15 recommended)
2. Clone the Telegraf repository:
   ```
   cd ~/src
   git clone https://github.com/influxdata/telegraf.git
   ```
3. Run `make` from the source directory
   ```
   cd ~/src/telegraf
   make
   ```

### Changelog

View the [changelog](/CHANGELOG.md) for the latest updates and changes by
version.

### Nightly Builds

[Nightly](/docs/NIGHTLIES.md) builds are available, generated from the master branch.

### 3rd Party Builds

Builds for other platforms or package formats are provided by members of the Telegraf community. These packages are not built, tested or supported by the Telegraf project or InfluxData, we make no guarantees that they will work. Please get in touch with the package author if you need support.

* Windows
  * [Chocolatey](https://chocolatey.org/packages/telegraf) by [ripclawffb](https://chocolatey.org/profiles/ripclawffb)
  * [Scoop](https://github.com/ScoopInstaller/Main/blob/master/bucket/telegraf.json)
* Linux
  * [Snap](https://snapcraft.io/telegraf) by Laurent Sesquès (sajoupa)

## How to use it:

See usage with:

```
telegraf --help
```

#### Generate a telegraf config file:

```
telegraf config > telegraf.conf
```

#### Generate config with only cpu input & influxdb output plugins defined:

```
telegraf --section-filter agent:inputs:outputs --input-filter cpu --output-filter influxdb config
```

#### Run a single telegraf collection, outputting metrics to stdout:

```
telegraf --config telegraf.conf --test
```

#### Run telegraf with all plugins defined in config file:

```
telegraf --config telegraf.conf
```

#### Run telegraf, enabling the cpu & memory input, and influxdb output plugins:

```
telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb
```

## Documentation

[Latest Release Documentation][release docs].

For documentation on the latest development code see the [documentation index][devel docs].

[release docs]: https://docs.influxdata.com/telegraf
[developer docs]: docs
- [Input Plugins](/telegraf/docs/INPUTS.md)
- [Output Plugins](/telegraf/docs/OUTPUTS.md)
- [Processor Plugins](/telegraf/docs/PROCESSORS.md)
- [Aggregator Plugins](/telegraf/docs/AGGREGATORS.md)


## Contributing

There are many ways to contribute:
- Fix and [report bugs](https://github.com/influxdata/telegraf/issues/new)
- [Improve documentation](https://github.com/influxdata/telegraf/issues?q=is%3Aopen+label%3Adocumentation)
- [Review code and feature proposals](https://github.com/influxdata/telegraf/pulls)
- Answer questions and discuss here on github and on the [Community Site](https://community.influxdata.com/)
- [Contribute plugins](CONTRIBUTING.md)
- [Contribute external plugins](docs/EXTERNAL_PLUGINS.md)