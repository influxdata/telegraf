# SQL plugin

The plugin executes simple queries or query scripts on multiple servers.
It permits to select the tags and the fields to export, if is needed fields can be forced to a choosen datatype. 
Supported drivers are  go-mssqldb (sqlserver) , oci8 ora.v4 (Oracle), mysql (MySQL), pq (Postgres) 
```
```

## Getting started :
First you need to grant read/select privileges on queried tables to the database user you use for the connection

### Non pure go drivers
For some not pure go drivers you may need external shared libraries and environment variables: look at sql driver implementation site 
For instance using oracle driver on rh linux you need to install oracle-instantclient12.2-basic-12.2.0.1.0-1.x86_64.rpm package and set 
```
export ORACLE_HOME=/usr/lib/oracle/12.2/client64
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$ORACLE_HOME/lib 
```
Actually the dependencies to all those drivers (oracle,db2,sap) are commented in the sql.go source. You can enable it, just remove the comment and perform a 'go get <driver git url>' and recompile telegraf


## Configuration:

``` 

	[[inputs.sql]]
		# debug=false						# Enables very verbose output
	
		## Database Driver
		driver = "oci8" 					# required. Valid options: go-mssqldb (sqlserver) , oci8 | ora.v4 (Oracle), mysql, postgres
		# keep_connection = false 			# true: keeps the connection with database instead to reconnect at each poll and uses prepared statements (false: reconnection at each poll, no prepared statements)
		
		## Server DSNs
		servers  = ["telegraf/monitor@10.0.0.5:1521/thesid", "telegraf/monitor@orahost:1521/anothersid"] # required. Connection DSN to pass to the DB driver
		hosts=["oraserver1", "oraserver2"]	# for each server a relative host entry should be specified and will be added as host tag
	
		## Queries to perform (block below can be repeated)
		[[inputs.sql.query]]
			# query has precedence on query_script, if both query and query_script are defined only query is executed
			query="select GROUP#,MEMBERS,STATUS,FIRST_TIME,FIRST_CHANGE#,BYTES,ARCHIVED from v$log"  
			# query_script = "/path/to/sql/script.sql" # if query is empty and a valid file is provided, the query will be read from file
			#
			measurement="log"				# destination measurement
			tag_cols=["GROUP#","NAME"]		# colums used as tags
			field_cols=["UNIT"]				# select fields and use the database driver automatic datatype conversion
			#
			# bool_fields=["ON"]				# adds fields and forces his value as bool
			# int_fields=["MEMBERS","BYTES"]	# adds fields and forces his value as integer
			# float_fields=["TEMPERATURE"]	# adds fields and forces his value as float
			# time_fields=["FIRST_TIME"]		# adds fields and forces his value as time
			#
			# field_name = "counter_name"		# the column that contains the name of the counter
			# field_value = "counter_value"		# the column that contains the value of the counter
			#
			# field_timestamp = "sample_time"	# the column where is to find the time of sample (should be a date datatype)
			
			ignore_other_fields = false 	# false: if query returns columns not defined, they are automatically added (true: ignore columns)
			null_as_zero = false			# true: converts null values into zero or empty strings (false: ignore fields)
			sanitize = false				# true: will perform some chars substitutions (false: use value as is)


```
sql_script is read only once, if you change the script you need to restart telegraf

## Field names
Field names are the same of the relative column name or taken from value of a column. If there is the need of rename the fields, just do it in the sql, try to use an ' AS ' .

## Datatypes:
Using field_cols list the values are converted by the go database driver implementation. 
In some cases this automatic conversion is not what we wxpect, therefore you can force the destination datatypes specifing the columns in the bool/int/float/time_fields lists, then if possible the plugin converts the data.
If an error in conversion occurs then telegraf exits, therefore a --test run is suggested.

## Tested Databases
Actually I run the plugin using oci8,mysql and mssql
The mechanism for get the timestamp from a table column has known problems

## Example for collect multiple counters defined as COLUMNS in a table (vertical counter structure):
Here we read a table where each counter is on a different row. Each row contains a column with the name of the counter (counter_name) and a column with his value (cntr_value) and some other columns that we use as tags  (instance_name,object_name)

###Config
```
[[inputs.sql]]
	interval = "60s"
	driver = "mssql"
	servers = [
		"Server=mssqlserver1.my.lan;Port=1433;User Id=telegraf;Password=secret;app name=telegraf"
		"Server=mssqlserver2.my.lan;Port=1433;User Id=telegraf;Password=secret;app name=telegraf"
	]
	hosts=["mssqlserver_cluster_1","mssqlserver_cluster_2"]

	[[inputs.sql.query]]
		measurement = "os_performance_counters"
		ignore_other_fields=true
		sanitize=true
		query="SELECT * FROM sys.dm_os_performance_counters WHERE object_name NOT LIKE '%Deprecated%' ORDER BY counter_name"
		tag_cols=["instance_name","object_name"]
		field_name = "counter_name"
		field_value = "cntr_value"
```
### Result:
```
> os_performance_counters,host=mssqlserver_cluster_1,object_name=MSSQL$TESTSQL2014:Broker_Statistics Activation_Errors_Total=0i 1494496261000000000
> os_performance_counters,host=mssqlserver_cluster_1,object_name=MSSQL$TESTSQL2014:Cursor_Manager_by_Type,instance_name=TSQL_Local_Cursor Active_cursors=0i 1494496261000000000
> os_performance_counters,instance_name=TSQL_Global_Cursor,host=mssqlserver_cluster_1,object_name=MSSQL$TESTSQL2014:Cursor_Manager_by_Type Active_cursors=0i 1494496261000000000
> os_performance_counters,host=mssqlserver_cluster_1,object_name=MSSQL$TESTSQL2014:Cursor_Manager_by_Type,instance_name=API_Cursor Active_cursors=0i 1494496261000000000
> os_performance_counters,host=mssqlserver_cluster_1,object_name=MSSQL$TESTSQL2014:Cursor_Manager_by_Type,instance_name=_Total Active_cursors=0i 1494496261000000000
...

```
## Example for collect multiple counters defined as ROWS in a table (horizontal counter structure):
Here we read multiple counters defined on same row where the counter name is the name of his column.
In this example we force some counters datatypes: "MEMBERS","FIRST_CHANGE#" as integer, "BYTES" as float, "FIRST_TIME" as time. The field "UNIT" is used with the automatic driver datatype conversion.
The column "ARCHIVED" is ignored

###Config
```
[[inputs.sql]]
	interval = "20s"

	driver = "oci8"
	keep_connection=true
	servers  = ["telegraf/monitor@10.62.6.1:1522/tunapit"]
	hosts=["oraclehost.my.lan"]
	## Queries to perform
	[[inputs.sql.query]]
		query="select GROUP#,MEMBERS,STATUS,FIRST_TIME,FIRST_CHANGE#,BYTES,ARCHIVED from v$log"
		measurement="log"
		tag_cols=["GROUP#","STATUS","NAME"]
		field_cols=["UNIT"]
		int_fields=["MEMBERS","FIRST_CHANGE#"]
		float_fields=["BYTES"]
		time_fields=["FIRST_TIME"]
		ignore_other_fields=true
```
### Result:
```
> log,host=pbzasplx001.wp.lan,GROUP#=1,STATUS=INACTIVE MEMBERS=1i,FIRST_TIME="2017-05-10 22:08:38 +0200 CEST",FIRST_CHANGE#=368234811i,BYTES=52428800 1494496874000000000
> log,host=pbzasplx001.wp.lan,GROUP#=2,STATUS=CURRENT MEMBERS=1i,FIRST_TIME="2017-05-10 22:08:38 +0200 CEST",FIRST_CHANGE#=368234816i,BYTES=52428800 1494496874000000000
> log,host=pbzasplx001.wp.lan,GROUP#=3,STATUS=INACTIVE MEMBERS=1i,FIRST_TIME="2017-05-10 16:00:55 +0200 CEST",FIRST_CHANGE#=368220858i,BYTES=52428800 1494496874000000000


```

## TODO
Give the possibility to define parameters to pass to the prepared statement
Get the host tag value automatically parsing the connection DSN string
Implement tests

