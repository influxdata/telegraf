# PgBouncer plugin

This PgBouncer plugin provides metrics for your PgBouncer load balancer.

More information about the meaning of these metrics can be found in the [PgBouncer Documentation](https://pgbouncer.github.io/usage.html)

## Configuration
Specify address via a url matching:

  `postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]`

All connection parameters are optional.

Without the dbname parameter, the driver will default to a database with the same name as the user.
This dbname is just for instantiating a connection with the server and doesn't restrict the databases we are trying to grab metrics for.

### Configuration example
```
[[inputs.pgbouncer]]
  address = "postgres://telegraf@localhost/pgbouncer"
```
