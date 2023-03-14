# PostgreSQL Input Plugin

The `postgresql` plugin provides metrics for your PostgreSQL Server instance.
Recorded metrics are lightweight and use Dynamic Management Views supplied
by PostgreSQL.

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
[[inputs.postgresql]]
  ## Specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]&statement_timeout=...
  ## or a simple string:
  ##   host=localhost user=pqgotest password=... sslmode=... dbname=app_production
  ## Users can pass the path to the socket as the host value to use a socket
  ## connection (e.g. `/var/run/postgresql`).
  ##
  ## All connection parameters are optional.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  address = "host=localhost user=postgres sslmode=disable"

  ## A custom name for the database that will be used as the "server" tag in the
  ## measurement output. If not specified, a default one generated from
  ## the connection address is used.
  # outputaddress = "db01"

  ## connection configuration.
  ## maxlifetime - specify the maximum lifetime of a connection.
  ## default is forever (0s)
  ##
  ## Note that this does not interrupt queries, the lifetime will not be enforced
  ## whilst a query is running
  # max_lifetime = "0s"

  ## A  list of databases to explicitly ignore.  If not specified, metrics for all
  ## databases are gathered.  Do NOT use with the 'databases' option.
  # ignored_databases = ["postgres", "template0", "template1"]

  ## A list of databases to pull metrics about. If not specified, metrics for all
  ## databases are gathered.  Do NOT use with the 'ignored_databases' option.
  # databases = ["app_production", "testing"]

  ## Whether to use prepared statements when connecting to the database.
  ## This should be set to false when connecting through a PgBouncer instance
  ## with pool_mode set to transaction.
  prepared_statements = true
```

Specify address via a postgresql connection string:

```text
host=localhost port=5432 user=telegraf database=telegraf
```

Or via an url matching:

```text
postgres://[pqgotest[:password]]@host:port[/dbname]?sslmode=[disable|verify-ca|verify-full]
```

Users can pass the path to the socket as the host value to use a socket
connection (e.g. `/var/run/postgresql`).

It is also possible to specify a query timeout maximum execution time (in ms)
for any individual statement passed over the connection

```text
postgres://[pqgotest[:password]]@host:port[/dbname]?sslmode=[disable|verify-ca|verify-full]&statement_timeout=10000
```

All connection parameters are optional. Without the dbname parameter, the driver
will default to a database with the same name as the user. This dbname is just
for instantiating a connection with the server and doesn't restrict the
databases we are trying to grab metrics for.

A list of databases to explicitly ignore.  If not specified, metrics for all
databases are gathered.  Do NOT use with the 'databases' option.

```text
ignored_databases = ["postgres", "template0", "template1"]`
```

A list of databases to pull metrics about. If not specified, metrics for all
databases are gathered.  Do NOT use with the 'ignored_databases' option.

```text
databases = ["app_production", "testing"]`
```

### TLS Configuration

Add the `sslkey`, `sslcert` and `sslrootcert` options to your DSN:

```shell
host=localhost user=pgotest dbname=app_production sslmode=require sslkey=/etc/telegraf/key.pem sslcert=/etc/telegraf/cert.pem sslrootcert=/etc/telegraf/ca.pem
```

## Metrics

This postgresql plugin provides metrics for your postgres database. It currently
works with postgres versions 8.1+. It uses data from the built in
_pg_stat_database_ and pg_stat_bgwriter views. The metrics recorded depend on
your version of postgres. See table:

```sh
pg version      9.2+   9.1   8.3-9.0   8.1-8.2   7.4-8.0(unsupported)
---             ---    ---   -------   -------   -------
datid            x      x       x         x
datname          x      x       x         x
numbackends      x      x       x         x         x
xact_commit      x      x       x         x         x
xact_rollback    x      x       x         x         x
blks_read        x      x       x         x         x
blks_hit         x      x       x         x         x
tup_returned     x      x       x
tup_fetched      x      x       x
tup_inserted     x      x       x
tup_updated      x      x       x
tup_deleted      x      x       x
conflicts        x      x
temp_files       x
temp_bytes       x
deadlocks        x
blk_read_time    x
blk_write_time   x
stats_reset*     x      x
```

_* value ignored and therefore not recorded._

More information about the meaning of these metrics can be found in the
[PostgreSQL Documentation][1].

[1]: http://www.postgresql.org/docs/9.2/static/monitoring-stats.html#PG-STAT-DATABASE-VIEW

## Example Output

```text
postgresql,db=postgres_global,server=dbname\=postgres\ host\=localhost\ port\=5432\ statement_timeout\=10000\ user\=postgres tup_fetched=1271i,tup_updated=5i,session_time=1451414320768.855,xact_rollback=2i,conflicts=0i,blk_write_time=0,temp_bytes=0i,datid=0i,sessions_fatal=0i,tup_returned=1339i,sessions_abandoned=0i,blk_read_time=0,blks_read=88i,idle_in_transaction_time=0,sessions=0i,active_time=0,tup_inserted=24i,tup_deleted=0i,temp_files=0i,numbackends=0i,xact_commit=4i,sessions_killed=0i,blks_hit=5616i,deadlocks=0i 1672399790000000000
postgresql,db=postgres,host=oss_cluster_host,server=dbname\=postgres\ host\=localhost\ port\=5432\ statement_timeout\=10000\ user\=postgres conflicts=0i,sessions_abandoned=2i,active_time=460340.823,tup_returned=119382i,tup_deleted=0i,blk_write_time=0,xact_commit=305i,blks_hit=16358i,deadlocks=0i,sessions=12i,numbackends=1i,temp_files=0i,xact_rollback=5i,sessions_fatal=0i,datname="postgres",blk_read_time=0,idle_in_transaction_time=0,temp_bytes=0i,tup_inserted=3i,tup_updated=0i,blks_read=299i,datid=5i,session_time=469056.613,sessions_killed=0i,tup_fetched=5550i 1672399790000000000
postgresql,db=template1,host=oss_cluster_host,server=dbname\=postgres\ host\=localhost\ port\=5432\ statement_timeout\=10000\ user\=postgres active_time=0,idle_in_transaction_time=0,blks_read=1352i,sessions_abandoned=0i,tup_fetched=28544i,session_time=0,sessions_killed=0i,temp_bytes=0i,tup_returned=188541i,xact_commit=1168i,blk_read_time=0,sessions_fatal=0i,datid=1i,datname="template1",conflicts=0i,xact_rollback=0i,numbackends=0i,deadlocks=0i,sessions=0i,tup_inserted=17520i,temp_files=0i,tup_updated=743i,blk_write_time=0,blks_hit=99487i,tup_deleted=34i 1672399790000000000
postgresql,db=template0,host=oss_cluster_host,server=dbname\=postgres\ host\=localhost\ port\=5432\ statement_timeout\=10000\ user\=postgres sessions=0i,datid=4i,tup_updated=0i,sessions_abandoned=0i,blk_write_time=0,numbackends=0i,blks_read=0i,blks_hit=0i,sessions_fatal=0i,temp_files=0i,deadlocks=0i,conflicts=0i,xact_commit=0i,xact_rollback=0i,session_time=0,datname="template0",tup_returned=0i,tup_inserted=0i,idle_in_transaction_time=0,tup_fetched=0i,active_time=0,temp_bytes=0i,tup_deleted=0i,blk_read_time=0,sessions_killed=0i 1672399790000000000
postgresql,db=postgres,host=oss_cluster_host,server=dbname\=postgres\ host\=localhost\ port\=5432\ statement_timeout\=10000\ user\=postgres buffers_clean=0i,buffers_alloc=426i,checkpoints_req=1i,buffers_checkpoint=50i,buffers_backend_fsync=0i,checkpoint_write_time=5053,checkpoints_timed=26i,checkpoint_sync_time=26,maxwritten_clean=0i,buffers_backend=9i 1672399790000000000
```
