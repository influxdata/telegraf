# Telegraf customization tool

Telegraf's `custom_builder` is a tool to select the plugins compiled into the
Telegraf binary. By doing so, Telegraf can become smaller, saving both disk
space and memory if only a sub-set of plugins is selected.

## Building

To build `custom_builder` run the following command:

```shell
# make build_tools
```

The resulting binary is located in the `tools/custom_builder` folder.

## Running

The easiest way of building a customized Telegraf is to use your
Telegraf configuration file(s). Assuming your configuration is
in `/etc/telegraf/telegraf.conf` you can run

```shell
# ./tools/custom_builder/custom_builder --config /etc/telegraf/telegraf.conf
```

to build a Telegraf binary tailored to your configuration.
You can also specify a configuration directory similar to
Telegraf itself. To additionally use the configurations in
`/etc/telegraf/telegraf.d` run

```shell
# ./tools/custom_builder/custom_builder                      \
    --config     /etc/telegraf/telegraf.conf \
    --config-dir /etc/telegraf/telegraf.d
```

Configurations can also be retrieved from remote locations just
like for Telegraf.

```shell
# ./tools/custom_builder/custom_builder --config http://myserver/telegraf.conf
```

will download the configuration from `myserver`.

The `--config` and `--config-dir` option can be used multiple times.
In case you want to deploy Telegraf to multiple systems with
different configurations, simply specify the super-set of all
configurations you have. `custom_builder` will figure out the list
for you

```shell
# ./tools/custom_builder/custom_builder             \
    --config system1/telegraf.conf  \
    --config system2/telegraf.conf  \
    --config ...                    \
    --config systemN/telegraf.conf  \
    --config-dir system1/telegraf.d \
    --config-dir system2/telegraf.d \
    --config-dir ...                \
    --config-dir systemN/telegraf.d
```

The Telegraf customization uses
[Golang's build-tags](https://pkg.go.dev/go/build#hdr-Build_Constraints) to
select the set of plugins. To see which tags are set use the `--tags` flag.

To get more help run

```shell
# ./tools/custom_builder/custom_builder --help
```
