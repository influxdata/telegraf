# MySQL Input Plugin

This plugin gathers the statistic data from MySQL server

* Global statuses
* Global variables
* Slave statuses
* Binlog size
* Process list
* User Statistics
* Info schema auto increment columns
* InnoDB metrics
* Table I/O waits
* Index I/O waits
* Perf Schema table lock waits
* Perf Schema event waits
* Perf Schema events statements
* File events statistics
* Table schema statistics

### Configuration

```toml
# Read metrics from one or many mysql servers
[[inputs.mysql]]
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    servers = ["user:passwd@tcp(127.0.0.1:3306)/?tls=false"]
  ##    servers = ["user@tcp(127.0.0.1:3306)/?tls=false"]
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
  ## gather thread state counts from INFORMATION_SCHEMA.USER_STATISTICS
  gather_user_statistics                    = true
  #
  ## gather auto_increment columns and max values from information schema
  gather_info_schema_auto_inc               = true
  #
  ## gather metrics from INFORMATION_SCHEMA.INNODB_METRICS
  gather_innodb_metrics                     = true
  #
  ## gather metrics from SHOW SLAVE STATUS command output
  gather_slave_status                       = true
  #
  ## gather metrics from SHOW BINARY LOGS command output
  gather_binary_logs                        = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_TABLE
  gather_table_io_waits                     = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS
  gather_table_lock_waits                   = false
  #
  ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_INDEX_USAGE
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

  ## Optional TLS Config (will be used if tls=custom parameter specified in server uri)
  tls_ca = "/etc/telegraf/ca.pem"
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"
```

#### Metric Version

When `metric_version = 2`, a variety of field type issues are corrected as well
as naming inconsistencies.  If you have existing data on the original version
enabling this feature will cause a `field type error` when inserted into
InfluxDB due to the change of types.  For this reason, you should keep the
`metric_version` unset until you are ready to migrate to the new format.

If preserving your old data is not required you may wish to drop conflicting
measurements:
```
DROP SERIES from mysql
DROP SERIES from mysql_variables
DROP SERIES from mysql_innodb
```

Otherwise, migration can be performed using the following steps:

1. Duplicate your `mysql` plugin configuration and add a `name_suffix` and
`metric_version = 2`, this will result in collection using both the old and new
style concurrently:
   ```toml
   [[inputs.mysql]]
     servers = ["tcp(127.0.0.1:3306)/"]

   [[inputs.mysql]]
     name_suffix = "_v2"
     metric_version = 2

     servers = ["tcp(127.0.0.1:3306)/"]
   ```

2. Upgrade all affected Telegraf clients to version >=1.6.

   New measurements will be created with the `name_suffix`, for example::
   - `mysql_v2`
   - `mysql_variables_v2`

3. Update charts, alerts, and other supporting code to the new format.
4. You can now remove the old `mysql` plugin configuration and remove old
   measurements.

If you wish to remove the `name_suffix` you may use Kapacitor to copy the
historical data to the default name.  Do this only after retiring the old
measurement name.

1. Use the techinique described above to write to multiple locations:
   ```toml
   [[inputs.mysql]]
     servers = ["tcp(127.0.0.1:3306)/"]
     metric_version = 2

   [[inputs.mysql]]
     name_suffix = "_v2"
     metric_version = 2

     servers = ["tcp(127.0.0.1:3306)/"]
   ```
2. Create a TICKScript to copy the historical data:
   ```
   dbrp "telegraf"."autogen"

   batch
       |query('''
           SELECT * FROM "telegraf"."autogen"."mysql_v2"
       ''')
           .period(5m)
           .every(5m)
           |influxDBOut()
                   .database('telegraf')
                   .retentionPolicy('autogen')
                   .measurement('mysql')
   ```
3. Define a task for your script:
   ```sh
   kapacitor define copy-measurement -tick copy-measurement.task
   ```
4. Run the task over the data you would like to migrate:
   ```sh
   kapacitor replay-live batch -start 2018-03-30T20:00:00Z -stop 2018-04-01T12:00:00Z -rec-time -task copy-measurement
   ```
5. Verify copied data and repeat for other measurements.

### Metrics:
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
* User Statistics - connection metrics from user statistics for each user. It has the following fields
    * access_denied
    * binlog_bytes_written
    * busy_time
    * bytes_received
    * bytes_sent
    * commit_transactions
    * concurrent_connections
    * connected_time
    * cpu_time
    * denied_connections
    * empty_queries
    * hostlost_connections
    * other_commands
    * rollback_transactions
    * rows_fetched
    * rows_updated
    * select_commands
    * server
    * table_rows_read
    * total_connections
    * total_ssl_connections
    * update_commands
    * user
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
* InnoDB metrics - all metrics of information_schema.INNODB_METRICS with a status "enabled"
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
* Perf events statements - gathers attributes of each event
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
* User Statistics measurement has following tags
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


#### mysql_metadatalock_session

> 元锁会话明细

| mysql_metadatalock_session | key         | 数据类型 | 说明                    |
| -------------------------- | ----------- | -------- | ----------------------- |
| `Tags`                     | `server`    | `string` | `数据库url地址或主机名` |
| `Fields`                   | `id`        | `int64`  | `会话id`                |
|                            | `user`      | `string` | `会话的登录用户名`      |
|                            | `host`      | `string` | `会话的来源地址`        |
|                            | `db`        | `string` | `会话访问的数据库名`    |
|                            | `command`   | `string` | `会话执行的语句类型`    |
|                            | `conn_time` | `int64`  | `会话持续时间`          |
|                            | `state`     | `string` | `会话状态`              |
|                            | `info`      | `string` | `会话执行的具体SQL语句` |

#### mysql_metadatalock_count

> 元锁个数

| mysql_metadatalock_count | key      | 数据类型 | 说明                    |
| ------------------------ | -------- | -------- | ----------------------- |
| `Tags`                   | `server` | `string` | `数据库url地址或主机名` |
| `Fields`                 | `count`  | `int64`  | `会话总数`              |

#### mysql_metadatalock_trx_id

>  导致元锁冲突的长时间未提交的事务会话id

| mysql_metadatalock_trx_id | key      | 数据类型 | 说明                    |
| ------------------------- | -------- | -------- | ----------------------- |
| `Tags`                    | `server` | `string` | `数据库url地址或主机名` |
| `Fields`                  | `id`     | `int64`  | `会话id`                |



#### mysql_innodb_blocking_trx_id

>  获取到innodb事务锁冲突的会话明细,以及一共阻塞了多少事务，用于快速解决行锁冲突

| mysql_innodb_blocking_trx_id | key                          | 数据类型 | 说明                                                         |
| ---------------------------- | ---------------------------- | -------- | ------------------------------------------------------------ |
| `Tags`                       | `server`                     | `string` | `数据库url地址或主机名`                                      |
| `Fields`                     | `id`                         | `int64`  | `会话id`                                                     |
|                              | `user`                       | `string` | `会话的登录用户名`                                           |
|                              | `host`                       | `string` | `会话的来源地址`                                             |
|                              | `db`                         | `string` | `会话访问的数据库名`                                         |
|                              | `command`                    | `string` | `会话执行的语句类型`                                         |
|                              | `time`                       | `int64`  | `会话持续时间`                                               |
|                              | `state`                      | `string` | `会话状态`                                                   |
|                              | `info`                       | `string` | `会话执行的具体SQL语句`                                      |
|                              | `trx_id`                     | `int64`  | `事务id`                                                     |
|                              | `trx_state string`           | `string` | `事务状态`                                                   |
|                              | `trx_started`                | `string` | `事务开始时间`                                               |
|                              | `trx_requested_lock_id`      | `string` | `等待事务的锁id`                                             |
|                              | `trx_wait_started`           | `string` | `事务等待开始的事件`                                         |
|                              | `trx_weight`                 | `int64`  | `事务的权重`                                                 |
|                              | `trx_mysql_thread_id`        | `int64`  | `事务线程id`                                                 |
|                              | `trx_query`                  | `string` | `事务运行的SQL`                                              |
|                              | `trx_operation_state`        | `string` | `事务的操作轧辊台`                                           |
|                              | `trx_tables_in_use`          | `int64`  | `事务使用的表`                                               |
|                              | `trx_tables_locked`          | `int64`  | `被锁住的表`                                                 |
|                              | `trx_lock_structs`           | `int64`  | `事务保留的锁`                                               |
|                              | `trx_lock_memory_bytes`      | `int64`  | `事务锁定的内存大小`                                         |
|                              | `trx_rows_locked`            | `int64`  | `事物锁定的最大行树`                                         |
|                              | `trx_rows_modified`          | `int64`  | `事务修改的行数`                                             |
|                              | `trx_concurrency_tickets`    | `int64`  | ``                                                           |
|                              | `trx_isolation_level`        | `string` | `事务隔离级别`                                               |
|                              | `trx_unique_checks`          | `int64`  | `事务的唯一键检查是打开还是关闭`                             |
|                              | `trx_foreign_key_checks`     | `int64`  | `事务的外键检查是否开启`                                     |
|                              | `trx_last_foreign_key_error` | `string` | `事务最近一次外键错误`                                       |
|                              | `trx_adaptive_hash_latched`  | `int64`  | `自适应哈希索引是否被当前事务锁定`                           |
|                              | `trx_adaptive_hash_timeout`  | `int64`  | ``                                                           |
|                              | `trx_is_read_only`           | `int64`  | `1表示事务是只读的`                                          |
|                              | `trx_autocommit_non_locking` | `int64`  | `值1表示事务是不使用for update或lock in shared mode子句的select语句，并且在启用autocommit设置的情况下执行，因此事务将只包含此语句。（5.6.4及更高版本。）当此列和trx_均为只读时，innodb会优化事务，以减少与更改表数据的事务相关的开销。` |
|                              | ` countnum`                  | `int64`  | `该事务阻塞了多少其他事务`                                   |

#### mysql_innodb_lock_waits

> ​	获取到innodb事务锁冲突锁信息,用于分析行所锁原因

| mysql_innodb_lock_waits | key           | 数据类型 | 说明                    |
| ----------------------- | ------------- | -------- | ----------------------- |
| `Tags`                  | `server`      | `string` | `数据库url地址或主机名` |
| `Fields`                | `lock_id`     | `string` | `锁id`                  |
|                         | `lock_trx_id` | `int64`  | `事务id`                |
|                         | `lock_mode`  | `string` | `锁的模式`              |
|                         | `lock_type`  | `string` | `锁的类型`              |
|                         | `lock_table` | `string` | `申请锁的表`            |
|                         | `lock_index` | `string` | `锁住的索引`            |
|                         | `lock_space` | `int64`  | `锁对象的space id`      |
|                         | `lock_page`  | `int64`  | `事务锁定页的数量`      |
|                         | `lock_rec`   | `int64`  | `事务锁定行的数量`      |
|                         | `lock_data`  | `int64`  | `事务锁定记录的主键值`  |

#### mysql_innodb_locks_counts

> 仅获取到innodb事务锁冲突的会话id，会话持续时间，以及一共阻塞了多少事务，用于告警使用

| mysql_innodb_locks_counts | key                          | 数据类型 | 说明                                                         |
| ---------------------------- | ---------------------------- | -------- | ------------------------------------------------------------ |
| `Tags`                       | `server`                     | `string` | `数据库url地址或主机名`                                      |
| `Fields`                     | `id`                    | `int64` | `会话id`                                          |
|                              | `time`                    | `int64` | `会话持续时间`                                       |
|                              | `countnum`                  | `int64` | `该事务阻塞事务数量`                      |

#### mysql_innodb_status

> 获取到死锁信息,用于分析行死锁原因

| mysql_innodb_status | key      | 数据类型 | 说明                    |
| ------------------- | -------- | -------- | ----------------------- |
| `Tags`              | `server` | `string` | `数据库url地址或主机名` |
| `Fields`            | `type`   | `string` | `engine类型`            |
|                     | `name`   | `string` | `engine别名`            |
|                     | `innodb` | `string` | `具体状态`              |

#### mysql_dead_lock_rows

> 判断当前是否存在死锁，用于告警

| mysql_dead_lock_rows | key        | 数据类型 | 说明                          |
| -------------------- | ---------- | -------- | ----------------------------- |
| `Tags`               | `server`   | `string` | `数据库url地址或主机名`       |
| `Fields`             | `deadlock` | `int64`  | `0代表没有死锁 1代表存在死锁` |