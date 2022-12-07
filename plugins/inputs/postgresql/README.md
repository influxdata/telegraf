# PostgreSQL Input Plugin

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

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Read metrics from one or many postgresql servers
[[inputs.postgresql]]
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqgotest password=... sslmode=... dbname=app_production
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
