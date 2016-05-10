# MySQL Input plugin

This plugin gathers the statistic data from MySQL server

* Global statuses
* Global variables
* Slave statuses
* Binlog size
* Process list
* Info schema auto increment columns
* Table I/O waits
* Index I/O waits
* Perf Schema table lock waits
* Perf Schema event waits
* Perf Schema events statements
* File events statistics
* Table schema statistics

## Configuration

```
# Read metrics from one or many mysql servers
[[inputs.mysql]]
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    db_user:passwd@tcp(127.0.0.1:3306)/?tls=false
  ##    db_user@tcp(127.0.0.1:3306)/?tls=false
  #
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:3306)/"]
  ## the limits for metrics form perf_events_statements
  perf_events_statements_digest_text_limit  = 120
  perf_events_statements_limit              = 250
  perf_events_statements_time_limit         = 86400
  #
  ## if the list is empty, then metrics are gathered from all database tables
  table_schema_databases                    = []
  #
  ## gather metrics from INFORMATION_SCHEMA.TABLES for databases provided above list
  gather_table_schema                       = false
  #
  ## gather thread state counts from INFORMATION_SCHEMA.PROCESSLIST
  gather_process_list                       = true
  #
  ## gather auto_increment columns and max values from information schema
  gather_info_schema_auto_inc               = true
  #
  ## gather metrics from SHOW SLAVE STATUS command output
  gather_slave_status                       = true
  #
  ## gather metrics from SHOW BINARY LOGS command output
  gather_binary_logs                        = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMART_BY_TABLE
  gather_table_io_waits                     = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS
  gather_table_lock_waits                   = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMART_BY_INDEX_USAGE
  gather_index_io_waits                     = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.EVENT_WAITS
  gather_event_waits                        = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.FILE_SUMMARY_BY_EVENT_NAME
  gather_file_events_stats                  = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.EVENTS_STATEMENTS_SUMMARY_BY_DIGEST
  gather_perf_events_statements             = false
  #
  ## Some queries we may want to run less often (such as SHOW GLOBAL VARIABLES)
  interval_slow                             = "30m"
```

## Measurements & Fields
* Global statuses - all numeric and boolean values of `SHOW GLOBAL STATUSES`
* Global variables - all numeric and boolean values of `SHOW GLOBAL VARIABLES`
* Slave status - metrics from `SHOW SLAVE STATUS` the metrics are gathered when
the single-source replication is on. If the multi-source replication is set,
then everything works differently, this metric does not work with multi-source
replication.
    * slave_[column name]()
* Binary logs - all metrics including size and count of all binary files.
Requires to be turned on in configuration.
    * binary_size_bytes(int, number)
    * binary_files_count(int, number)
* Process list - connection metrics from processlist for each user. It has the following tags
    * connections(int, number)
* Perf Table IO waits - total count and time of I/O waits event for each table
and process. It has following fields:
    * table_io_waits_total_fetch(float, number)
    * table_io_waits_total_insert(float, number)
    * table_io_waits_total_update(float, number)
    * table_io_waits_total_delete(float, number)
    * table_io_waits_seconds_total_fetch(float, milliseconds)
    * table_io_waits_seconds_total_insert(float, milliseconds)
    * table_io_waits_seconds_total_update(float, milliseconds)
    * table_io_waits_seconds_total_delete(float, milliseconds)
* Perf index IO waits - total count and time of I/O waits event for each index
and process. It has following fields:
    * index_io_waits_total_fetch(float, number)
    * index_io_waits_seconds_total_fetch(float, milliseconds)
    * index_io_waits_total_insert(float, number)
    * index_io_waits_total_update(float, number)
    * index_io_waits_total_delete(float, number)
    * index_io_waits_seconds_total_insert(float, milliseconds)
    * index_io_waits_seconds_total_update(float, milliseconds)
    * index_io_waits_seconds_total_delete(float, milliseconds)
* Info schema autoincrement statuses - autoincrement fields and max values
for them. It has following fields:
    * auto_increment_column(int, number)
    * auto_increment_column_max(int, number)
* Perf table lock waits - gathers total number and time for SQL and external
lock waits events for each table and operation. It has following fields.
The unit of fields varies by the tags.
    * read_normal(float, number/milliseconds)
    * read_with_shared_locks(float, number/milliseconds)
    * read_high_priority(float, number/milliseconds)
    * read_no_insert(float, number/milliseconds)
    * write_normal(float, number/milliseconds)
    * write_allow_write(float, number/milliseconds)
    * write_concurrent_insert(float, number/milliseconds)
    * write_low_priority(float, number/milliseconds)
    * read(float, number/milliseconds)
    * write(float, number/milliseconds)
* Perf events waits - gathers total time and number of event waits
    * events_waits_total(float, number)
    * events_waits_seconds_total(float, milliseconds)
* Perf file events statuses - gathers file events statuses
    * file_events_total(float,number)
    * file_events_seconds_total(float, milliseconds)
    * file_events_bytes_total(float, bytes)
* Perf file events statements - gathers attributes of each event
    * events_statements_total(float, number)
    * events_statements_seconds_total(float, millieconds)
    * events_statements_errors_total(float, number)
    * events_statements_warnings_total(float, number)
    * events_statements_rows_affected_total(float, number)
    * events_statements_rows_sent_total(float, number)
    * events_statements_rows_examined_total(float, number)
    * events_statements_tmp_tables_total(float, number)
    * events_statements_tmp_disk_tables_total(float, number)
    * events_statements_sort_merge_passes_totales(float, number)
    * events_statements_sort_rows_total(float, number)
    * events_statements_no_index_used_total(float, number)
* Table schema - gathers statistics of each schema. It has following measurements
    * info_schema_table_rows(float, number)
    * info_schema_table_size_data_length(float, number)
    * info_schema_table_size_index_length(float, number)
    * info_schema_table_size_data_free(float, number)
    * info_schema_table_version(float, number)

## Tags
* All measurements has following tags
    * server (the host name from which the metrics are gathered)
* Process list measurement has following tags
    * user (username for whom the metrics are gathered)
* Perf table IO waits measurement has following tags
    * schema
    * name (object name for event or process)
* Perf index IO waits has following tags
    * schema
    * name
    * index
* Info schema autoincrement statuses has following tags
    * schema
    * table
    * column
* Perf table lock waits has following tags
    * schema
    * table
    * sql_lock_waits_total(fields including this tag have numeric unit)
    * external_lock_waits_total(fields including this tag have numeric unit)
    * sql_lock_waits_seconds_total(fields including this tag have millisecond unit)
    * external_lock_waits_seconds_total(fields including this tag have millisecond unit)
* Perf events statements has following tags
    * event_name
* Perf file events statuses has following tags
    * event_name
    * mode
* Perf file events statements has following tags
    * schema
    * digest
    * digest_text
* Table schema has following tags
    * schema
    * table
    * component
    * type
    * engine
    * row_format
    * create_options
