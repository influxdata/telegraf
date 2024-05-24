# Quick Start

The following demos getting started with Telegraf quickly using Docker to
monitor the local system.

## Install

This example will use Docker to launch a Telegraf container:

```shell
docker pull telegraf
```

Refer to the [Install Guide][] for the full list of ways to install Telegraf.

[Install Guide]: /docs/INSTALL_GUIDE.md

## Configure

Telegraf requires a configuration to start up. A configuration requires at least
one input to collect data from and one output to send data to. The configuration
file is a [TOML][] file.

[TOML]: /docs/TOML.md

```sh
$ cat config.toml
[[inputs.cpu]]
[[inputs.mem]]
[[outputs.file]]
```

The above enables two inputs, CPU and Memory, and one output file. The inputs
will collect usage information about the CPU and Memory, while the file output
is used to print the metrics to STDOUT.

Note that defining plugins to use are a TOML array of tables. This means users
can define a plugin multiple times. This is more useful with other plugins that
may need to connect to different endpoints.

## Launch

With the image downloaded and a config file created, launch the image:

```sh
docker run --rm --volume $PWD/config.toml:/etc/telegraf/telegraf.conf telegraf
```

The user will see some initial information print out about which config file was
loaded, version information, and what plugins were loaded. After the initial few
seconds metrics will start to print out.

## Next steps

To go beyond this quick start, users should consider the following:

1. Determine where you want to collect data or metrics from and look at the
  available [input plugins][]
2. Determine where you want to send metrics to and look at the available
  [output plugins][]
3. Look at the [install guide][] for the complete list of methods to deploy and
  install Telegraf
4. If parsing arbitrary data or sending metrics or logs to Telegraf, read
  through the [parsing data][] guide.

[input plugins]: /plugins/inputs
[output plugins]: /plugins/outputs
[parsing data]: /docs/PARSING_DATA.md
