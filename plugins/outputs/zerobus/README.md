# Zerobus Output Plugin

This plugin writes metrics to a Unity Catalog Delta table using the
[Databricks Zerobus Ingest][zerobus] service, which accepts records over gRPC
and commits them into Delta directly. Tags and fields are mapped onto the
columns of the destination table.

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

  ## OAuth service-principal credentials.
  client_id = ""
  client_secret = ""

  ## Timestamp column. Set to "" to turn off.
  # timestamp_column = "timestamp"

  ## Column receiving the measurement name. The name is omitted if empty.
  # measurement_column = ""

  ## Optional identifier appended to Telegraf's product token.
  # application_name = ""

  ## Stream startup timeout.
  # connect_timeout = "30s"
```

The service principal identified by `client_id` needs the `USE CATALOG`,
`USE SCHEMA`, `SELECT` and `MODIFY` [privileges][privileges] on the destination
table. Startup waits up to `connect_timeout` for the stream, so network,
authentication and permission errors surface before any metric is written.

[privileges]: https://docs.databricks.com/aws/en/data-governance/unity-catalog/manage-privileges/privileges

## Metric mapping

The plugin writes one flat row per metric, taking the column layout from the
destination table:

- The metric timestamp is written to `timestamp_column` as Unix microseconds,
  which is the representation expected by a Delta `TIMESTAMP`. The default
  column name is `timestamp`. Set `timestamp_column = ""` to omit it.
- Tags and fields become same-named columns.
- The measurement name is omitted unless `measurement_column` is configured.

For example, a metric with the tag `host` and field `usage` can target:

```sql
CREATE TABLE catalog.schema.cpu_metrics (
  timestamp TIMESTAMP NOT NULL,
  host STRING,
  usage DOUBLE
);
```

Declare columns as `BIGINT` for `int64` and `uint64`, `DOUBLE` for `float64`,
`BOOLEAN` for `bool`, and `STRING` for `string`. Two value ranges cannot be
written: `uint64` values above `math.MaxInt64`, because Delta has no unsigned
64-bit type, and non-finite `float64` values such as `NaN`, because JSON cannot
represent them.

All metrics of a plugin instance target the configured table and must match its
schema. Use Telegraf filtering or processors when separate measurements need
different tables or column layouts. The schema is read from Unity Catalog when
the stream is opened, so altered columns are picked up by the next stream
without restarting Telegraf.

## Batching and durability

Each batch is split into requests that stay within the Zerobus size limits and
the write succeeds only once Databricks acknowledged every record. A metric that
cannot be encoded, for example because a tag and a field share a name, is
rejected on its own so the rest of the batch is still written.

A failing write returns an error and Telegraf retries the buffered batch on a
new stream. Records the failed attempt already got acknowledged can therefore be
written twice, so treat the destination table as at-least-once and deduplicate
in Delta when exactly-once is required.
