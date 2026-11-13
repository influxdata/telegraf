# PostgreSQL Extensible Input Plugin

This plugin queries a [PostgreSQL][postgres] server and provides metrics for
the returned result. This is useful when using PostgreSQL extensions to collect
additional metrics. The queries can be run in many databases of the server,
restricted to the primary or replica role, limited in runtime and executed
concurrently.

> [!TIP]
> Please also check the more generic [sql input plugin][inputs_sql].

⭐ Telegraf v0.12.0
🏷️ datastore
💻 all

[postgres]: https://www.postgresql.org/
[inputs_sql]: /plugins/inputs/sql/README.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret store support

This plugin supports secrets from secret stores for the `address` option.
See the [secret store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read metrics from one or many postgresql servers and databases
[[inputs.postgresql_extensible]]
  ## Specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@host:port[/dbname]?sslmode=...
  ## or a simple string:
  ##   host=localhost port=5432 user=pqgotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional. Without the dbname parameter,
  ## the driver will default to a database with the same name as the user.
  address = "host=localhost user=postgres sslmode=disable"

  ## Whether to use prepared statements. Each new connection prepares the
  ## queries again, which costs at least one extra round trip per query. With
  ## keep_idle_connections = false that happens needlessly on every collection,
  ## so it makes sense to set this to false.
  ##
  ## Also set this to false when connecting through a PgBouncer with pool_mode
  ## set to transaction, unless its max_prepared_statements is above zero
  ## (available since PgBouncer 1.21, the default since 1.24).
  # prepared_statements = true

  ## Maximum lifetime of a connection. A connection older than this is closed
  ## once it is not in use. Zero keeps the connections forever. Note that this
  ## does not interrupt queries, the lifetime is not enforced while a query is
  ## running.
  # max_lifetime = "0s"

  ## Timeout for a complete collection cycle. Zero means no limit.
  ## Queries may set their own timeout.
  # timeout = "0s"

  ## Server role required to run the queries. Queries may override this.
  ## Valid values are "any", "primary" and "replica".
  # role = "any"

  ## Run the queries in all databases matching the filters instead of the
  ## database of the connection only. Each entry is a glob pattern.
  # datname_include = []
  # datname_exclude = []

  ## Interval for refreshing the cached server version, server role and list
  ## of databases. Zero reads them on every collection.
  # metadata_refresh_interval = "0s"

  ## Number of databases processed at the same time. Only relevant when the
  ## datname filters above select more than one database.
  # max_parallel_databases = 1

  ## Number of queries run at the same time within one database.
  # max_connections_per_database = 1

  ## Keep the connections open between collections instead of closing them
  ## once a collection is done. With many databases this may exceed the
  ## max_connections setting of the server.
  # keep_idle_connections = true

  ## Use the columns of the text types (text, varchar, char and name) as tags
  ## instead of fields. Queries may override this.
  # string_columns_as_tags = false

  ## Field type for columns of the PostgreSQL numeric type, which has more
  ## precision than a float can hold. Available values are "string" and
  ## "float". Queries may override this.
  # numeric_type = "string"

  ## Queries to run
  ##
  ## The sqlquery option contains the SQL text to run, the script option a
  ## path to a file containing it. If both are given, sqlquery is used.
  ##
  ## The measurement option defines the measurement name for the metrics
  ## produced by the query. Default is "postgresql".
  ##
  ## The tagvalue option is a comma separated list of columns to use as tags.
  ##
  ## The timestamp option names a column whose value is used as the metric
  ## timestamp instead of the collection time.
  ##
  ## The min_version and max_version options restrict the query to the given
  ## server versions. The version is the server_version_num divided by 100,
  ## e.g. 9.6.2 -> 906 and 15.2 -> 1500. The query is not run on max_version.
  ##
  ## The timeout option limits the runtime of the query.
  ##
  ## The role, string_columns_as_tags and numeric_type options override the
  ## plugin-level settings for this query.
  [[inputs.postgresql_extensible.query]]
    measurement = "pg_stat_database"
    sqlquery = "SELECT * FROM pg_stat_database"
    min_version = 901
    tagvalue = ""
    # timeout = "0s"
    # role = "any"
    # string_columns_as_tags = false
    # numeric_type = "string"
  [[inputs.postgresql_extensible.query]]
    script = "your_sql-filepath.sql"
    min_version = 901
    max_version = 1300
    role = "primary"
```

### Example with several intervals

The following example is not a ready configuration. It shows the basic
patterns: what can be collected at which interval and how the options work
together.

```toml
## Collecting general PostgreSQL metrics
[[inputs.postgresql_extensible]]
  interval = "60s"
  address = "host=localhost user=postgres dbname=postgres sslmode=disable"

  ## Keep the connection open between collections
  keep_idle_connections = true

  ## Replace a kept connection once it is older than this
  max_lifetime = "1h"

  ## Read the metadata every 10 minutes instead of on every collection: the
  ## server version, pg_is_in_recovery() and, with the datname filters, the
  ## list of databases
  metadata_refresh_interval = "10m"

  ## Stop the collection after this time. For example, if the agent collecting
  ## the metrics has a 30 second timeout, there is no point in continuing.
  timeout = "30s"

  [[inputs.postgresql_extensible.query]]
    measurement = "pg_stat_database"
    sqlquery = "SELECT datname, numbackends, xact_commit, xact_rollback, deadlocks FROM pg_stat_database WHERE datname IS NOT NULL"

  [[inputs.postgresql_extensible.query]]
    measurement = "pg_replication_slots"
    sqlquery = "SELECT slot_name, active, pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS lag_bytes FROM pg_replication_slots"
    tagvalue = "slot_name"
    role = "primary"
    numeric_type = "float"

  ## Other queries
  # [[inputs.postgresql_extensible.query]]
  #   measurement = "..."
  #   sqlquery = "SELECT ..."

## Collecting the query statistics, requires the pg_stat_statements extension
[[inputs.postgresql_extensible]]
  interval = "5m"
  address = "host=localhost user=postgres dbname=postgres sslmode=disable"

  ## Close the connection after each collection
  keep_idle_connections = false

  ## Without a kept connection prepared statements only add round trips
  prepared_statements = false

  metadata_refresh_interval = "10m"
  timeout = "30s"

  [[inputs.postgresql_extensible.query]]
    measurement = "pg_stat_statements"
    sqlquery = "SELECT userid, dbid, queryid, calls, total_exec_time, rows FROM pg_stat_statements"
    tagvalue = "userid,dbid,queryid"

## Collecting the tables and indexes of the databases
[[inputs.postgresql_extensible]]
  interval = "30m"
  address = "host=localhost user=postgres dbname=postgres sslmode=disable"
  prepared_statements = false
  timeout = "30s"

  ## Run the queries in every database except postgres, two databases at a
  ## time and two queries at a time within each database, and close the
  ## connections of each database once its queries are done
  datname_include = ["*"]
  datname_exclude = ["postgres"]
  max_parallel_databases = 2
  max_connections_per_database = 2
  keep_idle_connections = false

  ## Each query below stops after 10 seconds, so that a slow one does not use
  ## up the collection time left for the other queries
  [[inputs.postgresql_extensible.query]]
    measurement = "pg_stat_user_tables"
    sqlquery = "SELECT * FROM pg_stat_user_tables"
    tagvalue = "schemaname,relname"
    timeout = "10s"

  [[inputs.postgresql_extensible.query]]
    measurement = "pg_stat_user_indexes"
    sqlquery = "SELECT * FROM pg_stat_user_indexes"
    tagvalue = "schemaname,relname,indexrelname"
    timeout = "10s"

  ## A file with the SQL query instead of sqlquery
  [[inputs.postgresql_extensible.query]]
    measurement = "pg_invalid_indexes"
    script = "/etc/telegraf/sql/invalid_indexes.sql"
    tagvalue = "table_name,index_name"
    timeout = "10s"
```

## Databases

Without `datname_include` and `datname_exclude` all queries are run in the
database given in the `address` (or the driver default). If any of the
filters is set, the list of databases is read from `pg_database` and the
queries are run in every database matching the filters. Template databases
and databases not accepting connections are never included.

Every filter entry is a glob pattern matching the database name, so `app`
only matches the database `app` while `app_*` matches `app_prod` and
`app_test`. A database is used if it matches any of the
`datname_include` entries (or the list is empty) and none of the
`datname_exclude` entries.

A new database shows up on the next collection, or with
`metadata_refresh_interval` set, after at most that interval plus one
collection interval.

## Server role

The server role is read with `pg_is_in_recovery()` and cached if
`metadata_refresh_interval` is set. Queries with `role = "primary"` are only
run if the server is not in recovery, queries with `role = "replica"` only if
it is. The plugin-level `role` option sets the default for all queries.

## Cached server information

The server version, the server role and the list of databases are read
together from the database of the `address` and cached for
`metadata_refresh_interval`. They are refreshed on the first collection after
each interval boundary on the wall clock. A collection between two refreshes
uses the cached values and does not read the server information again. After
a role change the plugin therefore keeps using the previous role for up to
`metadata_refresh_interval` plus one collection interval. The default of zero
reads them on every collection.

## Timeouts and concurrency

The plugin-level `timeout` limits a complete collection cycle. Each query
is run with its own `timeout` if set, or the remaining time of the cycle
otherwise. The default of zero means no limit.

Concurrency has two independent knobs. `max_parallel_databases` sets how
many databases are processed at the same time, `max_connections_per_database`
how many queries run at the same time within one of them. Every query
occupies one connection, so a collection opens at most the product of the
two:

| `max_parallel_databases` | `max_connections_per_database` | connections |
|---|---|---|
| 1 | 1 | 1 |
| 1 | 5 | 5 |
| 5 | 1 | 5 |
| 5 | 2 | 10 |

Both default to one: a single connection running the queries one after
another.

One connection comes on top of that product while a collection refreshes
the server information, and only when the database of the `address` is not
among the ones queried. Otherwise that
connection is taken from the pool the queries use anyway.

`keep_idle_connections` keeps the connections of every queried database
open until the next collection. A collection querying many databases then
holds up to the number of databases times `max_connections_per_database`
connections, plus one for the database of the `address` if it is not among
the ones queried, which may exceed the `max_connections` setting of the server.
Set it to false to close the connections as soon as a collection is done.

The pool of a database that drops out of the filtered list is closed on the
next collection.

## Example Output

The example out below was taken by running the query

```sql
select count(*)*100 / (select cast(nullif(setting, '') AS integer) from pg_settings where name='max_connections') as percentage_of_used_cons from pg_stat_activity
```

Which generates the following

```text
postgresql,db=postgres,server=dbname\=postgres\ host\=localhost\ port\=5432\ statement_timeout\=10000\ user\=postgres percentage_of_used_cons=6i 1672400531000000000
```

## Metrics

The metrics collected by this input plugin will depend on the configured query.

By default, the following format will be used

* postgresql
  * tags:
    * db - the database the query was run in, or the value of the `datname`
      column if present
    * server - the sanitized connection address
    * all columns listed in `tagvalue`
    * all columns of the `text`, `varchar`, `char` and `name` types if
      `string_columns_as_tags` is set
  * fields:
    * all remaining columns; `numeric` columns are converted to floats if
      `numeric_type` is `"float"`, otherwise they are reported as strings

The `stats_reset` column is never reported. To drop further columns use the
`fieldexclude` and `tagexclude` options every input plugin supports.
