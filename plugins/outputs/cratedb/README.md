# CrateDB Output Plugin for Telegraf

This plugin writes to [CrateDB](https://crate.io/) via its [PostgreSQL protocol](https://crate.io/docs/crate/reference/protocols/postgres.html).

## Table Schema

The plugin requires a table with the following schema.


```sql
CREATE TABLE my_metrics (
  "hash_id" LONG INDEX OFF,
  "timestamp" TIMESTAMP,
  "name" STRING,
  "tags" OBJECT(DYNAMIC),
  "fields" OBJECT(DYNAMIC),
  PRIMARY KEY ("timestamp", "hash_id","day")
) PARTITIONED BY("day");
```

The plugin can create this table for you automatically via the `table_create`
config option, see below.

## Configuration

```toml
# Configuration for CrateDB to send metrics to.
[[outputs.cratedb]]
  # A github.com/jackc/pgx connection string.
  # See https://godoc.org/github.com/jackc/pgx#ParseDSN
  url = "postgres://user:password@localhost/schema?sslmode=disable"
  # Timeout for all CrateDB queries.
  timeout = "5s"
  # Name of the table to store metrics in.
  table = "metrics"
  # If true, and the metrics table does not exist, create it automatically.
  table_create = true
```
