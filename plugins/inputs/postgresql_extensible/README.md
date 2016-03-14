# PostgreSQL plugin

This postgresql plugin provides metrics for your postgres database. It has been designed to parse ithe sql queries in the plugin section of your telegraf.conf.

For now only two queries are specified and it's up to you to add more; some per query parameters have been added :

* The SQl query itself
* The minimum version supported (here in numeric display visible in pg_settings)
* A boolean to define if the query have to be run against some specific variables (defined in the databaes variable of the plugin section)
* The list of the column that have to be defined has tags

```
  # specify address via a url matching:
  #   postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
  # or a simple string:
  #   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  #
  # All connection parameters are optional.  #
  # Without the dbname parameter, the driver will default to a database
  # with the same name as the user. This dbname is just for instantiating a
  # connection with the server and doesn't restrict the databases we are trying
  # to grab metrics for.
  #
  address = "host=localhost user=postgres sslmode=disable"
  # A list of databases to pull metrics about. If not specified, metrics for all
  # databases are gathered.
  # databases = ["app_production", "testing"]
  #
  # Define the toml config where the sql queries are stored
  # New queries can be added, if the withdbname is set to true and there is no databases defined
  # in the 'databases field', the sql query is ended by a 'is not null' in order to make the query
  # succeed.
  # Be careful that the sqlquery must contain the where clause with a part of the filtering, the plugin will
  # add a 'IN (dbname list)' clause if the withdbname is set to true
  # Example :
  # The sqlquery : "SELECT * FROM pg_stat_database where datname" become "SELECT * FROM pg_stat_database where datname IN ('postgres', 'pgbench')"
  # because the databases variable was set to ['postgres', 'pgbench' ] and the withdbname was true.
  # Be careful that if the withdbname is set to false you d'ont have to define the where clause (aka with the dbname)
  # the tagvalue field is used to define custom tags (separated by comas)
  #
  # Structure :
  # [[inputs.postgresql_extensible.query]]
  #   sqlquery string
  #   version string
  #   withdbname boolean
  #   tagvalue string (coma separated)
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_database where datname"
    version=901
    withdbname=false
    tagvalue=""
  [[inputs.postgresql_extensible.query]]
    sqlquery="SELECT * FROM pg_stat_bgwriter"
    version=901
    withdbname=false
    tagvalue=""
```

The system can be easily extended using homemade metrics collection tools or using postgreql extensions ([pg_stat_statements](http://www.postgresql.org/docs/current/static/pgstatstatements.html), [pg_proctab](https://github.com/markwkm/pg_proctab), [powa](http://dalibo.github.io/powa/)...)
