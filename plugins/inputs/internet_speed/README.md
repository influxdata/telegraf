# Internet Speed Monitor Input Plugin

The `Internet Speed Monitor` collects data about the internet speed on the
system.

On some systems, the default settings may cause speed tests to fail; if this
affects you then try enabling `memory_saving_mode`. This reduces the memory
requirements for the test, and may reduce the runtime of the test. However,
please be aware that this may also reduce the accuracy of the test for fast
(>30Mb/s) connections. This setting enables the upstream
[Memory Saving Mode](https://github.com/showwin/speedtest-go#memory-saving-mode)

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

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
```

## Metrics

It collects latency, download speed and upload speed

| Name           | filed name | type    | Unit |
| -------------- | ---------- | ------- | ---- |
| Download Speed | download   | float64 | Mbps |
| Upload Speed   | upload     | float64 | Mbps |
| Latency        | latency    | float64 | ms   |

## Example Output

```sh
internet_speed,host=Sanyam-Ubuntu download=41.791,latency=28.518,upload=59.798 1631031183000000000
```
