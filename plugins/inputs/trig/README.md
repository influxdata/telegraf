# Trig Input Plugin

The `trig` plugin is for demonstration purposes and inserts sine and cosine

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Inserts sine and cosine waves for demonstration purposes
[[inputs.trig]]
  ## Set the amplitude
  amplitude = 10.0
```

## Metrics

- trig
  - fields:
    - cosine (float)
    - sine (float)

## Example Output

```shell
trig,host=MBP15-SWANG.local cosine=10,sine=0 1632338680000000000
trig,host=MBP15-SWANG.local sine=5.877852522924732,cosine=8.090169943749473 1632338690000000000
trig,host=MBP15-SWANG.local sine=9.510565162951535,cosine=3.0901699437494745 1632338700000000000
```
