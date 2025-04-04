# Multifile Input Plugin

This plugin reads the combined data from multiple files into a single metric,
creating one field or tag per file.  This is often useful creating custom
metrics from the `/sys` or `/proc` filesystems.

> [!NOTE]
> To parse metrics from a single file you should use the [file][file_plugin]
> input plugin instead.

⭐ Telegraf v1.10.0
🏷️ system
💻 all

[file_plugin]: /plugins/inputs/file/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Aggregates the contents of multiple files into a single point
[[inputs.multifile]]
  ## Base directory where telegraf will look for files.
  ## Omit this option to use absolute paths.
  base_dir = "/sys/bus/i2c/devices/1-0076/iio:device0"

  ## If true discard all data when a single file can't be read.
  ## Else, Telegraf omits the field generated from this file.
  # fail_early = true

  ## Files to parse each interval.
  [[inputs.multifile.file]]
    file = "in_pressure_input"
    dest = "pressure"
    conversion = "float"
  [[inputs.multifile.file]]
    file = "in_temp_input"
    dest = "temperature"
    conversion = "float(3)"
  [[inputs.multifile.file]]
    file = "in_humidityrelative_input"
    dest = "humidityrelative"
    conversion = "float(3)"
```

## Metrics

Each file table can contain the following options:

* `file`:
Path of the file to be parsed, relative to the `base_dir`.
* `dest`:
Name of the field/tag key, defaults to `$(basename file)`.
* `conversion`:
Data format used to parse the file contents:
  * `float(X)`: Converts the input value into a float and divides by the Xth
    power of 10. Effectively just moves the decimal left X places. For example
    a value of `123` with `float(2)` will result in `1.23`.
  * `float`: Converts the value into a float with no adjustment.
    Same as `float(0)`.
  * `int`: Converts the value into an integer.
  * `string`, `""`: No conversion.
  * `bool`: Converts the value into a boolean.
  * `tag`: File content is used as a tag.

## Example Output

This example shows a BME280 connected to a Raspberry Pi, using the sample
config.

```text
multifile pressure=101.343285156,temperature=20.4,humidityrelative=48.9 1547202076000000000
```

To reproduce this, connect a BMP280 to the board's GPIO pins and register the
BME280 device driver

```sh
cd /sys/bus/i2c/devices/i2c-1
echo bme280 0x76 > new_device
```

The kernel driver provides the following files in
`/sys/bus/i2c/devices/1-0076/iio:device0`:

* `in_humidityrelative_input`: `48900`
* `in_pressure_input`: `101.343285156`
* `in_temp_input`: `20400`
