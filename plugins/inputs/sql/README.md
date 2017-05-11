# SQL plugin

The plugin executes simple queries or query scripts on multiple servers.
It permits to select the tags and the fields to export, if is needed fields can be forced to a choosen datatype. 
Supported drivers are  go-mssqldb (sqlserver) , oci8 ora.v4 (Oracle), mysql (MySQL), pq (Postgres) 
```
```

## Getting started :

First you need to grant read/select privileges on queried tables to the database user you use for the connection
For some drivers you need external shared libraries and environment variables (for instance   
```
```


## Configuration:

``` 
	[[inputs.sql]]
		# debug=false						# Enables very verbose output
	
		## Database Driver
		driver = "oci8" 					# required. Valid options: go-mssqldb (sqlserver) , oci8 ora.v4 (Oracle), mysql, pq (Postgres)
		# keep_connection = false 			# true: keeps the connection with database instead to reconnect at each poll and uses prepared statements (false: reconnection at each poll, no prepared statements)
		
		## Server URLs
		servers  = ["telegraf/monitor@10.0.0.5:1521/thesid", "telegraf/monitor@orahost:1521/anothersid"] # required. Connection URL to pass to the DB driver
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
			#bool_fields=["ON"]				# adds fields and forces his value as bool
			#int_fields=["MEMBERS","BYTES"]	# adds fields and forces his value as integer
			#float_fields=["TEMPERATURE"]	# adds fields and forces his value as float
			#time_fields=["FIRST_TIME"]		# adds fields and forces his value as time
			
			ignore_other_fields = false 	# false: if query returns columns not defined, they are automatically added (true: ignore columns)
			null_as_zero = false			# true: Push null results as zeros/empty strings (false: ignore fields)
			sanitize = false				# true: will perform some chars substitutions (false: use value as is)


```


## Datatypes:


## Example for collect multiple counters defined as COLUMNS in a table:


## Example for collect multiple counters defined as ROWS in a table:


