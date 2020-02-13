# PgBouncer plugin

This PgBouncer plugin provides metrics for your PgBouncer load balancer.

More information about the meaning of these metrics can be found in the [PgBouncer Documentation](https://pgbouncer.github.io/usage.html)

## Reproduce behaviour with postgresql_extensible input

```
[[inputs.postgresql_extensible]]
  address = "host=localhost port=6432 user=postgres password=postgres sslmode=disable dbname=pgbouncer"
  [[inputs.postgresql_extensible.query]]
    measurement = "pgbouncer"
    sqlquery = "show stats;"
    withdbname = false
  [[inputs.postgresql_extensible.query]]
    measurement = "pgbouncer_pools"
    sqlquery = "show pools;"
    withdbname = false
```

Output metrics will be the same. Also you can use any of this queries to expose internal data:
```
SHOW STATS;
SHOW SERVERS;
SHOW CLIENTS;
SHOW POOLS;
```

## Configuration
Specify address via a postgresql connection string:

  `host=/run/postgresql port=6432 user=telegraf database=pgbouncer`

Or via an url matching:

  `postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]`

All connection parameters are optional.

Without the dbname parameter, the driver will default to a database with the same name as the user.
This dbname is just for instantiating a connection with the server and doesn't restrict the databases we are trying to grab metrics for.

### Configuration example
```
[[inputs.pgbouncer]]
  address = "postgres://telegraf@localhost/pgbouncer"
```
