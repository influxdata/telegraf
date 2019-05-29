# PostgreSQL Output Plugin

This output plugin writes all metrics to PostgreSQL.

### Configuration:

```toml
# Send metrics to postgres
[[outputs.postgresql]]
  address = "host=localhost user=postgres sslmode=verify-full"

  ## Store tags as foreign keys in the metrics table. Default is false.
  # tags_as_foreignkeys = false

  ## Template to use for generating tables
  ## Available Variables:
  ##   {TABLE} - tablename as identifier
  ##   {TABLELITERAL} - tablename as string literal
  ##   {COLUMNS} - column definitions
  ##   {KEY_COLUMNS} - comma-separated list of key columns (time + tags)

  ## Default template
  # table_template = "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})"
  ## Example for timescaledb
  # table_template = "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS}); SELECT create_hypertable({TABLELITERAL},'time',chunk_time_interval := '1 week'::interval, if_not_exists := true);"

  ## Schema to create the tables into
  # schema = "public"

  ## Use jsonb datatype for tags. Default is true.
  # tags_as_jsonb = false

  ## Use jsonb datatype for fields. Default is true.
  # fields_as_jsonb = false

```
