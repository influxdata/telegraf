# Internet Speed Monitor Input Plugin

This plugin collects metrics about the internet speed on the system like
download/upload speed, latency etc using the [speedtest.net service][speedtest].

⭐ Telegraf v1.20.0
🏷️ network
💻 all

[speedtest]: https://www.speedtest.net/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Monitors internet speed using speedtest.net service
[[inputs.internet_speed]]
  ## This plugin downloads many MB of data each time it is run. As such
  ## consider setting a higher interval for this plugin to reduce the
  ## demand on your internet connection.
  # interval = "60m"

  ## Enable to reduce memory usage
  # memory_saving_mode = false

  ## Caches the closest server location
  # cache = false

  ## Number of concurrent connections
  ## By default or set to zero, the number of CPU cores is used. Use this to
  ## reduce the impact on system performance or to increase the connections on
  ## faster connections to ensure the fastest speed.
  # connections = 0

  ## Test mode
  ## By default, a single sever is used for testing. This may work for most,
  ## however, setting to "multi" will reach out to multiple servers in an
  ## attempt to get closer to ideal internet speeds.
  ## And "multi" will use all available servers to calculate average packet loss.
  # test_mode = "single"

  ## Server ID exclude filter
  ## Allows the user to exclude or include specific server IDs received by
  ## speedtest-go. Values in the exclude option will be skipped over. Values in
  ## the include option are the only options that will be picked from.
  ##
  ## See the list of servers speedtest-go will return at:
  ##     https://www.speedtest.net/api/js/servers?engine=js&limit=10
  ##
  # server_id_exclude = []
  # server_id_include = []
```

> [!TIP]
> On some systems, the default settings may cause speed tests to fail. If this
> affects your system, try enabling `memory_saving_mode`, which reduces the
> memory requirements and the runtime of the test at the cost of a reduced
> accuracy especially for fast (>30Mb/s) connections. Check the
> [upstream documentation][docs] for details

[docs]: https://github.com/showwin/speedtest-go#memory-saving-mode

## Metrics

It collects the following fields:

| Name           | Field Name  | Type    | Unit       |
|----------------|-------------|---------|------------|
| Download Speed | download    | float64 | Mbps       |
| Upload Speed   | upload      | float64 | Mbps       |
| Latency        | latency     | float64 | ms         |
| Jitter         | jitter      | float64 | ms         |
| Packet Loss    | packet_loss | float64 | percentage |
| Location       | location    | string  | -          |

The `packet_loss` will return -1, if packet loss not applicable.

And the following tags:

| Name      | tag name  |
|-----------|-----------|
| Source    | source    |
| Server ID | server_id |
| Test Mode | test_mode |

## Example Output

```text
internet_speed,source=speedtest02.z4internet.com:8080,server_id=54619,test_mode=single download=318.37580265897725,upload=30.444407341274385,latency=37.73174,jitter=1.99810,packet_loss=0.05377,location="Somewhere, TX" 1675458921000000000
internet_speed,source=speedtest02.z4internet.com:8080,server_id=54619,test_mode=multi download=318.37580265897725,upload=30.444407341274385,latency=37.73174,jitter=1.99810,packet_loss=-1,location="Somewhere, TX" 1675458921000000000
```
