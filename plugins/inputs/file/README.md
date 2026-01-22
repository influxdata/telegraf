# File Input Plugin

This plugin reads the __complete__ contents of the configured files in
__every__ interval. The file content is split line-wise and parsed according to
one of the supported [data formats][data_formats].

> [!TIP]
> If you wish to only process newly appended lines use the [tail][tail] input
> plugin instead.

⭐ Telegraf v1.8.0
🏷️ system
💻 all

[data_formats]: /docs/DATA_FORMATS_INPUT.md
[tail]: /plugins/inputs/tail

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Parse a complete file each interval
[[inputs.file]]
  ## Files to parse each interval.  Accept standard unix glob matching rules,
  ## as well as ** to match recursive files and directories.
  files = ["/tmp/metrics.out"]

  ## Character encoding to use when interpreting the file contents.  Invalid
  ## characters are replaced using the unicode replacement character.  When set
  ## to the empty string the data is not decoded to text.
  ##   ex: character_encoding = "utf-8"
  ##       character_encoding = "utf-16le"
  ##       character_encoding = "utf-16be"
  ##       character_encoding = ""
  # character_encoding = ""

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Please use caution when using the following options: when file name
  ## variation is high, this can increase the cardinality significantly. Read
  ## more about cardinality here:
  ## https://docs.influxdata.com/influxdb/cloud/reference/glossary/#series-cardinality

  ## Name of tag to store the name of the file. Disabled if not set.
  # file_tag = ""

  ## Name of tag to store the absolute path and name of the file. Disabled if
  ## not set.
  # file_path_tag = ""
```

## Metrics

The format of metrics produced by this plugin depends on the content and data
format of the file.

## Example Output
