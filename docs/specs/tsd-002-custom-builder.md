# Telegraf Custom-Builder

## Objective

Provide a tool to build a customized, smaller version of Telegraf with only
the required plugins included.

## Keywords

tool, binary size, customization

## Overview

The Telegraf binary continues to grow as new plugins and features are added
and dependencies are updated. Users running on resource constraint systems
such as embedded-systems or inside containers might suffer from the growth.

This document specifies a tool to build a smaller Telegraf binary tailored to
the plugins configured and actually used, removing unnecessary and unused
plugins. The implementation should be able to cope with configured parsers and
serializers including defaults for those plugin categories.

Please note, the customization tool might not be available for older versions
of Telegraf. Furthermore, the degree of customization and thus the effective
size reduction might vary across versions.

The specified tool will *not* produce distribution packages or containers but
*only* the customized, static Telegraf binary.

Requirements to produce a customized Telegraf binary are listed below.

### Configuration files

The user has to provide one or more valid Telegraf configuration files or
configuration directories that are used with the produced binary later on. If
you plan to use the customized binary in multiple scenarios, the configuration
files should be a superset of all expected use-cases.

### Telegraf source code

To build a customized version of Telegraf you need access to the [Telegraf
source-code repository](https://github.com/influxdata/telegraf).

### Build tools

For compiling the customized binary you need the [Golang language](https://go.dev/)
as well as the `make` build system. The minimum required version of Golang
can be found in the *Build From Source* section of the `README.md` file of your
version. Both the `go` and the `make` command must be available in your path.

## Workflow

The first step is to download the Telegraf repository for the version you are
planning to customize.

```shell
$ git clone  --branch v1.29.5 --single-branch https://github.com/influxdata/telegraf.git
Cloning into 'telegraf'...
remote: Enumerating objects: 82314, done.
remote: Counting objects: 100% (82314/82314), done.
remote: Compressing objects: 100% (27977/27977), done.
remote: Total 82314 (delta 52910), reused 81858 (delta 52723), pack-reused 0
Receiving objects: 100% (82314/82314), 57.49 MiB | 11.79 MiB/s, done.
Resolving deltas: 100% (52910/52910), done.
Note: switching to '138d0d54add1e3dcd70592d069bed4218bae2bd2'.
...

$ cd telegraf
```

This will clone a specific version of Telegraf, `v1.29.5` in this case. You can
also use the latest master or download a source tarball or zip-archive.

Next, you need to build the customization tool itself

```shell
$ make build_tools
env -u GOOS -u GOARCH -u GOARM -- go build -o ./tools/custom_builder/custom_builder ./tools/custom_builder
env -u GOOS -u GOARCH -u GOARM -- go build -o ./tools/license_checker/license_checker ./tools/license_checker
env -u GOOS -u GOARCH -u GOARM -- go build -o ./tools/readme_config_includer/generator ./tools/readme_config_includer/generator.go
env -u GOOS -u GOARCH -u GOARM -- go build -o ./tools/config_includer/generator ./tools/config_includer/generator.go
env -u GOOS -u GOARCH -u GOARM -- go build -o ./tools/readme_linter/readme_linter ./tools/readme_linter
```

Check out the help for information on the usage of the tool

```shell
$ ./tools/custom_builder/custom_builder --help

This is a tool build Telegraf with a custom set of plugins. The plugins are
select according to the specified Telegraf configuration files. This allows
to shrink the binary size by only selecting the plugins you really need.
A more detailed documentation is available at
http://github.com/influxdata/telegraf/tools/custom_builder/README.md

...
```

To tailor Telegraf to the configuration files in the default locations use

```shell
$ ./tools/custom_builder/custom_builder --config /etc/telegraf/telegraf.conf --config-dir /etc/telegraf/telegraf.d
2024/02/27 17:31:56 Importing configuration file(s)...
2024/02/27 17:31:56 Found 3 configuration files...
-------------------------------------------------------------------------------
Enabled plugins:
-------------------------------------------------------------------------------
aggregators (0):
-------------------------------------------------------------------------------
inputs (24):
  apache                          plugins/inputs/apache
  cpu                             plugins/inputs/cpu
  diskio                          plugins/inputs/diskio
  disque                          plugins/inputs/disque
  elasticsearch                   plugins/inputs/elasticsearch
  exec                            plugins/inputs/exec
  haproxy                         plugins/inputs/haproxy
  kafka_consumer                  plugins/inputs/kafka_consumer
  leofs                           plugins/inputs/leofs
  mem                             plugins/inputs/mem
  memcached                       plugins/inputs/memcached
  mesos                           plugins/inputs/mesos
  mongodb                         plugins/inputs/mongodb
  mysql                           plugins/inputs/mysql
  net                             plugins/inputs/net
  nginx                           plugins/inputs/nginx
  ping                            plugins/inputs/ping
  postgresql                      plugins/inputs/postgresql
  prometheus                      plugins/inputs/prometheus
  rabbitmq                        plugins/inputs/rabbitmq
  redis                           plugins/inputs/redis
  rethinkdb                       plugins/inputs/rethinkdb
  swap                            plugins/inputs/swap
  system                          plugins/inputs/system
-------------------------------------------------------------------------------
outputs (2):
  influxdb                        plugins/outputs/influxdb
  kafka                           plugins/outputs/kafka
-------------------------------------------------------------------------------
parsers (2):
  influx                          plugins/parsers/influx
  json                            plugins/parsers/json
-------------------------------------------------------------------------------
processors (0):
-------------------------------------------------------------------------------
secretstores (0):
-------------------------------------------------------------------------------
serializers (1):
  influx                          plugins/serializers/influx
-------------------------------------------------------------------------------
2024/02/27 17:31:56 Running build...
CGO_ENABLED=0 go build -tags "custom,inputs.apache,inputs.cpu,inputs.diskio,inputs.disque,inputs.elasticsearch,inputs.exec,inputs.haproxy,inputs.kafka_consumer,inputs.leofs,inputs.mem,inputs.memcached,inputs.mesos,inputs.mongodb,inputs.mysql,inputs.net,inputs.nginx,inputs.ping,inputs.postgresql,inputs.prometheus,inputs.rabbitmq,inputs.redis,inputs.rethinkdb,inputs.swap,inputs.system,outputs.influxdb,outputs.kafka,parsers.influx,parsers.json,serializers.influx" -ldflags " -X github.com/influxdata/telegraf/internal.Commit=138d0d54 -X github.com/influxdata/telegraf/internal.Branch=HEAD -X github.com/influxdata/telegraf/internal.Version=1.29.5" ./cmd/telegraf
```

The resulting customized binary called `telegraf` or `telegraf.exe` should be
located in the current directory with a reduced size. Optionally, users can set
`GOOS` and `GOARCH` environment variables to compile Telegraf for different
architectures or platforms.

## Prior art

[PR #5809](https://github.com/influxdata/telegraf/pull/5809) and
[telegraf-lite-builder](https://github.com/influxdata/telegraf/tree/telegraf-lite-builder/cmd/telegraf-lite-builder):

- Uses docker
- Uses browser:
  - Generates a webpage to pick what options you want. User chooses plugins;
    does not take a config file
  - Build a binary, then minifies by stripping and compressing that binary
- Does some steps that belong in makefile, not builder
  - Special case for upx
  - Makes gzip, zip, tar.gz
- Uses gopkg.in?
- Can also work from the command line

[PR #8519](https://github.com/influxdata/telegraf/pull/8519)

- User chooses plugins OR provides a config file

[powers/telegraf-build](https://github.com/powersj/telegraf-build)

- User chooses plugins OR provides a config file
- Currently kept in separate repo
- Undoes changes to all.go files

[rawkode/bring-your-own-telegraf](https://github.com/rawkode/bring-your-own-telegraf)

- Users docker

## Additional information

You might be able to further reduce the binary size of Telegraf by removing
debugging information. This is done by adding `-w` and `-s` to the linker flags
before building `LDFLAGS="-w -s"`.

However, please note that this removes information helpful for debugging issues
in Telegraf.

Additionally, you can use a binary packer such as [UPX](https://upx.github.io/)
to reduce the required *disk* space. This compresses the binary and decompresses
it again at runtime. However, this does not reduce memory footprint at runtime.
