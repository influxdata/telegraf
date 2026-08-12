# Zerobus Output Plugin

This plugin writes metrics to a Unity Catalog Delta table using the
[Databricks Zerobus Ingest][zerobus] service. It supports a static schema that
stores arbitrary metrics in a fixed envelope and an opt-in table-schema mode
that maps metrics onto the columns of the destination table.

> [!IMPORTANT]
> Be aware that this plugin accesses APIs that are [chargeable][pricing] and
> might incur costs.

⭐ Telegraf v1.40.0
🏷️ cloud, datastore
💻 all

[pricing]: https://www.databricks.com/product/pricing/lakeflow-connect
[zerobus]: https://docs.databricks.com/aws/en/ingestion/zerobus-ingest

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Startup error behavior options <!-- @/docs/includes/startup_error_behavior.md -->

In addition to the plugin-specific and global configuration settings the plugin
supports options for specifying the behavior when experiencing startup errors
using the `startup_error_behavior` setting. Available values are:

- `error`:  Telegraf with stop and exit in case of startup errors. This is the
            default behavior.
- `ignore`: Telegraf will ignore startup errors for this plugin and disables it
            but continues processing for all other plugins.
- `retry`:  Telegraf will try to startup the plugin in every gather or write
            cycle in case of startup errors. The plugin is disabled until
            the startup succeeds.
- `probe`:  Telegraf will probe the plugin's function (if possible) and disables
            the plugin in case probing fails. If the plugin does not support
            probing, Telegraf will behave as if `ignore` was set instead.

## Secret store support

This plugin supports secrets from secret stores for the `client_secret` option.
See the [secret store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Configuration for sending metrics to Databricks Zerobus
[[outputs.zerobus]]
  ## Zerobus gRPC service endpoint.
  server_endpoint = "https://<workspace-id>.zerobus.<region>.cloud.databricks.com"

  ## Databricks workspace URL used for OAuth authentication.
  workspace_url = "https://<workspace>.cloud.databricks.com"

  ## Fully qualified Unity Catalog destination table.
  table_name = "catalog.schema.telegraf_metrics"

  ## Schema mode: static stores fields in a VARIANT column; table_schema maps
  ## tags and fields to columns from the destination table schema.
  # schema_mode = "static"

  ## Optional timestamp column for table_schema mode. Leave empty if the table
  ## has no timestamp column.
  # timestamp_column = "timestamp"

  ## Optional measurement-name column for table_schema mode.
  # measurement_column = ""

  ## OAuth service-principal credentials.
  client_id = ""
  client_secret = ""

  ## Optional identifier appended to Telegraf's product token.
  # application_name = ""

  ## Stream startup timeout.
  # connect_timeout = "30s"

  ## Number of streams each batch is spread over (maximum 100).
  ## NOTE: Ordering is only guaranteed per stream, so sort on the metric
  ##       timestamp when using more than one.
  # concurrent_streams = 1

  ## Number of times a broken stream is recovered before the write fails.
  # recovery_retries = 4

  ## Time to wait for Databricks to acknowledge a request.
  # lack_of_ack_timeout = "60s"

  ## Time to wait for a batch to be written completely.
  # flush_timeout = "5m"
```

The service principal identified by `client_id` needs the `USE CATALOG`,
`USE SCHEMA`, `SELECT` and `MODIFY` [privileges][privileges] on the destination
table. Startup waits up to `connect_timeout` for the stream, so network,
authentication and permission errors surface before any metric is written.

[privileges]: https://docs.databricks.com/aws/en/data-governance/unity-catalog/manage-privileges/privileges

## Schema modes

### Static schema

Static mode is the default. It writes one row per metric, with the metric fields
in a `VARIANT` column. Create the destination table with this exact schema and
column order:

```sql
CREATE TABLE catalog.schema.telegraf_metrics (
  measurement STRING NOT NULL,
  timestamp_ns BIGINT NOT NULL,
  tags MAP<STRING, STRING> NOT NULL,
  fields VARIANT NOT NULL
);
```

The table has to match this definition one-to-one, so do not reorder, rename,
remove, or change the nullability of these columns. Later revisions of the
plugin will only ever append nullable columns to it.

`timestamp_ns` is a raw Unix nanosecond `BIGINT`, not a Delta `TIMESTAMP`.

### Field mapping

All fields of a metric become one JSON object in the `fields` `VARIANT` column,
so field names and types can change without altering the destination table.
`int64`, `bool` and `string` values are stored as they are, while two cases are
rejected: `uint64` values above `math.MaxInt64`, because Delta has no unsigned
64-bit type, and non-finite `float64` values such as `NaN`, because JSON cannot
represent them.

Since the values are stored as JSON, the type of a field follows the value
rather than its Telegraf type. A `float64` holding a whole number is
indistinguishable from an `int64`, so the same field can come back as a
`BIGINT` in one row and a decimal in another. Select a field with `:` and cast
it with `::` to pin the type, which Databricks requires before values can be
filtered, grouped, ordered or aggregated:

```sql
SELECT measurement, tags['host'] AS host, fields:usage_idle::double AS idle
FROM catalog.schema.telegraf_metrics;
```

See the [VARIANT documentation][variant] for reading values that may not
convert.

[variant]: https://docs.databricks.com/aws/en/semi-structured/variant

### Table schema

Set `schema_mode = "table_schema"` to take the record layout from the columns of
the destination table instead of the fixed schema above. The schema is read from
Unity Catalog at startup, and if the table is altered later the plugin picks up
the new columns on its own, without a Telegraf restart.

Table-schema mode creates one flat record per metric:

- If configured, the metric timestamp is written to `timestamp_column` as Unix
  microseconds, which is the representation expected by a Delta `TIMESTAMP`.
- Tags and fields become same-named top-level columns.
- The measurement name is omitted unless `measurement_column` is configured.

For example, a metric with the tag `host` and field `usage` can target:

```sql
CREATE TABLE catalog.schema.cpu_metrics (
  timestamp TIMESTAMP NOT NULL,
  host STRING,
  usage DOUBLE
);
```

All metrics from one plugin instance target the configured table and must match
its schema. Use Telegraf filtering or processors when separate measurements need
different tables or column layouts.

Declare columns as `BIGINT` for `int64` and `uint64`, `DOUBLE` for `float64`,
`BOOLEAN` for `bool`, and `STRING` for `string`; the value limits from
[Field mapping](#field-mapping) apply here too. A table with a nullable array or
map, or one whose collections allow null elements, cannot be used at all, and an
individual metric is rejected when a tag or field name collides with another
column.

## Batching and durability

Every batch is split into requests that stay within the Zerobus size limits and
sent together. The plugin reports success only once Databricks has acknowledged
each record.

A retry resumes where the previous attempt stopped rather than re-sending
acknowledged records. A metric that cannot be encoded or is too large to send is
rejected on its own, so the rest of the batch is still written.

## Concurrent streams

Raise the agent's `metric_batch_size` before adding streams. Each write waits a
fixed amount of time for Databricks to acknowledge it, which extra streams
cannot shorten, so they only pay off once a batch is large enough that the
per-record work outweighs that wait. See the [Zerobus quotas][quotas] for what a
single stream can sustain.

Each batch is divided into one contiguous share per stream, sent in parallel. If
one stream fails, Telegraf keeps the whole batch and the retry resumes only the
streams that did not finish.

[quotas]: https://docs.databricks.com/aws/en/ingestion/zerobus-quotas
