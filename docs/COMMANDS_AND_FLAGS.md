# Telegraf Commands & Flags

## Usage

```shell
telegraf [commands]
telegraf [flags]
```

## Commands

|command|description|
|--------|-----------------------------------------------|
|`config` |print out full sample configuration to stdout|
|`version`|print the version to stdout|

## Flags

|flag|description|
|-------------------|------------|
|`--aggregator-filter <filter>`   |filter the aggregators to enable, separator is `:`|
|`--config <file>`                |configuration file to load|
|`--config-directory <directory>` |directory containing additional *.conf files|
|`--watch-config`                 |Telegraf will restart on local config changes. Monitor changes using either fs notifications or polling. Valid values: `inotify` or `poll`. Monitoring is off by default.|
|`--plugin-directory`             |directory containing *.so files, this directory will be searched recursively. Any Plugin found will be loaded and namespaced.|
|`--debug`                        |turn on debug logging|
|`--deprecation-list`             |print all deprecated plugins or plugin options|
|`--input-filter <filter>`        |filter the inputs to enable, separator is `:`|
|`--input-list`                   |print available input plugins.|
|`--output-filter <filter>`       |filter the outputs to enable, separator is `:`|
|`--output-list`                  |print available output plugins.|
|`--pidfile <file>`               |file to write our pid to|
|`--pprof-addr <address>`         |pprof address to listen on, don't activate pprof if empty|
|`--processor-filter <filter>`    |filter the processors to enable, separator is `:`|
|`--quiet`                        |run in quiet mode|
|`--section-filter`               |filter config sections to output, separator is `:`.  Valid values are `agent`, `global_tags`, `outputs`, `processors`, `aggregators` and `inputs`|
|`--sample-config`                |print out full sample configuration|
|`--once`                         |enable once mode: gather metrics once, write them, and exit|
|`--test`                         |enable test mode: gather metrics once and print them. **No outputs are executed!**|
|`--test-wait`                    |wait up to this many seconds for service inputs to complete in test or once mode.  **Implies `--test` if not used with `--once`**|
|`--usage <plugin>`               |print usage for a plugin, ie, `telegraf --usage mysql`|
|`--version`                      |display the version and exit|

## Examples

**Generate a telegraf config file:**

`telegraf config > telegraf.conf`

**Generate config with only cpu input & influxdb output plugins defined:**

`telegraf --input-filter cpu --output-filter influxdb config`

**Run a single telegraf collection, outputting metrics to stdout:**

`telegraf --config telegraf.conf --test`

**Run telegraf with all plugins defined in config file:**

`telegraf --config telegraf.conf`

**Run telegraf, enabling the cpu & memory input, and influxdb output plugins:**

`telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb`

**Run telegraf with pprof:**

`telegraf --config telegraf.conf --pprof-addr localhost:6060`
