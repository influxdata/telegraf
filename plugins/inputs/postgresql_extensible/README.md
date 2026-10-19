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
  ## This database is used to determine the server version and role, the list
  ## of databases and to run the queries when no database filter is set.
  address = "host=localhost user=postgres sslmode=disable"

  ## Whether to use prepared statements when connecting to the database.
  ## This should be set to false when connecting through a PgBouncer instance
  ## with pool_mode set to transaction.
  # prepared_statements = true

  ## Maximum lifetime of a connection. A connection older than this is closed
  ## when it is returned to the pool instead of being reused. Zero keeps the
  ## connections forever. Note that this does not interrupt queries, the
  ## lifetime is not enforced while a query is running.
  # max_lifetime = "0s"

  ## Timeout for a complete collection cycle, i.e. all queries in all
  ## databases. Individual queries may specify a shorter timeout.
  ## Zero means no limit on the duration of a collection.
  # timeout = "0s"

  ## Server role required to run the queries. Queries may override this.
  ## Valid values are "any", "primary" and "replica". The role is determined
  ## on every collection using pg_is_in_recovery().
  # role = "any"

  ## Run the queries in all databases matching the filters instead of the
  ## database of the connection only. Each entry is a regular expression
  ## matching the complete database name. Template databases and databases
  ## not allowing connections are never included.
  # datname_include = []
  # datname_exclude = []

  ## Interval for refreshing the server version, the server role and the
  ## list of databases. They are cached and refreshed on the first collection
  ## after each interval boundary on the wall clock, so a collection between
  ## two refreshes does not query the connection database at all.
  ## Zero reads them on every collection.
  # metadata_refresh_interval = "0s"

  ## Number of databases processed at the same time. Only relevant when the
  ## datname filters below select more than one database.
  # max_parallel_databases = 1

  ## Number of queries run at the same time within one database. Every query
  ## occupies one connection, so the maximum number of connections this
  ## plugin opens is max_parallel_databases * max_connections_per_database.
  # max_connections_per_database = 1

  ## Keep the idle connections of the pool for the database given in the
  ## address open between collections. If unset, they are closed after a
  ## collection. The connection to any other database is always closed as
  ## soon as its queries are done, whatever this is set to.
  # keep_idle_connections = true

  ## Use all string columns as tags instead of fields. Queries may override this.
  # string_columns_as_tags = false

  ## Convert numeric columns to floating point fields instead of string
  ## fields. Queries may override this.
  # numeric_as_float = false

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
  ## The timeout option limits the runtime of the query and defaults to the
  ## remaining time of the collection cycle.
  ##
  ## The role, string_columns_as_tags and numeric_as_float options override
  ## the plugin-level settings for this query.
  [[inputs.postgresql_extensible.query]]
    measurement = "pg_stat_database"
    sqlquery = "SELECT * FROM pg_stat_database"
    min_version = 901
    tagvalue = ""
    # timeout = "10s"
    # role = "any"
    # string_columns_as_tags = false
    # numeric_as_float = false
  [[inputs.postgresql_extensible.query]]
    script = "your_sql-filepath.sql"
    min_version = 901
    max_version = 1300
    role = "primary"
```

The system can be easily extended using homemade metrics collection tools or
using the postgresql extensions [pg_stat_statements][pg_stat_statements],
[pg_proctab][pg_proctab] or [powa][powa].

[pg_stat_statements]: http://www.postgresql.org/docs/current/static/pgstatstatements.html
[pg_proctab]: https://github.com/markwkm/pg_proctab
[powa]: http://dalibo.github.io/powa/

### Sample Queries

* telegraf.conf postgresql_extensible queries (assuming that you have configured
 correctly your connection)

```toml
[[inputs.postgresql_extensible.query]]
  sqlquery="SELECT * FROM pg_stat_database"
  version=901
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="SELECT * FROM pg_stat_bgwriter"
  version=901
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="select * from sessions"
  version=901
  withdbname=false
  tagvalue="db,username,state"
[[inputs.postgresql_extensible.query]]
  sqlquery="""\
  select setting as max_connections from pg_settings where \
  name='max_connections'"""
  version=801
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="select * from pg_stat_kcache"
  version=901
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="""\
  select setting as shared_buffers from pg_settings where \
  name='shared_buffers'"""
  version=801
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="""\
  SELECT db, count( distinct blocking_pid ) AS num_blocking_sessions,\
  count( distinct blocked_pid) AS num_blocked_sessions FROM \
  public.blocking_procs group by db"""
  version=901
  withdbname=false
  tagvalue="db"
[[inputs.postgresql_extensible.query]]
  sqlquery="""
    SELECT type, (enabled || '') AS enabled, COUNT(*)
      FROM application_users
      GROUP BY type, enabled
  """
  version=901
  withdbname=false
  tagvalue="type,enabled"
```

### Postgresql Side

postgresql.conf :

```sql
shared_preload_libraries = 'pg_stat_statements,pg_stat_kcache'
```

Please follow the requirements to setup those extensions.

In the database (can be a specific monitoring db)

```sql
create extension pg_stat_statements;
create extension pg_stat_kcache;
create extension pg_proctab;
```

(assuming that the extension is installed on the OS Layer)

* pg_stat_kcache is available on the postgresql.org yum repo
* pg_proctab is available at : <https://github.com/markwkm/pg_proctab>

### Views

* Blocking sessions

```sql
CREATE OR REPLACE VIEW public.blocking_procs AS
 SELECT a.datname AS db,
    kl.pid AS blocking_pid,
    ka.usename AS blocking_user,
    ka.query AS blocking_query,
    bl.pid AS blocked_pid,
    a.usename AS blocked_user,
    a.query AS blocked_query,
    to_char(age(now(), a.query_start), 'HH24h:MIm:SSs'::text) AS age
   FROM pg_locks bl
     JOIN pg_stat_activity a ON bl.pid = a.pid
     JOIN pg_locks kl ON bl.locktype = kl.locktype AND NOT bl.database IS
     DISTINCT FROM kl.database AND NOT bl.relation IS DISTINCT FROM kl.relation
     AND NOT bl.page IS DISTINCT FROM kl.page AND NOT bl.tuple IS DISTINCT FROM
     kl.tuple AND NOT bl.virtualxid IS DISTINCT FROM kl.virtualxid AND NOT
     bl.transactionid IS DISTINCT FROM kl.transactionid AND NOT bl.classid IS
     DISTINCT FROM kl.classid AND NOT bl.objid IS DISTINCT FROM kl.objid AND
      NOT bl.objsubid IS DISTINCT FROM kl.objsubid AND bl.pid <> kl.pid
     JOIN pg_stat_activity ka ON kl.pid = ka.pid
  WHERE kl.granted AND NOT bl.granted
  ORDER BY a.query_start;
```

* Sessions Statistics

```sql
CREATE OR REPLACE VIEW public.sessions AS
 WITH proctab AS (
         SELECT pg_proctab.pid,
                CASE
                    WHEN pg_proctab.state::text = 'R'::bpchar::text
                      THEN 'running'::text
                    WHEN pg_proctab.state::text = 'D'::bpchar::text
                      THEN 'sleep-io'::text
                    WHEN pg_proctab.state::text = 'S'::bpchar::text
                      THEN 'sleep-waiting'::text
                    WHEN pg_proctab.state::text = 'Z'::bpchar::text
                      THEN 'zombie'::text
                    WHEN pg_proctab.state::text = 'T'::bpchar::text
                      THEN 'stopped'::text
                    ELSE NULL::text
                END AS proc_state,
            pg_proctab.ppid,
            pg_proctab.utime,
            pg_proctab.stime,
            pg_proctab.vsize,
            pg_proctab.rss,
            pg_proctab.processor,
            pg_proctab.rchar,
            pg_proctab.wchar,
            pg_proctab.syscr,
            pg_proctab.syscw,
            pg_proctab.reads,
            pg_proctab.writes,
            pg_proctab.cwrites
           FROM pg_proctab() pg_proctab(pid, comm, fullcomm, state, ppid, pgrp,
             session, tty_nr, tpgid, flags, minflt, cminflt, majflt, cmajflt,
             utime, stime, cutime, cstime, priority, nice, num_threads,
             itrealvalue, starttime, vsize, rss, exit_signal, processor,
             rt_priority, policy, delayacct_blkio_ticks, uid, username, rchar,
             wchar, syscr, syscw, reads, writes, cwrites)
        ), stat_activity AS (
         SELECT pg_stat_activity.datname,
            pg_stat_activity.pid,
            pg_stat_activity.usename,
                CASE
                    WHEN pg_stat_activity.query IS NULL THEN 'no query'::text
                    WHEN pg_stat_activity.query IS NOT NULL AND
                    pg_stat_activity.state = 'idle'::text THEN 'no query'::text
                    ELSE regexp_replace(pg_stat_activity.query, '[\n\r]+'::text,
                       ' '::text, 'g'::text)
                END AS query
           FROM pg_stat_activity
        )
 SELECT stat.datname::name AS db,
    stat.usename::name AS username,
    stat.pid,
    proc.proc_state::text AS state,
('"'::text || stat.query) || '"'::text AS query,
    (proc.utime/1000)::bigint AS session_usertime,
    (proc.stime/1000)::bigint AS session_systemtime,
    proc.vsize AS session_virtual_memory_size,
    proc.rss AS session_resident_memory_size,
    proc.processor AS session_processor_number,
    proc.rchar AS session_bytes_read,
    proc.rchar-proc.reads AS session_logical_bytes_read,
    proc.wchar AS session_bytes_written,
    proc.wchar-proc.writes AS session_logical_bytes_writes,
    proc.syscr AS session_read_io,
    proc.syscw AS session_write_io,
    proc.reads AS session_physical_reads,
    proc.writes AS session_physical_writes,
    proc.cwrites AS session_cancel_writes
   FROM proctab proc,
    stat_activity stat
  WHERE proc.pid = stat.pid;
```

## Databases

Without `datname_include` and `datname_exclude` all queries are run in the
database given in the `address` (or the driver default), which is the
behavior of previous versions of this plugin. If any of the filters is set,
the list of databases is read from `pg_database` and the queries are run in
every database matching the filters. Template databases and databases not
accepting connections are never included.

Every filter entry is a regular expression matching the complete database
name, so `app` only matches the database `app` while `app_.*` matches
`app_prod` and `app_test`. A database is used if it matches any of the
`datname_include` entries (or the list is empty) and none of the
`datname_exclude` entries.

The database list is refreshed on the first collection after each
`metadata_refresh_interval` boundary on the wall clock, so a new database
shows up after at most `metadata_refresh_interval` plus one collection
interval.

## Server role

The server role is determined using `pg_is_in_recovery()`. Queries with
`role = "primary"` are only run if the server is not in recovery, queries
with `role = "replica"` only if it is. The plugin-level `role` option sets
the default for all queries.

## Cached server information

The server version, the server role and the list of databases are read
together and cached for `metadata_refresh_interval`. A collection between
two refreshes uses the cached values and does not query the connection
database at all. After a role change the plugin therefore keeps using the
previous role for up to one refresh interval, so choose the interval short
enough for how fast a promoted server should be picked up. The default of
zero reads them on every collection.

## Timeouts and concurrency

The plugin-level `timeout` limits a complete collection cycle. Each query
is run with its own `timeout` if set, or the remaining time of the cycle
otherwise. The default of zero means no limit, as in previous versions of
this plugin.

Concurrency has two independent knobs. `max_parallel_databases` sets how
many databases are processed at the same time, `max_connections_per_database`
how many queries run at the same time within one of them. Every query
occupies one connection, so the plugin never opens more than the product of
the two:

| `max_parallel_databases` | `max_connections_per_database` | connections |
|---|---|---|
| 1 | 1 | 1 |
| 1 | 5 | 5 |
| 5 | 1 | 5 |
| 5 | 2 | 10 |

Both default to one, which is the behavior of previous versions of this
plugin: one connection running the queries one after another.

`keep_idle_connections` only covers the pool for the database given in the
`address`. The connection to any other database is always closed as soon as
its queries are done, to not hold one connection per database on a server
that may have hundreds of them.

## Compatibility

The defaults keep the behavior of previous versions of this plugin: no
timeout, no caching of the server information, a single connection running
the queries sequentially in the database of the `address`, `numeric` columns
as string fields and the connection kept open between collections. The
deprecated `databases`, `withdbname` and `version` options keep working as
before.

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
    * all string columns if `string_columns_as_tags` is set
  * fields:
    * all remaining columns; `numeric` columns are converted to floats if
      `numeric_as_float` is set

The `stats_reset` column is never reported. To drop further columns use the
`fieldexclude` and `tagexclude` options every input plugin supports.
