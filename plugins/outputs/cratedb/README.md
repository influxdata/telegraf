# CrateDB Output Plugin

This plugin writes to [CrateDB](https://crate.io/) via its [PostgreSQL
protocol](https://crate.io/docs/crate/reference/protocols/postgres.html).

## Table Schema

The plugin requires a table with the following schema.

```sql
CREATE TABLE my_metrics (
  "hash_id" LONG INDEX OFF,
  "timestamp" TIMESTAMP,
  "name" STRING,
  "tags" OBJECT(DYNAMIC),
  "fields" OBJECT(DYNAMIC),
  PRIMARY KEY ("timestamp", "hash_id","day")
) PARTITIONED BY("day");
```

The plugin can create this table for you automatically via the `table_create`
config option, see below.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Configuration for CrateDB to send metrics to.
[[outputs.cratedb]]
  # A github.com/jackc/pgx/v4 connection string.
  # See https://pkg.go.dev/github.com/jackc/pgx/v4#ParseConfig
  url = "postgres://user:password@localhost/schema?sslmode=disable"
  # Timeout for all CrateDB queries.
  timeout = "5s"
  # Name of the table to store metrics in.
  table = "metrics"
  # If true, and the metrics table does not exist, create it automatically.
  table_create = true
  # The character(s) to replace any '.' in an object key with
  key_separator = "_"
```
