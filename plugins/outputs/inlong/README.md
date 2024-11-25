# Inlong Output Plugin

This plugin writes telegraf metrics to Inlong

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send telegraf metrics to Inlong
[[outputs.inlong]]
  ## From the Inlong system, data streams group, it contains multiple data streams, and one Group represents
  ## one data business unit.
  group_id = "test_group"

  ## From the Inlong system, data stream, a stream has a specific data source, data format and data sink.
  stream_id = "test_stream"

  ## The URL used to obtain the Inlong DataProxy IP list to which the data will be sent
  manager_url = "http://127.0.0.1:8083/inlong/manager/openapi/dataproxy/getIpList"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  ## Suggest using CSV format here, as inlong is also processed in CSV format
  data_format = "csv"

  ## The delimiter used when serializing data in CSV format needs to be consistent with the delimiter
  ## configured for inlong, so that the data can be parsed properly after it reaches inlong
  csv_separator = "|"

  ## The final output field order here needs to be consistent with the field order defined by the data
  ## stream in inlong
  csv_columns = ["field.key","file.value"]
```
