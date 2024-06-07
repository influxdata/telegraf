# Parquet Parser Plugin

The Parquet parser allows for the parsing of Parquet files that were read in.

## Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "parquet"

  ## Tag column is an array of columns that should be added as tags.
  # tag_columns = []

  ## Name column is the column to use as the measurement name.
  # measurement_column = ""

  ## Timestamp column is the column containing the time that should be used to
  ## create the metric. If not set, then the time of parsing is used.
  # timestamp_column = ""

  ## Timestamp format is the time layout that should be used to interpret the
  ## timestamp_column. The time must be `unix`, `unix_ms`, `unix_us`, `unix_ns`,
  ## or a time in the "reference time".  To define a different format, arrange
  ## the values from the "reference time" in the example to match the format
  ## you will be using.  For more information on the "reference time", visit
  ## https://golang.org/pkg/time/#Time.Format
  ##   ex: timestamp_format = "Mon Jan 2 15:04:05 -0700 MST 2006"
  ##       timestamp_format = "2006-01-02T15:04:05Z07:00"
  ##       timestamp_format = "01/02/2006 15:04:05"
  ##       timestamp_format = "unix"
  ##       timestamp_format = "unix_ms"
  # timestamp_format = ""

  ## Timezone allows you to provide an override for timestamps that
  ## do not already include an offset
  ## e.g. 04/06/2016 12:41:45
  ##
  ## Default: "" which renders UTC
  ## Options are as follows:
  ##   1. Local               -- interpret based on machine localtime
  ##   2. "America/New_York"  -- Unix TZ values like those found in
  ##      https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  ##   3. UTC                 -- or blank/unspecified, will return timestamp in UTC
  # timestamp_timezone = ""
```
