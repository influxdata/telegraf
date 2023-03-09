# PostgreSQL Extensible Input Plugin

This postgresql plugin provides metrics for your postgres database. It has been
designed to parse SQL queries in the plugin section of your `telegraf.conf`.

The example below has two queries are specified, with the following parameters:

* The SQL query itself
* The minimum PostgreSQL version supported (the numeric display visible in pg_settings)
* A boolean to define if the query has to be run against some specific database (defined in the `databases` variable of the plugin section)
* The name of the measurement
* A list of the columns to be defined as tags

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `address` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read metrics from one or many postgresql servers
[[inputs.postgresql_extensible]]
  # specify address via a url matching:
  # postgres://[pqgotest[:password]]@host:port[/dbname]?sslmode=...&statement_timeout=...
  # or a simple string:
  #   host=localhost port=5432 user=pqgotest password=... sslmode=... dbname=app_production
  #
  # All connection parameters are optional.
  # Without the dbname parameter, the driver will default to a database
  # with the same name as the user. This dbname is just for instantiating a
  # connection with the server and doesn't restrict the databases we are trying
  # to grab metrics for.
  #
  address = "host=localhost user=postgres sslmode=disable"

  ## A list of databases to pull metrics about.
  ## deprecated in 1.22.3; use the sqlquery option to specify database to use
  # databases = ["app_production", "testing"]

  ## Whether to use prepared statements when connecting to the database.
  ## This should be set to false when connecting through a PgBouncer instance
  ## with pool_mode set to transaction.
  prepared_statements = true

  # Define the toml config where the sql queries are stored
  # The script option can be used to specify the .sql file path.
  # If script and sqlquery options specified at same time, sqlquery will be used
  #
  # the tagvalue field is used to define custom tags (separated by comas).
  # the query is expected to return columns which match the names of the
  # defined tags. The values in these columns must be of a string-type,
  # a number-type or a blob-type.
  #
  # The timestamp field is used to override the data points timestamp value. By
  # default, all rows inserted with current time. By setting a timestamp column,
  # the row will be inserted with that column's value.
  #
  # Structure :
  # [[inputs.postgresql_extensible.query]]
  #   sqlquery string
  #   version string
  #   withdbname boolean
  #   tagvalue string (coma separated)
  #   timestamp string
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_database where datname"
    version=901
    withdbname=false
    tagvalue=""
  [[inputs.postgresql_extensible.query]]
    script="your_sql-filepath.sql"
    version=901
    withdbname=false
    tagvalue=""
```

The system can be easily extended using homemade metrics collection tools or
using postgresql extensions ([pg_stat_statements][1], [pg_proctab][2] or
[powa][3])

[1]: http://www.postgresql.org/docs/current/static/pgstatstatements.html

[2]: https://github.com/markwkm/pg_proctab

[3]: http://dalibo.github.io/powa/

## Sample Queries

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
  sqlquery="select setting as max_connections from pg_settings where \
  name='max_connections'"
  version=801
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="select * from pg_stat_kcache"
  version=901
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="select setting as shared_buffers from pg_settings where \
  name='shared_buffers'"
  version=801
  withdbname=false
  tagvalue=""
[[inputs.postgresql_extensible.query]]
  sqlquery="SELECT db, count( distinct blocking_pid ) AS num_blocking_sessions,\
  count( distinct blocked_pid) AS num_blocked_sessions FROM \
  public.blocking_procs group by db"
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

## Postgresql Side

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

## Views

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
    * db
    * server
