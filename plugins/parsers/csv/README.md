# CSV

The `csv` parser creates metrics from a document containing comma separated
values.

### Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "csv"

  ## Indicates how many rows to treat as a header. By default, the parser assumes
  ## there is no header and will parse the first row as data. If set to anything more
  ## than 1, column names will be concatenated with the name listed in the next header row.
  ## If `csv_column_names` is specified, the column names in header will be overridden.
  csv_header_row_count = 0

  ## For assigning custom names to columns
  ## If this is specified, all columns should have a name
  ## Unnamed columns will be ignored by the parser.
  ## If `csv_header_row_count` is set to 0, this config must be used
  csv_column_names = []

  ## For assigning explicit data types to columns.
  ## Supported types: "int", "float", "bool", "string".
  ## If this is not specified, type conversion will be done on the types above.
  csv_column_types = []

  ## Indicates the number of rows to skip before looking for header information.
  csv_skip_rows = 0

  ## Indicates the number of columns to skip before looking for data to parse.
  ## These columns will be skipped in the header as well.
  csv_skip_columns = 0

  ## The seperator between csv fields
  ## By default, the parser assumes a comma (",")
  csv_delimiter = ","

  ## The character reserved for marking a row as a comment row
  ## Commented rows are skipped and not parsed
  csv_comment = ""

  ## If set to true, the parser will remove leading whitespace from fields
  ## By default, this is false
  csv_trim_space = false

  ## Columns listed here will be added as tags. Any other columns
  ## will be added as fields.
  csv_tag_columns = []

  ## The column to extract the name of the metric from
  csv_measurement_column = ""

  ## The column to extract time information for the metric
  ## `csv_timestamp_format` must be specified if this is used
  csv_timestamp_column = ""

  ## The format of time data extracted from `csv_timestamp_column`
  ## this must be specified if `csv_timestamp_column` is specified
  csv_timestamp_format = ""
  ```
#### csv_timestamp_column, csv_timestamp_format

By default the current time will be used for all created metrics, to set the
time using the JSON document you can use the `csv_timestamp_column` and
`csv_timestamp_format` options together to set the time to a value in the parsed
document.

The `csv_timestamp_column` option specifies the key containing the time value and
`csv_timestamp_format` must be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`,
or a format string in using the Go "reference time" which is defined to be the
**specific time**: `Mon Jan 2 15:04:05 MST 2006`.

Consult the Go [time][time parse] package for details and additional examples
on how to set the time format.

### Metrics

One metric is created for each row with the columns added as fields.  The type
of the field is automatically determined based on the contents of the value.

### Examples

Config:
```
[[inputs.file]]
  files = ["example"]
  data_format = "csv"
  csv_header_row_count = 1
  csv_timestamp_column = "time"
  csv_timestamp_format = "2006-01-02T15:04:05Z07:00"
```

Input:
```
measurement,cpu,time_user,time_system,time_idle,time
cpu,cpu0,42,42,42,2018-09-13T13:03:28Z
```

Output:
```
cpu cpu=cpu0,time_user=42,time_system=42,time_idle=42 1536869008000000000
```
