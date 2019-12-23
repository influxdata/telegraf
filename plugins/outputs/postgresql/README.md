# PostgreSQL Output Plugin

This output plugin writes all metrics to PostgreSQL. 
The plugin manages the schema automatically updating missing columns, and checking if existing ones are of the proper type. 

**_WARNING_**: In order to enable automatic schema update, the connection to the database must
be established with a user that has sufficient permissions. Either be a admin, or an owner of the 
target schema.


### Configuration:

```toml
# Send metrics to postgres
[[outputs.postgresql]]
    ## specify address via a url:
    ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
    ##       ?sslmode=[disable|verify-ca|verify-full]
    ## or a simple string:
    ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
    ##
    ## All connection parameters are optional. Also supported are PG environment vars
    ## e.g. PGPASSWORD, PGHOST, PGUSER, PGDATABASE 
    ## all supported vars here: https://www.postgresql.org/docs/current/libpq-envars.html
    connection = "host=localhost user=postgres sslmode=verify-full"

    ## Update existing tables to match the incoming metrics. Default is true
    # do_schema_updates = true

    ## Store tags as foreign keys in the metrics table. Default is false.
    # tags_as_foreignkeys = false
  
    ## If tags_as_foreignkeys is set to true you can choose the number of tag sets to cache
    ## per measurement (metric name). Default is 1000, if set to 0 => cache has no limit.
    # cached_tagsets_per_measurement = 1000

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

    ## Use jsonb datatype for tags. Default is false.
    # tags_as_jsonb = false

    ## Use jsonb datatype for fields. Default is false.
    # fields_as_jsonb = false

```
