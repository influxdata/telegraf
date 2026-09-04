# Parquet Output Plugin

This plugin writes metrics to [parquet][parquet] files. By default, metrics are
grouped by metric name and written all to the same file.

> [!IMPORTANT]
> A parquet file fixes its schema when it is created. When new fields or tags
> appear, the plugin starts a new file so those columns are not dropped.

To lean more about the parquet format, check out the [parquet docs][docs] as
well as a blog post on [querying parquet][querying].

⭐ Telegraf v1.32.0
🏷️ datastore
💻 all

[parquet]: https://parquet.apache.org
[docs]: https://parquet.apache.org/docs/
[querying]: https://www.influxdata.com/blog/querying-parquet-millisecond-latency/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# A plugin that writes metrics to parquet files
[[outputs.parquet]]
  ## Directory to write parquet files in. If a file already exists the output
  ## will attempt to continue using the existing file.
  # directory = "."

  ## Files are rotated after the time interval specified. When set to 0 no time
  ## based rotation is performed. Files also rotate when new fields or tags
  ## appear so those columns are written instead of dropped.
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
not present a null value is added. If additional fields or tags appear after
the file is created, the plugin closes that file and starts a new one whose
schema includes the new columns. Earlier columns are kept so missing values
are still written as null. Rotation for a schema change is logged.

Because files in the directory can have different columns, readers should join
by column name (for example DuckDB `union_by_name = true`).

The `timestamp_field_name` column holds the metric time, so a field or tag with
that same name is dropped and logged. Set `timestamp_field_name` to another name
or rename the field or tag to keep both.

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
set. Files are also rotated when the column schema grows. Due to the usage of a
buffered writer, a size based rotation is not possible as the file may not
actually get data at each interval.

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
