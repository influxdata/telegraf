# Telegraf Commands & Flags

## Usage

```bash
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
|`--watch-config`                 |Telegraf will restart on local config changes.  Monitor changes using either fs notifications or polling.  Valid values: `inotify` or `poll`.  Monitoring is off by default.|
|`--plugin-directory`             |directory containing *.so files, this directory will be searched recursively. Any Plugin found will be loaded and namespaced.|
|`--debug`                        |turn on debug logging|
|`--input-filter <filter>`        |filter the inputs to enable, separator is `:`|
|`--input-list`                   |print available input plugins.|
|`--output-filter <filter>`       |filter the outputs to enable, separator is `:`|
|`--output-list`                  |print available output plugins.|
|`--pidfile <file>`               |file to write our pid to|
|`--pprof-addr <address>`         |pprof address to listen on, don't activate pprof if empty|
|`--processor-filter <filter>`    |filter the processors to enable, separator is `:`|
|`--quiet`                        |run in quiet mode|
|`--section-filter`               |filter config sections to output, separator is `:` Valid values are `agent`, `global_tags`, `outputs`, `processors`, `aggregators` and `inputs`|
|`--sample-config`                |print out full sample configuration|
|`--once`                         |enable once mode: gather metrics once, write them, and exit|
|`--test`                         |enable test mode: gather metrics once and print them|
|`--test-wait`                    |wait up to this many seconds for service inputs to complete in test or once mode|
|`--usage <plugin>`               |print usage for a plugin, ie, `telegraf --usage mysql`|
|`--version`                      |display the version and exit|
|`--watch-interval <interval>`    |Interval to monitor http based config files ( default 0 = deactivated) it sets a background process continuously checking for new config files each `interval` duration. Server side should control if changed,a HTTP 200 (OK) response will mean there is a change since last download, a HTTP 304 (Not modified) will mean no changes. |
|`--watch-jitter <jitter>`        |time variation to ensure avoid all agents downloading the config file from the server hosting it at the same time (default 10s) (only used if --watch-interval is set) |
|`--watch-retry-interval <interval>`| time in seconds to retry download config if download failed (default 20s) |
|`--watch-max-retries <retries>`  |number of retries to download config file if previously failed (default 3) |
|`--watch-tls-cert  <path>`       |Certificate File path for TLS Config on HTTP(S) Config downloads |
|`--watch-tls-key   <path>`       |Certificate Key File path for TLS Config on HTTP(S) Config downloads |
|`--watch-tls-key-pwd <password>` |Password to decode Key file |
|`--watch-tls-ca    <path>`       |CA File path for TLS Config on HTTP(S) Config downloads |
|`--watch-tls-sni   <name>`       |SNI(Server Name Indication) indicates which hostname it is attempting to connect to at the start of the TLS handshaking process |
|`--watch-insecure-skip-verify`   |If set this flag we use TLS but skip chain & host verification (default false) |

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

**download some config files from a central server:**

```bash
telegraf --config https://myserver/telegraf_base.conf \
        --config https://myserver/telegraf_inputs.conf \
        --config https://myserver/telegraf_outputs.conf \
        --watch-interval 10m --watch-jitter 5m \
        --watch-max-retries 2 -watch-retry-interval 5s \
        --watch-insecure-skip-verify
```
