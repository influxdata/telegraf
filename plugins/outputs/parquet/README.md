# Parquet Output Plugin

This plugin writes metrics to [parquet][parquet] files. By default, metrics are
grouped by metric name and written all to the same file.

> [!IMPORTANT]
> If a metric schema does not match the schema in the file it will be dropped.

To lean more about the parquet format, check out the [parquet docs][docs] as
well as a blog post on [querying parquet][querying].

⭐ Telegraf v1.32.0
🏷️ datastore
💻 all

[parquet]: https://parquet.apache.org
[docs]: https://parquet.apache.org/docs/
[querying]: https://www.influxdata.com/blog/querying-parquet-millisecond-latency/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# A plugin that writes metrics to parquet files
[[outputs.parquet]]
  ## Directory to write parquet files in. If a file already exists the output
  ## will attempt to continue using the existing file.
  # directory = "."

  ## Files are rotated after the time interval specified. When set to 0 no time
  ## based rotation is performed.
  # rotation_interval = "0h"

  ## Timestamp field name
  ## Field name to use to store the timestamp. If set to an empty string, then
  ## the timestamp is omitted.
  # timestamp_field_name = "timestamp"
```

## Building Parquet Files

### Schema

Parquet files require a schema when writing files. To generate a schema,
Telegraf will go through all grouped metrics and generate an Apache Arrow schema
based on the union of all fields and tags. If a field and tag have the same name
then the field takes precedence.

The consequence of schema generation is that the very first flush sequence a
metric is seen takes much longer due to the additional looping through the
metrics to generate the schema. Subsequent flush intervals are significantly
faster.

When writing to a file, the schema is used to look for each value and if it is
not present a null value is added. The result is that if additional fields are
present after the first metric flush those fields are omitted.

### Write

The plugin makes use of the buffered writer. This may buffer some metrics into
memory before writing it to disk. This method is used as it can more compactly
write multiple flushes of metrics into a single Parquet row group.

Additionally, the Parquet format requires a proper footer, so close must be
called on the file to ensure it is properly formatted.

### Close

Parquet files must close properly or the file will not be readable. The parquet
format requires a footer at the end of the file and if that footer is not
present then the file cannot be read correctly.

If Telegraf were to crash while writing parquet files there is the possibility
of this occurring.

## File Rotation

If a file with the same target name exists at start, the existing file is
rotated to avoid over-writing it or conflicting schema.

File rotation is available via a time based interval that a user can optionally
set. Due to the usage of a buffered writer, a size based rotation is not
possible as the file may not actually get data at each interval.

## Explore Parquet Files

If a user wishes to explore a schema or data in a Parquet file quickly, then
consider the options below:

### CLI

The Arrow repo contains a Go CLI tool to read and parse Parquet files:

```s
go install github.com/apache/arrow-go/v18/parquet/cmd/parquet_reader@latest
parquet_reader <file>
```

### Python

Users can also use the [pyarrow][] library to quick open and explore Parquet
files:

```python
import pyarrow.parquet as pq

table = pq.read_table('example.parquet')
```

Once created, a user can look the various [pyarrow.Table][] functions to further
explore the data.

[pyarrow]: https://arrow.apache.org/docs/python/generated/pyarrow.parquet.read_table.html
[pyarrow.Table]: https://arrow.apache.org/docs/python/generated/pyarrow.Table.html#pyarrow.Table
