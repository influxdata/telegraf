# Telegraf Commands & Flags

The following page describes some of the commands and flags available via the
Telegraf command line interface.

## Usage

General usage of Telegraf, requires passing in at least one config file with
the plugins the user wishes to use:

```bash
telegraf --config config.toml
```

## Help

To get the full list of subcommands and flags run:

```bash
telegraf help
```

Here are some commonly used flags that users should be aware of:

* `--config-directory`: Read all config files from a directory
* `--debug`: Enable additional debug logging
* `--once`: Run one collection and flush interval then exit
* `--test`: Run only inputs, output to stdout, and exit

Check out the full help out for more available flags and options.

## Version

While telegraf will print out the version when running, if a user is uncertain
what version their binary is, run the version subcommand:

```bash
telegraf version
```

## Config

The config subcommand allows users to print out a sample configuration to
stdout. This subcommand can very quickly print out the default values for all
or any of the plugins available in Telegraf.

For example to print the example config for all plugins run:

```bash
telegraf config > telegraf.conf
```

If a user only wanted certain inputs or outputs, then the filters can be used:

```bash
telegraf config --input-filter cpu --output-filter influxdb
```
