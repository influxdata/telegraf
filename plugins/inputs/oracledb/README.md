# Oracle plugin

This oracle plugin provides metrics for your oracle database. It has been designed
to parse SQL queries in the plugin section of your `telegraf.conf`. Plugin requires
[Oracle Instant Client Library](https://www.oracle.com/database/technologies/instant-client/downloads.html) at run time.
Either Basic or Basic Light package is sufficient. Library path can be provided via
standard OS environment variables - `LD_LIBRARY_PATH`, `DYLD_LIBRARY_PATH` and `PATH` for
Linux, MacOS and Windows respectively, or via `client_lib_dir` plugin parameter for Windows and MacOS.

### Configuration example
```toml
[[inputs.oracledb]]
  ## Connection string, e.g. easy connect string like 
  #    "host:port/service_name"
  #  or oracle net connect descriptor string like 
  #    (DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))
  connection_string = ""

  ## Database credentials
  username = ""
  password = ""

  ## Role, either SYSDBA, SYSASM, SYSOPER or empty
  role = ""

  ## Path to the Oracle Client library directory, optional.
  # Should be used if there is no LD_LIBRARY_PATH variable 
  # or not possible to confugire it properly.
  client_lib_dir = ""

  ## Define the toml config where the sql queries are stored
  # Structure :
  # [[inputs.oracledb.query]]
  #   sqlquery string
  #   script string
  #   schema string
  #   tag_columns array of strings
  [[inputs.oracledb.query]]
    # Query name, optional. Used in logging.
    name = ""
    # OracleDB sql query
    sqlquery = "SELECT 1 AS \"alive\", 'some_value' as \"some_tag\" FROM dual"
    # The script option can be used to specify the .sql file path.
    # If script and sqlquery options specified at same time, sqlquery will be used.
    script = ""
    # Schema name. If provided, then ALTER SESSION SET CURRENT_SCHEMA query will be executed
    schema = ""
    # Query execution timeout, in seconds.
    timeout = 10
    # Array of column names, which would be stored as tags
    tag_columns = ["some_tag"]
```

### Example Output:
```
$ ./telegraf --config oracledb.conf --input-filter oracledb --test  
2020-12-15T10:49:36Z I! Starting Telegraf 
> oracledb,db_unique_name=test,host=MacBook-Pro.local,instance_name=test,server_host=dbhost,service_name=test,some_tag=some_value alive=1i 1608029380000000000
```