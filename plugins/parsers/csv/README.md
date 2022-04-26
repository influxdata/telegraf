# CSV

The `csv` parser creates metrics from a document containing comma separated
values.

## Configuration

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
  ## Specify types in order by column (e.g. `["string", "int", "float"]`)
  ## If this is not specified, type conversion will be done on the types above.
  csv_column_types = []

  ## Indicates the number of rows to skip before looking for metadata and header information.
  csv_skip_rows = 0
  
  ## Indicates the number of rows to parse as metadata before looking for header information. 
  ## By default, the parser assumes there are no metadata rows to parse. 
  ## If set, the parser would use the provided separators in the csv_metadata_separators to look for metadata.
  ## Please note that by default, the (key, value) pairs will be added as tags. 
  ## If fields are required, use the converter processor.
  csv_metadata_rows = 0
  
  ## A list of metadata separators. If csv_metadata_rows is set,
  ## csv_metadata_separators must contain at least one separator.
  ## Please note that separators are case sensitive and the sequence of the seperators are respected.
  csv_metadata_separators = [":", "="]
  
  ## A set of metadata trim characters. 
  ## If csv_metadata_trim_set is not set, no trimming is performed.
  ## Please note that the trim cutset is case sensitive.
  csv_metadata_trim_set = ""

  ## Indicates the number of columns to skip before looking for data to parse.
  ## These columns will be skipped in the header as well.
  csv_skip_columns = 0

  ## The separator between csv fields
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

  ## The column to extract the name of the metric from. Will not be
  ## included as field in metric.
  csv_measurement_column = ""

  ## The column to extract time information for the metric
  ## `csv_timestamp_format` must be specified if this is used.
  ## Will not be included as field in metric.
  csv_timestamp_column = ""

  ## The format of time data extracted from `csv_timestamp_column`
  ## this must be specified if `csv_timestamp_column` is specified
  csv_timestamp_format = ""

  ## The timezone of time data extracted from `csv_timestamp_column`
  ## in case of there is no timezone information.
  ## It follows the  IANA Time Zone database.
  csv_timezone = ""

  ## Indicates values to skip, such as an empty string value "".
  ## The field will be skipped entirely where it matches any values inserted here.
  csv_skip_values = []

  ## If set to true, the parser will skip csv lines that cannot be parsed.
  ## By default, this is false
  csv_skip_errors = false
  ```

### csv_timestamp_column, csv_timestamp_format

By default, the current time will be used for all created metrics, to set the
time using the JSON document you can use the `csv_timestamp_column` and
`csv_timestamp_format` options together to set the time to a value in the parsed
document.

The `csv_timestamp_column` option specifies the key containing the time value and
`csv_timestamp_format` must be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`,
or a format string in using the Go "reference time" which is defined to be the
**specific time**: `Mon Jan 2 15:04:05 MST 2006`.

Consult the Go [time][time parse] package for details and additional examples
on how to set the time format.

## Metrics

One metric is created for each row with the columns added as fields.  The type
of the field is automatically determined based on the contents of the value.

In addition to the options above, you can use [metric filtering][] to skip over
columns and rows.

## Examples

Config:

```toml
[[inputs.file]]
  files = ["example"]
  data_format = "csv"
  csv_header_row_count = 1
  csv_timestamp_column = "time"
  csv_timestamp_format = "2006-01-02T15:04:05Z07:00"
```

Input:

```csv
measurement,cpu,time_user,time_system,time_idle,time
cpu,cpu0,42,42,42,2018-09-13T13:03:28Z
```

Output:

```text
cpu cpu=cpu0,time_user=42,time_system=42,time_idle=42 1536869008000000000
```

Config:

```toml
[[inputs.file]]
  files = ["example"]
  data_format = "csv"
  csv_metadata_rows = 2
  csv_metadata_separators = [":", "="]
  csv_metadata_trim_set = " #"
  csv_header_row_count = 1
  csv_tag_columns = ["Version","File Created"]
  csv_timestamp_column = "time"
  csv_timestamp_format = "2006-01-02T15:04:05Z07:00"
```

Input:

```csv
# Version=1.1
# File Created: 2021-11-17T07:02:45+10:00
measurement,cpu,time_user,time_system,time_idle,time
cpu,cpu0,42,42,42,2018-09-13T13:03:28Z
```

Output:

```text
cpu,File\ Created=2021-11-17T07:02:45+10:00,Version=1.1 cpu=cpu0,time_user=42,time_system=42,time_idle=42 1536869008000000000
```

[metric filtering]: /docs/CONFIGURATION.md#metric-filtering
