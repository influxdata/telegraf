# CrateDB Output Plugin

This plugin writes metrics to [CrateDB][cratedb] via its
[PostgreSQL protocol][psql_protocol].

⭐ Telegraf v1.5.0
🏷️ cloud, datastore
💻 all

[cratedb]: https://crate.io/
[psql_protocol]: https://crate.io/docs/crate/reference/protocols/postgres.html

## Table Schema

The plugin requires a table with the following schema.

```sql
CREATE TABLE IF NOT EXISTS my_metrics (
  "hash_id" LONG INDEX OFF,
  "timestamp" TIMESTAMP,
  "name" STRING,
  "tags" OBJECT(DYNAMIC),
  "fields" OBJECT(DYNAMIC),
  "day" TIMESTAMP GENERATED ALWAYS AS date_trunc('day', "timestamp"),
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

## Configuration

```toml @sample.conf
# Configuration for CrateDB to send metrics to.
[[outputs.cratedb]]
  ## Connection parameters for accessing the database see
  ##   https://pkg.go.dev/github.com/jackc/pgx/v4#ParseConfig
  ## for available options
  url = "postgres://user:password@localhost/schema?sslmode=disable"

  ## Timeout for all CrateDB queries.
  # timeout = "5s"

  ## Name of the table to store metrics in.
  # table = "metrics"

  ## If true, and the metrics table does not exist, create it automatically.
  # table_create = false

  ## The character(s) to replace any '.' in an object key with
  # key_separator = "_"
```
