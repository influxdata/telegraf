# Available SQL drivers for the SQL input plugin

This is a list of available drivers for the SQL input plugin. The data-source-name (DSN) is driver specific and
might change between versions. Please check the driver documentation for available options and the format.

| database             | driver                                                    | aliases         | example DSN                                                                            | comment                                                                                                               |
| -------------------- | --------------------------------------------------------- | --------------- | -------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| CockroachDB          | [cockroach](https://github.com/jackc/pgx)                 | postgres or pgx | see _postgres_ driver                                                                  | uses PostgresQL driver                                                                                                |
| MariaDB              | [maria](https://github.com/go-sql-driver/mysql)           | mysql           | see _mysql_ driver                                                                     | uses MySQL driver                                                                                                     |
| Microsoft SQL Server | [sqlserver](https://github.com/denisenkom/go-mssqldb)     | mssql           | `username:password@host/instance?param1=value&param2=value`                            | uses newer _sqlserver_ driver                                                                                         |
| MySQL                | [mysql](https://github.com/go-sql-driver/mysql)           |                 | `[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]` | see [driver docs](https://github.com/go-sql-driver/mysql) for more information                                        |
| PostgreSQL           | [postgres](https://github.com/jackc/pgx)                  | pgx             | `[user[:password]@][netloc][:port][,...][/dbname][?param1=value1&...]`                 | see [postgres docs](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) for more information |
| SQLite               | [sqlite](https://gitlab.com/cznic/sqlite)                 |                 | `filename`                                                                             | see [driver docu](https://pkg.go.dev/modernc.org/sqlite) for more information                                         |
| TiDB                 | [tidb](https://github.com/go-sql-driver/mysql)            | mysql           | see _mysql_ driver                                                                     | uses MySQL driver                                                                                                     |
| ClickHouse           | [clickhouse](https://github.com/ClickHouse/clickhouse-go) |                 | `tcp://host:port[?param1=value&...&paramN=value]"`                                     | see [clickhouse-go docs](https://github.com/ClickHouse/clickhouse-go#dsn) for more information                        |

## Comments

### Driver aliases

Some database drivers are supported though another driver (e.g. CockroachDB). For other databases we provide a more
obvious name (e.g. postgres) compared to the driver name. For all of those drivers you might use an _alias_ name
during configuration.

### Example data-source-name DSN

The given examples are just that, so please check the driver documentation for the exact format
and available options and parameters. Please note that the format of a DSN might also change
between driver version.

### Type conversions

Telegraf relies on type conversion of the database driver and/or the golang sql framework. In case you find
any problem, please open an issue!

## Help

If nothing seems to work, you might find help in the telegraf forum or in the chat.

### The documentation is wrong

Please open an issue or even better send a pull-request!

### I found a bug

Please open an issue or even better send a pull-request!

### My database is not supported

We currently cannot support CGO drivers in telegraf! Please check if a **pure Go** driver for the [golang sql framework](https://golang.org/pkg/database/sql/) exists.
If you found such a driver, please let us know by opening an issue or even better by sending a pull-request!
