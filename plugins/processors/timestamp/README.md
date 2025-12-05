# Timestamp Processor Plugin

This plugin allows to parse fields containing timestamps into timestamps of
other format.

⭐ Telegraf v1.31.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Convert a timestamp field to other timestamp format
[[processors.timestamp]]
  ## Timestamp key to convert
  ## Specify the field name that contains the timestamp to convert. The result
  ## will replace the current field value.
  field = ""

  ## Timestamp Format
  ## This defines the time layout used to interpret the source timestamp field.
  ## The time must be `unix`, `unix_ms`, `unix_us`, `unix_ns`, or a time in Go
  ## "reference time". For more information on Go "reference time". For more
  ## see: https://golang.org/pkg/time/#Time.Format
  source_timestamp_format = ""

  ## Timestamp Timezone
  ## Source timestamp timezone. If not set, assumed to be in UTC.
  ## Options are as follows:
  ##   1. UTC                 -- or unspecified will return timestamp in UTC
  ##   2. Local               -- interpret based on machine localtime
  ##   3. "America/New_York"  -- Unix TZ values like those found in
  ##        https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  # source_timestamp_timezone = ""

  ## Target timestamp format
  ## This defines the destination timestamp format. It also can accept either
  ## `unix`, `unix_ms`, `unix_us`, `unix_ns`, or a time in Go "reference time".
  destination_timestamp_format = ""

  ## Target Timestamp Timezone
  ## Source timestamp timezone. If not set, assumed to be in UTC.
  ## Options are as follows:
  ##   1. UTC                 -- or unspecified will return timestamp in UTC
  ##   2. Local               -- interpret based on machine localtime
  ##   3. "America/New_York"  -- Unix TZ values like those found in
  ##        https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  # destination_timestamp_timezone = ""
```

## Example

Convert a timestamp to unix timestamp:

```toml
[[processors.timestamp]]
  source_timestamp_field = "timestamp"
  source_timestamp_format = "2006-01-02T15:04:05.999999999Z"
  destination_timestamp_format = "unix"
```

```diff
- metric value=42i,timestamp="2024-03-04T10:10:32.123456Z" 1560540094000000000
+ metric value=42i,timestamp=1709547032 1560540094000000000
```

Convert the same timestamp to a nanosecond unix timestamp:

```toml
[[processors.timestamp]]
  source_timestamp_field = "timestamp"
  source_timestamp_format = "2006-01-02T15:04:05.999999999Z"
  destination_timestamp_format = "unix_ns"
```

```diff
- metric value=42i,timestamp="2024-03-04T10:10:32.123456789Z" 1560540094000000000
+ metric value=42i,timestamp=1709547032123456789 1560540094000000000
```

Convert the timestamp to another timestamp format:

```toml
[[processors.timestamp]]
  source_timestamp_field = "timestamp"
  source_timestamp_format = "2006-01-02T15:04:05.999999999Z"
  destination_timestamp_format = "2006-01-02T15:04"
```

```diff
- metric value=42i,timestamp="2024-03-04T10:10:32.123456Z" 1560540094000000000
+ metric value=42i,timestamp="2024-03-04T10:10" 1560540094000000000
```
