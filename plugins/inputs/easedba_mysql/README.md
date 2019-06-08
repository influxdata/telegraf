

- [Mysql Monitoring Scheam](#mysql-monitoring-scheam)
  - [1 Indices](#1-indices)
    - [1.1 Document json schema](#11-document-json-schema)
  - [2 global tags](#2-global-tags)
  - [3 Metric fields](#3-metric-fields)
    - [3.1 Throughtput Index](#31-throughtput-index)
    - [3.2 Connection Index](#32-connection-index)
    - [3.3 innodb Index](#33-innodb-index)
    - [3.4 disk index](#34-disk-index)
    - [3.5 replication index](#35-replication-index)
    - [3.6 snapshot index](#36-snapshot-index)
    - [3.7 network index](#37-network-index)
    - [3.8 disk index](#38-disk-index)
    - [3.9 cpu index](#39-cpu-index)
    - [3.10 Mem index](#310-mem-index)



# Mysql Monitoring Scheam


> ATTENTION: For convenience, all fields in the document are reserve the same form with DataBase raw resultset, but in the Elasticsearch, all fields are converted to snake case to save.




## 1 Indices

| Index mapping template               | Index pattern                                    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| :----------------------------------- | :----------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| easedba-monitor-metrics-\*           | easedba-monitor-metrics-\*-YYYY.MM.DD*           | Saves time series based metrics of monitored object from different categories. The metrics from different monitored object will be saved into a dedicated document type.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| easedba-monitor-aggregate-metrics-\* | easedba-monitor-aggregate-metrics-\*-YYYY.MM.DD* | Saves calculated performance statistics from different dimensions monitoring requirement needed. The statistics from different dimensions will be saved into a dedicated document type. Due to the statistic calculation are executed on these input metrics directly as streaming and the results will be saved into this index in advance, so the statistics can be loaded and used without any further aggregation（e.g. grouping and computing). This will definitely help the performance of ad-hoc query on the fine-grained metrics ES stored, especially on a large metrics data volume. This index was  designed only to save these statistics ones can be calculated by a simple (fast) and fixed (can be implemented on product design stage instead of runtime stage) functions. |
| easedba-monitor-logs-\*              | easedba-monitor-logs-\*-YYYY.MM.DD*              | Saves the logs outputted from OS, middleware and application. The different logs will be saved into a dedicated document type.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
### 1.1 Document json schema


The data collected from the agent were saved as documents in the Elasticsearch. According to current technical selection(telegraf), in order to decrease the modification of agent, we make the documents JSON schema consistent with the [telegraf metrics model](!https://docs.influxdata.com/telegraf/v1.10/concepts/metrics/). We will try our best to leverage this model to organize our data.

In our scene, we have two part of monitoring data. One is base configuration and another is monitor metrics. The configuration we described as **global tags**, the metrics are **metrics fields**.

So,  the data sent by the agent must follow JSON formation.

```
{
    "fields":{
      METRIC FIELDS 
      ...
    },
    "name":"mysql",
    "tags":{
      GLOBAL TAGS
      ...
    },
    "timestamp":1559118730
}
```

The follow chapters will describe details inforamtion about global tags and  metric fields.



## 2 global tags

| Field       |  Type  | Indexed? | Analyzed? | Required? | Description                                                                                                                                                                            |
| :---------- | :----: | :------: | :-------: | :-------: | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| timestamp   |  date  |   true   |   false   |   true    | The timestamp of current document saved into ES, collectd or SM provided timestamp should be applied to this field, otherwise uses ES generated one. This field uses **UTC** timezone. |
| category    | string |   true   |   false   |   true    | In our case, the value of this field is ``infrastructure``, ``platform`` or ``application``.                                                                                           |
| host_name   | string |   true   |   false   |   true    | The name of host original data collected from. For example, it could be used to indicates a particular application instance.                                                           |
| host_ipv4   | string |   true   |   false   |   true    | The IPv4 address of the host original data collected from. For instance, it could be used to indicates a deployed application instance.                                                |
| system      | string |   true   |   false   |   true    | The name of monitored system.  The busisness domain name is recommended for mysql instance, e.g. CRM, ORDER etc.                                                                       |
| db_instance | string |   true   |   false   |   true    | A name to indicate a single db instace. A `system` (CRM) may contains 1 writing `db_instance` and 2 read-only `db_instance`                                                            |



## 3 Metric fields

### 3.1 Throughtput Index

* Index mapping template: `easedba-monitor-metrics-mysql-throughput`
* Category: ``platform``
* SQL : `show global status`

| Field              |  Type   | Indexed? | Analyzed? | doc_values | Required? | Unit | Description |
| :----------------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| Com_insert         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_select         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_insert_select  | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_replace        | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_replace_select | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_update         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_update_multi   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_delete         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_delete_multi   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_commit         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_rollback       | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_stmt_exexute   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Com_call_procedure | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Slow_sql_count     | integer |    no    |    yes    |     no     | yes       | hnum |             |


### 3.2 Connection Index
* Index mapping template: `easedba-monitor-metrics-mysql-connection`
* Category: ``platform``
* SQL : `show global status`

| Field             |  Type   | Indexed? | Analyzed? | doc_values | Required? | Unit | Description |
| :---------------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| Threads_connected | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Aborted_clients   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Aborted_connects  | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Locked_connects   | integer |    no    |    no     |    yes     | yes       | hnum |             |



### 3.3 innodb  Index
* Index mapping template: `easedba-monitor-metrics-mysql-innodb`
* Category: ``platform``
* SQL : `show global status`
* telegraf should calcualte ratio

| Field                             |  Type   | Indexed? | Analyzed? | doc_values | Required? | Unit | Description |
| :-------------------------------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| Innodb_rows_read                  | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_read_ratio            | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_deleted               | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_deleted_ratio         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_inserted              | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_inserted_ratio        | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_updated               | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_rows_updated_ratio         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_buffer_pool_reads          | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_buffer_pool_read_requests  | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_buffer_pool_write_requests | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_buffer_pool_pages_flushed  | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_buffer_pool_wait_free      | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Innodb_row_lock_current_waits     | integer |    no    |    no     |    yes     | yes       | hnum |             |


### 3.4 disk  index
* index mapping template: `easedba-monitor-metrics-mysql-disk`
* category: `platform`
* sql : `show global status`
* sql : `select truncate(sum(data_length)/1024/1024,0) as data_size, truncate(sum(index_length)/1024/1024,0) as index_size from information_schema.tables;`
* sql : sum of `show binary logs` 

| field                      |  type   | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :------------------------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| Binlog_cache_disk_use      | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Binlog_stmt_cache_disk_use | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Created_tmp_disk_tables    | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Table_data_size            | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Table_index_size           | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Binary_log_size            | integer |    no    |    no     |    yes     | yes       | hnum |             |


### 3.5 replication index
* index mapping template: `easedba-monitor-metrics-mysql-replication`
* category: `platform`
* SQL : `show slave status`

| field                 |  type   | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :-------------------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| Slave_IO_Running      |  bool   |    no    |    no     |    yes     | yes       | hnum |             |
| Slave_SQL_Running     |  bool   |    no    |    no     |    yes     | yes       | hnum |             |
| Seconds_Behind_Master | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Read_Master_Log_Pos   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Exec_Master_Log_Pos   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| SQL_Delay             | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Last_SQL_Errno        | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Last_IO_Errno         | integer |    no    |    no     |    yes     | yes       | hnum |             |
| Last_SQL_Error        | string  |    no    |    no     |    yes     | yes       | hnum |             |
| Last_IO_Error         | string  |    no    |    no     |    yes     | yes       | hnum |             |

### 3.6 snapshot index
* index mapping template: `easedba-monitor-metrics-mysql-slapshot`
* category: ``platform``
* Running SQL : 

  ```
  SELECT
      id process_id,
      user,
      host,
      db,
      time,
      info sql_text,
      state 
  FROM
      information_schema.processlist 
  WHERE
      info IS NOT NULL;
  ```
* Blocked Transactions :

  ```
  SELECT
    a.trx_mysql_thread_id process_id,
    d.thread_id,
    a.trx_id,
    a.trx_state,
    a.trx_started,
    a.trx_wait_started,
    a.trx_query,
    a.trx_isolation_level,
    b.blocking_trx_id,
    e.thread_id blocking_thread_id,
    c.trx_mysql_thread_id blocking_process_id,
    d.processlist_user user,
    d.processlist_host client,
    d.processlist_db db 
FROM
    information_schema.innodb_trx a 
    LEFT JOIN
        information_schema.innodb_lock_waits b 
        ON a.trx_id = b.requesting_trx_id 
    LEFT JOIN
        information_schema.innodb_trx c 
        ON b.blocking_trx_id = c.trx_id 
    LEFT JOIN
        performance_schema.threads d 
        ON a.trx_mysql_thread_id = d.processlist_id 
    LEFT JOIN
        performance_schema.threads e 
        ON c.trx_mysql_thread_id = e.processlist_id
  ```

* History sqls in blocking transactions:

  ```
  SELECT
    b.processlist_id process_id,
    a.thread_id,
    a.sql_text,
    b.processlist_user USER,
    b.processlist_host client,
    b.processlist_db db 
FROM
    performance_schema.events_statements_history a 
    LEFT JOIN
        performance_schema.threads b 
        ON a.thread_id = b.thread_id 
WHERE
    a.thread_id IN ( %s )
ORDER BY
    a.event_id DESC LIMIT 20;
  ```


| field        |  type  | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :----------- | :----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| Sql_snapshot | object |    no    |    yes    |     no     | yes       | hnum |             |
| Trx_snapshot | object |    no    |    yes    |     no     | yes       | hnum |             |
| Trx_history  | object |    no    |    yes    |     no     | yes       | hnum |             |



### 3.7 network index
* index mapping template: `easedba-monitor-metrics-mysql-net`
* category: `infrastructure`
* bytes in MBytes

| field        |  type   | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :----------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| bytes_sent   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| bytes_recv   | integer |    no    |    no     |    yes     | yes       | hnum |             |
| packets_sent | integer |    no    |    no     |    yes     | yes       | hnum |             |
| packets_recv | integer |    no    |    no     |    yes     | yes       | hnum |             |
| err_in       | integer |    no    |    no     |    yes     | yes       | hnum |             |
| err_out      | integer |    no    |    no     |    yes     | yes       | hnum |             |
| drop_in      | integer |    no    |    no     |    yes     | yes       | hnum |             |
| drop_out     | integer |    no    |    no     |    yes     | yes       | hnum |             |


### 3.8 disk index
* index mapping template: `easedba-monitor-metrics-mysql-disk`
* category: `infrastructure`
* size in MBytes

| field            |  type   | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :--------------- | :-----: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| free             | integer |    no    |    no     |    yes     | yes       | hnum |             |
| total            | integer |    no    |    no     |    yes     | yes       | hnum |             |
| used             | integer |    no    |    no     |    yes     | yes       | hnum |             |
| used_percent     |  float  |    no    |    no     |    yes     | yes       | hnum |             |
| inodes_free      | integer |    no    |    no     |    yes     | yes       | hnum |             |
| inodes_total     | integer |    no    |    no     |    yes     | yes       | hnum |             |
| inodes_used      | integer |    no    |    no     |    yes     | yes       | hnum |             |
| io_time          | integer |    no    |    no     |    yes     | yes       | hnum |             |
| weighted_io_time | integer |    no    |    no     |    yes     | yes       | hnum |             |
| iops_in_progress | integer |    no    |    no     |    yes     | yes       | hnum |             |

### 3.9 cpu index
* index mapping template: `easedba-monitor-metrics-mysql-disk`
* category: `infrastructure`
* size in MBytes

| field            | type  | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :--------------- | :---: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| cpu_usage_user   | float |    no    |    no     |    yes     | yes       | hnum |             |
| cpu_usage_system | float |    no    |    no     |    yes     | yes       | hnum |             |
| cpu_usage_idle   | float |    no    |    no     |    yes     | yes       | hnum |             |
| cpu_usage_nice   | float |    no    |    no     |    yes     | yes       | hnum |             |



### 3.10 Mem index
* index mapping template: `easedba-monitor-metrics-mysql-disk`
* category: `infrastructure`
* size in MBytes

| field        | type  | indexed? | analyzed? | doc_values | required? | Unit | Description |
| :----------- | :---: | :------: | :-------: | :--------: | :-------- | :--- | :---------- |
| total        | float |    no    |    no     |    yes     | yes       | hnum |             |
| used         | float |    no    |    no     |    yes     | yes       | hnum |             |
| used_percent | float |    no    |    no     |    yes     | yes       | hnum |             |
| buffered     | float |    no    |    no     |    yes     | yes       | hnum |             |
| cached       | float |    no    |    no     |    yes     | yes       | hnum |             |

