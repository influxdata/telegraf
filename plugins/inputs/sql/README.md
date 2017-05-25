# SQL plugin

The plugin executes simple queries or query scripts on multiple servers.
It permits to select the tags and the fields to export, if is needed fields can be forced to a choosen datatype. 
Supported/integrated drivers are mssql (SQLServer), mysql (MySQL), postgres (Postgres)
Activable drivers (read below) are all golang SQL compliant drivers (see https://github.com/golang/go/wiki/SQLDrivers): for instance oci8 for Oracle or sqlite3 (SQLite)

## Getting started :
First you need to grant read/select privileges on queried tables to the database user you use for the connection

### Non pure go drivers
For some not pure go drivers you may need external shared libraries and environment variables: look at sql driver implementation site
Actually the dependencies to all those drivers (oracle,db2,sap) are commented in the sql.go source. You can enable it, just remove the comment and perform a 'go get <driver git url>' and recompile telegraf. As alternative you can use the 'golang 1.8 plugins feature'like described here below

### Oracle driver with golang 1.8 plugins feature
Follow the docu in https://github.com/mattn/go-oci8 for build the oci8 driver.
If all is going well now golang oci8 driver is compiled and linked against oracle shared libs. But not linked in telegraf.

For let i use in telegraf, do the following:
create a file plugin.go with this content:

```
package main

import "C"

import (
	"log"
	// .. here you can add import to other drivers
	_ "github.com/mattn/go-oci8" // requires external prorietary libs
	// _ "bitbucket.org/phiggins/db2cli" // requires external prorietary libs
	// _ "github.com/mattn/go-sqlite3" // not compiles on windows
)
func main() {
	log.Printf("I! Loaded plugin of shared libs")
}
``` 
build it with
``` 
mkdir $GOPATH/lib
go build -buildmode=plugin -o $GOPATH/lib/oci8_go.so plugin.go
```
in the input plugin configuration specigy the path of the created shared lib 
``` 
[[inputs.sql]]
	...
	driver = "oci8"
	shared_lib = "/home/luca/.gocode/lib/oci8_go.so"
	...
``` 

The steps of above can be reused in a similar way for other proprietary and non proprietary drivers


## Configuration:

```
	[[inputs.sql]]
		# debug=false						# Enables very verbose output

		## Database Driver
		driver = "mysql" 					# required. Valid options: mssql (SQLServer), mysql (MySQL), postgres (Postgres), sqlite3 (SQLite), [oci8 ora.v4 (Oracle)]
		# shared_lib = "/home/luca/.gocode/lib/oci8_go.so"		# optional: path to the golang 1.8 plugin shared lib
		# keep_connection = false 			# true: keeps the connection with database instead to reconnect at each poll and uses prepared statements (false: reconnection at each poll, no prepared statements)

		## Server DSNs
		servers  = ["readuser:sEcReT@tcp(neteye.wp.lan:3307)/rue", "readuser:sEcReT@tcp(hostmysql.wp.lan:3307)/monitoring"] # required. Connection DSN to pass to the DB driver
		hosts=["neteye", "hostmysql"]	# optional: for each server a relative host entry should be specified and will be added as host tag
		db_names=["rue", "monitoring"]	# optional: for each server a relative db name entry should be specified and will be added as dbname tag

		## Queries to perform (block below can be repeated)
		[[inputs.sql.query]]
			# query has precedence on query_script, if both query and query_script are defined only query is executed
			query="SELECT avg_application_latency,avg_bytes,act_throughput FROM Baselines WHERE application>0"
			# query_script = "/path/to/sql/script.sql" # if query is empty and a valid file is provided, the query will be read from file
			#
			measurement="connection_errors"	# destination measurement
			tag_cols=["application"]		# colums used as tags
			field_cols=["avg_application_latency","avg_bytes","act_throughput"]	# select fields and use the database driver automatic datatype conversion
			#
			# bool_fields=["ON"]			# adds fields and forces his value as bool
			# int_fields=["MEMBERS",".*BYTES"]	# adds fields and forces his value as integer
			# float_fields=["TEMPERATURE"]	# adds fields and forces his value as float
			# time_fields=[".*_TIME"]		# adds fields and forces his value as time
			#
			# field_measurement = "CLASS"		# the golumn that contains the name of the measurement
			# field_host = "DBHOST"				# the column that contains the name of the database host used for host tag value
			# field_database = "DBHOST"			# the column that contains the name of the database used for dbname tag value
			# field_name = "counter_name"		# the column that contains the name of the counter
			# field_value = "counter_value"		# the column that contains the value of the counter
			#
			# field_timestamp = "sample_time"	# the column where is to find the time of sample (should be a date datatype)
			#
			ignore_other_fields = false 	# false: if query returns columns not defined, they are automatically added (true: ignore columns)
			null_as_zero = false			# true: converts null values into zero or empty strings (false: ignore fields)
			sanitize = false				# true: will perform some chars substitutions (false: use value as is)
			ignore_row_errors				# true: if an error in row parse is raised then the row will be skipped and the parse continue on next row (false: fatal error)
```
sql_script is read only once, if you change the script you need to restart telegraf

## Field names
Field names are the same of the relative column name or taken from value of a column. If there is the need of rename the fields, just do it in the sql, try to use an ' AS ' .

## Datatypes:
Using field_cols list the values are converted by the go database driver implementation. 
In some cases this automatic conversion is not what we expect, therefore you can force the destination datatypes specifing the columns in the bool/int/float/time_fields lists, then if possible the plugin converts the data.
All field lists can contain an regex for column name matching.
If an error in conversion occurs then telegraf exits, therefore a --test run is suggested.

## Tested Databases
Actually I run the plugin using oci8,mysql,mssql,postgres,sqlite3


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
1) Implement tests
2) Keep trace of timestamp of last poll for use in the where statement
3) Group by serie if timestamp and measurement are the same within a query for perform single insert in db instead of multiple
4) Give the possibility to define parameters to pass to the prepared statement
5) Get the host and database tag value automatically parsing the connection DSN string
6) Add option for parse tags once and reuse it for all rows in a query
X) Add your needs here .....

## ENJOY
Luca

