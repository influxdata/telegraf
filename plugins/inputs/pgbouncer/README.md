# PgBouncer Input Plugin

This plugin collects metrics from a [PgBouncer load balancer][pgbouncer]
instance. Check the [documentation][metric_docs] for available metrics and their
meaning.

> [!NOTE]
> This plugin requires PgBouncer v1.5+.

⭐ Telegraf v1.8.0
🏷️ server, web
💻 all

[pgbouncer]: https://pgbouncer.github.io
[metric_docs]: https://pgbouncer.github.io/usage.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from one or many pgbouncer servers
[[inputs.pgbouncer]]
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@host:port[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost port=5432 user=pqgotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional.
  ##
  address = "host=localhost user=pgbouncer sslmode=disable"

  ## Specify which "show" commands to gather metrics for.
  ## Choose from: "stats", "pools", "lists", "databases"
  # show_commands = ["stats", "pools"]
```

To specify the `address` use either a PostgreSQL connection string:

```text
host=/run/postgresql port=6432 user=telegraf database=pgbouncer
```

or via an URL of the following form:

```text
postgres://[pqgotest[:password]]@host:port[/dbname]?sslmode=[disable|verify-ca|verify-full]
```

All connection parameters are optional.

Without the `dbname` parameter, the driver will default to a database with the
same name as the user. The `dbname` is for instantiating a connection with the
server only and doesn't restrict access to the databases queried.

## Metrics

- pgbouncer
  - tags:
    - db
    - server
  - fields:
    - avg_query_count
    - avg_query_time
    - avg_wait_time
    - avg_xact_count
    - avg_xact_time
    - total_query_count
    - total_query_time
    - total_received
    - total_sent
    - total_wait_time
    - total_xact_count
    - total_xact_time

- pgbouncer_pools
  - tags:
    - db
    - pool_mode
    - server
    - user
  - fields:
    - cl_active
    - cl_waiting
    - maxwait
    - maxwait_us
    - sv_active
    - sv_idle
    - sv_login
    - sv_tested
    - sv_used

- pgbouncer_lists
  - tags:
    - db
    - server
    - user
  - fields:
    - databases
    - users
    - pools
    - free_clients
    - used_clients
    - login_clients
    - free_servers
    - used_servers
    - dns_names
    - dns_zones
    - dns_queries

- pgbouncer_databases
  - tags:
    - db
    - pg_dbname
    - server
    - user
  - fields:
    - current_connections
    - pool_size
    - min_pool_size
    - reserve_pool
    - max_connections
    - paused
    - disabled

## Example Output

```text
pgbouncer,db=pgbouncer,server=host\=debian-buster-postgres\ user\=dbn\ port\=6432\ dbname\=pgbouncer\  avg_query_count=0i,avg_query_time=0i,avg_wait_time=0i,avg_xact_count=0i,avg_xact_time=0i,total_query_count=26i,total_query_time=0i,total_received=0i,total_sent=0i,total_wait_time=0i,total_xact_count=26i,total_xact_time=0i 1581569936000000000
pgbouncer_pools,db=pgbouncer,pool_mode=statement,server=host\=debian-buster-postgres\ user\=dbn\ port\=6432\ dbname\=pgbouncer\ ,user=pgbouncer cl_active=1i,cl_waiting=0i,maxwait=0i,maxwait_us=0i,sv_active=0i,sv_idle=0i,sv_login=0i,sv_tested=0i,sv_used=0i 1581569936000000000
pgbouncer_lists,db=pgbouncer,server=host\=debian-buster-postgres\ user\=dbn\ port\=6432\ dbname\=pgbouncer\ ,user=pgbouncer databases=1i,dns_names=0i,dns_queries=0i,dns_zones=0i,free_clients=47i,free_servers=0i,login_clients=0i,pools=1i,used_clients=3i,used_servers=0i,users=4i 1581569936000000000
pgbouncer_databases,db=pgbouncer,pg_dbname=pgbouncer,server=host\=debian-buster-postgres\ user\=dbn\ port\=6432\ dbname\=pgbouncer\ name=pgbouncer disabled=0i,pool_size=2i,current_connections=0i,min_pool_size=0i,reserve_pool=0i,max_connections=0i,paused=0i 1581569936000000000
pgbouncer_databases,db=postgres,pg_dbname=postgres,server=host\=debian-buster-postgres\ user\=dbn\ port\=6432\ dbname\=pgbouncer\ name=postgres current_connections=0i,disabled=0i,pool_size=20i,min_pool_size=0i,reserve_pool=0i,paused=0i,max_connections=0i 1581569936000000000
```
