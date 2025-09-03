# KaiwuDBLite Output Plugin

Introduced in Telegraf v1.36.0

This plugin allows Telegraf to write collected metrics
to a KaiwuDBLite instance.
It supports both compact and expanded schema modes, providing flexible options  
for storing tags and fields according to your needs.

KaiwuDBLite is a lightweight, single-node version of [KaiwuDB][KaiwuDB],
designed for edge computing and IoT scenarios.
It provides high-performance metric storage in low-resource environments,
capable of handling millions of data points per second and responding to
tens of millions of records within milliseconds.

⭐ Telegraf v1.36.0
🏷️ datastore
💻 all

[KaiwuDB]:
 https://www.kaiwudb.com/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Getting started

To use the plugin:

1. Configure the data source name (DSN). The DSN format includes host, port,
   and user.  
   - Default port: 36257  
   - Example:

     ```toml
     data_source_name = "host=127.0.0.1 port=36257 user=admin"
     ```

2. Ensure the database user has permission to create and alter tables.

## Schema Modes

### Compact schema (default)

If `enable_compact_schema` = true, each metric table contains three columns:

```sql
CREATE TABLE <metric>(
  ts TIMESTAMP,
  tags MAP(VARCHAR, VARCHAR),
  fields JSON
);
```

### Expanded schema

If `enable_compact_schema` = false, the plugin creates a full column for each
tag and field.

## Configuration

```toml @sample.conf
# Save metrics to kaiwudb-lite
[[outputs.kaiwudb]]
  ## Database driver
  ## Valid options: kaiwudb, kaiwudb-lite
  ## Kaiwudb is currently not supported, only Kaiwudb-lite
  driver = "kaiwudb-lite"

  ## Data source name, compatible with psql
  ## By default, port=36257 and user=admin
  data_source_name = "host=127.0.0.1 port=36257 user=admin connect_timeout=5"

  ## Timestamp with time zone, default false
  ## If true, the timestamp will be stored with time zone information.
  ## Storage still uses UTC format, and only displays timestamps with time zone information when querying based on the set timezone.
  # timestamp_with_time_zone = false

  ## Timestamp column name, default "ts"
  # timestamp_column_name    = "ts"

  ## The table structure only contains three fields, such as:
  ## CREATE TABLE test(ts TIMESTAMP, tags MAP(VARCHAR, VARCHAR), fields JSON);
  # tags_column_name   = "tags"
  # fileds_column_name = "fields"

  ## Initialization SQL
  # init_sql = ""

  ## Maximum amount of time a connection may be idle. "0s" means connections are
  ## never closed due to idle time.
  # connection_max_idle_time = "0s"

  ## Maximum amount of time a connection may be reused. "0s" means connections
  ## are never closed due to age.
  # connection_max_lifetime = "0s"

  ## Maximum number of connections in the idle connection pool. 0 means unlimited.
  # connection_max_idle = 2

  ## Maximum number of open connections to the database. 0 means unlimited.
  # connection_max_open = 0

  ## NOTE: Due to the way TOML is parsed, tables must be at the END of the
  ## plugin definition, otherwise additional config options are read as part of
  ## the table

  ## Metric type to SQL type conversion
  ## The values on the left are the data types Telegraf has and the values on
  ## the right are the data types Telegraf will use when sending to a database.
  ##
  ## The database values used must be data types the destination database
  ## understands. It is up to the user to ensure that the selected data type is
  ## available in the database they are using. Refer to your database
  ## documentation for what data types are available and supported.
  # [outputs.kaiwudb.convert]
  #   integer              = "INT"
  #   uinteger             = "UINTEGER"
  #   bigint               = "BIGINT"
  #   ubigint              = "UBIGINT"
  #   real                 = "REAL"
  #   double               = "DOUBLE"
  #   text                 = "TEXT"
  #   timestamp            = "TIMESTAMP"
  #   timestamptz          = "TIMESTAMP WITH TIME ZONE"
  #   defaultvalue         = "TEXT"
  #   unsigned             = "UNSIGNED"
  #   bool                 = "BOOL"
  #   json                 = "JSON"
  #   blob                 = "BLOB"
  #  ## This setting controls the behavior of the unsigned value. By default the
  #  ## setting will take the integer value and append the unsigned value to it. The other
  #  ## option is "literal", which will use the actual value the user provides to
  #  ## the unsigned option. This is useful for a database like ClickHouse where
  #  ## the unsigned value should use a value like "uint64".
  #  # conversion_style = "unsigned_suffix"
```
