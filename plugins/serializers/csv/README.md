# CSV Serializer

The `csv` output data format converts metrics into CSV lines.

## Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "csv"

  ## The default timestamp format is Unix epoch time.
  # Other timestamp layout can be configured using the Go language time
  # layout specification from https://golang.org/pkg/time/#Time.Format
  # e.g.: csv_timestamp_format = "2006-01-02T15:04:05Z07:00"
  # csv_timestamp_format = "unix"

  ## The default separator for the CSV format.
  # csv_separator = ","

  ## Output the CSV header in the first line.
  ## Enable the header when outputting metrics to a new file.
  ## Disable when appending to a file or when using a stateless
  ## output to prevent headers appearing between data lines.
  # csv_header = false

  ## Prefix tag and field columns with "tag_" and "field_" respectively.
  ## This can be helpful if you need to know the "type" of a column.
  # csv_column_prefix = false
```

## Examples

Standard form:

```csv
1458229140,docker,raynor,30,4,...,59,660
```

When an output plugin needs to emit multiple metrics at one time, it may use
the batch format. The use of batch format is determined by the plugin,
reference the documentation for the specific plugin. With `csv_header = true`
you get

```csv
timestamp,measurement,host,field_1,field_2,...,field_N,n_images
1458229140,docker,raynor,30,4,...,59,660
1458229143,docker,raynor,28,5,...,60,665
```
