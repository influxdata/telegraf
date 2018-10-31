# SQL plugin

The plugin inserts values to SQL various database.
Supported/integrated drivers are mssql (SQLServer), mysql (MySQL), postgres (Postgres)
Activable drivers (read below) are all golang SQL compliant drivers (see https://github.com/golang/go/wiki/SQLDrivers): for instance oci8 for Oracle or sqlite3 (SQLite)

## Getting started :
First you need to grant insert (if auto create table create) privileges to the database user you use for the connection

## Configuration:

```
# Send metrics to SQL-Database (Example configuration for MySQL/MariaDB)
[[outputs.sql]]
  ## Database Driver, required.
  ## Valid options: mssql (SQLServer), mysql (MySQL), postgres (Postgres), sqlite3 (SQLite), [oci8 ora.v4 (Oracle)]
  driver = "mysql"

  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  address = "username:password@tcp(server:port)/table"

  ## Available Variables:
  ##   {TABLE} - tablename as identifier
  ##   {TABLELITERAL} - tablename as string literal
  ##   {COLUMNS} - column definitions
  ##   {KEY_COLUMNS} - comma-separated list of key columns (time + tags)
  ##

  ## Check with this is table exists
  ##
  ## Template for MySQL is "SELECT 1 FROM {TABLE} LIMIT 1"
  ##
  table_exists_template = "SELECT 1 FROM {TABLE} LIMIT 1"

  ## Template to use for generating tables

  ## Default template
  ##
  # table_template = "CREATE TABLE {TABLE}({COLUMNS})"

  ## Convert Telegraf datatypes to these types
  [[outputs.sql.convert]]
    integer              = "INT"
    real                 = "DOUBLE"
    text                 = "TEXT"
    timestamp            = "TIMESTAMP"
    defaultvalue         = "TEXT"
    unsigned             = "UNSIGNED"
```
sql_script is read only once, if you change the script you need to reload telegraf

## Field names
If database table is not pre-created tries driver to create database. There can be errors as
SQL has strict scheming.

## Tested Databases
Actually I run the plugin using MySQL

## TODO
1) Test with other databases
2) More sane testing
