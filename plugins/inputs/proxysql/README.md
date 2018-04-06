# ProxySQL Input Plugin

The ProxySQL plugin reports metrics from ProxySQL's admin interface

* Global status
* Connection pool info
* Command timing

(https://github.com/sysown/proxysql/blob/master/doc/admin_tables.md)

### Configuration:

```toml
[[inputs.proxysql]]
  servers = ["admin:admin@tcp(127.0.0.1:6032)/"]
  # servers = ["admin:admin@unix(/tmp/proxysql_admin.sock)/"]
```
### Measurements & Fields:

- proxysql
    - active_transactions (integer, count)
    - backend_query_time_nsec (integer, count)
    - client_connections_aborted (integer, count)
    - client_connections_connected (integer, count)
    - client_connections_created (integer, count)
    - client_connections_non_idle (integer, count)
    - com_autocommit (integer, count)
    - com_autocommit_filtered (integer, count)
    - com_commit (integer, count)
    - com_commit_filtered (integer, count)
    - com_rollback (integer, count)
    - com_rollback_filtered (integer, count)
    - com_stmt_close (integer, count)
    - com_stmt_execute (integer, count)
    - com_stmt_prepare (integer, count)
    - connpool_get_conn_failure (integer, count)
    - connpool_get_conn_immediate (integer, count)
    - connpool_get_conn_success (integer, count)
    - connpool_memory_bytes (integer, count)
    - mysql_monitor_workers (integer, count)
    - mysql_thread_workers (integer, count)
    - proxysql_uptime (integer, count)
    - queries_backends_bytes_recv (integer, count)
    - queries_backends_bytes_sent (integer, count)
    - query_cache_entries (integer, count)
    - query_cache_memory_bytes (integer, count)
    - query_cache_purged (integer, count)
    - query_cache_bytes_in (integer, count)
    - query_cache_bytes_out (integer, count)
    - query_cache_count_get (integer, count)
    - query_cache_count_get_ok (integer, count)
    - query_cache_count_set (integer, count)
    - query_processor_time_nsec (integer, count)
    - questions (integer, count)
    - sqlite3_memory_bytes (integer, count)
    - server_connections_aborted (integer, count)
    - server_connections_connected (integer, count)
    - server_connections_created (integer, count)
    - servers_table_version (integer, count)
    - slow_queries (integer, count)
    - stmt_active_total (integer, count)
    - stmt_active_unique (integer, count)
    - stmt_max_stmt_id (integer, gauge) 
    - mysql_backend_buffers_bytes (integer, count)
    - mysql_frontend_buffers_bytes (integer, count)
    - mysql_session_internal_bytes (integer, count)
- proxysql_connection_pool
    - connections_used (integer, count)
    - connections_free (integer, count)
    - connections_ok (integer, count)
    - connections_err (integer, count)
    - queries (integer, count)
    - bytes_sent (integer, count)
    - bytes_received (integer, count)
- proxysql_commands
    - total_time (integer, count)
    - count_total (integer, count)
    - count_100us (integer, count)
    - count_500us (integer, count)
    - count_1ms (integer, count)
	- count_5ms   (integer, count)
	- count_10ms (integer, count)
	- count_50ms (integer, count)
	- count_100ms (integer, count)
	- count_500ms (integer, count)
	- count_1s (integer, count)
	- count_5s (integer, count)
	- count_10s (integer, count)
	- count_inf (integer, count)

**NOTE** The `count_` metrics (other than `count_total`) are counting all the queries that took at most that long but
took longer than the previous bucket. For example, a value in `count_1ms` means that number of queries took less than 1ms
but took longer than 500us

### Tags:

- All measurements:
    - server
    
- proxysql_connection_pool
    - hostgroup
    - server (ip:port)
    - status (eg. ONLINE, OFFLINE)

- proxysql_commands
    - command (eg. SELECT, UPDATE)

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter proxysql --test
> proxysql,host=localhost,server=127.0.0.1:6032 active_transactions=0i,backend_query_time_nsec=761516438i,client_connections_aborted=0i,client_connections_connected=12i,client_connections_created=2745i,client_connections_non_idle=8i,com_autocommit=12i,com_autocommit_filtered=12i,com_commit=2735i,com_commit_filtered=1597i,mysql_backend_buffers_bytes=567296i,mysql_frontend_buffers_bytes=786432i,mysql_session_internal_bytes=42232i,proxysql_uptime=82130i,queries_backends_bytes_recv=2285374i,queries_backends_bytes_sent=2161652i,query_processor_time_nsec=84388416i,server_connections_aborted=88i,server_connections_connected=22i,server_connections_created=2780i 1523048942000000000
> proxysql,host=localhost,server=127.0.0.1:6032 com_rollback=0i,com_rollback_filtered=0i,com_stmt_close=5400i,com_stmt_execute=5400i,com_stmt_prepare=5400i,connpool_get_conn_failure=221505i,connpool_get_conn_immediate=65i,connpool_get_conn_success=9865i,connpool_memory_bytes=997440i,mysql_monitor_workers=8i,mysql_thread_workers=4i,query_cache_count_get=0i,query_cache_memory_bytes=0i,questions=30048i,servers_table_version=140064i,slow_queries=24i,sqlite3_memory_bytes=1575488i,stmt_active_total=0i,stmt_active_unique=0i,stmt_max_stmt_id=103i 1523048942000000000
> proxysql,host=localhost,server=127.0.0.1:6032 query_cache_bytes_in=0i,query_cache_bytes_out=0i,query_cache_count_get_ok=0i,query_cache_count_set=0i,query_cache_entries=0i,query_cache_purged=0i 1523048942000000000
> proxysql_connection_pool,host=localhost,hostgroup=0,hostgroup_host=127.0.0.1:3307,server=127.0.0.1:6032,status=ONLINE bytes_received=568885i,bytes_sent=529267i,connections_err=88i,connections_free=0i,connections_ok=2704i,connections_used=8i,queries=8206i 1523048942000000000
> proxysql_connection_pool,host=localhost,hostgroup=1,hostgroup_host=127.0.0.1:3307,server=127.0.0.1:6032,status=ONLINE bytes_received=1716489i,bytes_sent=1632385i,connections_err=0i,connections_free=14i,connections_ok=32i,connections_used=0i,queries=7730i 1523048942000000000
> proxysql_commands,command=ALTER_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=1i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=9i,count_5s=0i,count_inf=0i,count_total=10i,total_time=24463i 1523048942000000000
> proxysql_commands,command=ALTER_VIEW,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=ANALYZE_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=BEGIN,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CALL,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CHANGE_MASTER,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=COMMIT,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=1648i,count_10ms=2i,count_10s=0i,count_1ms=930i,count_1s=0i,count_500ms=0i,count_500us=50i,count_50ms=0i,count_5ms=105i,count_5s=0i,count_inf=0i,count_total=2735i,total_time=874358i 1523048942000000000
> proxysql_commands,command=CREATE_DATABASE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CREATE_INDEX,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CREATE_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=70i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=6i,count_5ms=1i,count_5s=0i,count_inf=0i,count_total=77i,total_time=719380i 1523048942000000000
> proxysql_commands,command=CREATE_TEMPORARY,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CREATE_TRIGGER,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CREATE_USER,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=CREATE_VIEW,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DEALLOCATE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DELETE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=1i,count_10ms=0i,count_10s=0i,count_1ms=1i,count_1s=0i,count_500ms=0i,count_500us=18i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=20i,total_time=4001i 1523048942000000000
> proxysql_commands,command=DESCRIBE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DROP_DATABASE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DROP_INDEX,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DROP_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=77i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=77i,count_5s=0i,count_inf=0i,count_total=154i,total_time=621399i 1523048942000000000
> proxysql_commands,command=DROP_TRIGGER,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DROP_USER,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=DROP_VIEW,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=GRANT,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=EXECUTE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=EXPLAIN,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=FLUSH,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=INSERT,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=88i,count_10ms=0i,count_10s=0i,count_1ms=88i,count_1s=0i,count_500ms=0i,count_500us=98i,count_50ms=0i,count_5ms=47i,count_5s=0i,count_inf=0i,count_total=321i,total_time=166798i 1523048942000000000
> proxysql_commands,command=KILL,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=LOAD,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=LOCK_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=OPTIMIZE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=PREPARE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=PURGE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=RENAME_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=RESET_MASTER,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=RESET_SLAVE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=REPLACE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=REVOKE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=ROLLBACK,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=SAVEPOINT,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=SELECT,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=5119i,count_10ms=6i,count_10s=0i,count_1ms=1392i,count_1s=0i,count_500ms=0i,count_500us=8003i,count_50ms=3i,count_5ms=295i,count_5s=0i,count_inf=23i,count_total=14841i,total_time=234864094i 1523048942000000000
> proxysql_commands,command=SELECT_FOR_UPDATE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=SET,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=2763i,count_10ms=0i,count_10s=0i,count_1ms=2i,count_1s=0i,count_500ms=0i,count_500us=266i,count_50ms=0i,count_5ms=5i,count_5s=0i,count_inf=0i,count_total=3036i,total_time=72572i 1523048942000000000
> proxysql_commands,command=SHOW_TABLE_STATUS,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=5i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=5i,total_time=2968i 1523048942000000000
> proxysql_commands,command=START_TRANSACTION,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=837i,count_10ms=0i,count_10s=0i,count_1ms=4i,count_1s=0i,count_500ms=0i,count_500us=129i,count_50ms=0i,count_5ms=1i,count_5s=0i,count_inf=0i,count_total=971i,total_time=81879i 1523048942000000000
> proxysql_commands,command=TRUNCATE_TABLE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=UNLOCK_TABLES,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=5i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=5i,total_time=935i 1523048942000000000
> proxysql_commands,command=UPDATE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=1038i,count_10ms=0i,count_10s=0i,count_1ms=29i,count_1s=0i,count_500ms=0i,count_500us=940i,count_50ms=0i,count_5ms=4i,count_5s=0i,count_inf=0i,count_total=2011i,total_time=232319i 1523048942000000000
> proxysql_commands,command=USE,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
> proxysql_commands,command=SHOW,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=22i,count_10ms=3i,count_10s=0i,count_1ms=373i,count_1s=0i,count_500ms=0i,count_500us=46i,count_50ms=0i,count_5ms=11i,count_5s=0i,count_inf=1i,count_total=456i,total_time=10295773i 1523048942000000000
> proxysql_commands,command=UNKNOWN,host=localhost,server=127.0.0.1:6032 count_100ms=0i,count_100us=0i,count_10ms=0i,count_10s=0i,count_1ms=0i,count_1s=0i,count_500ms=0i,count_500us=0i,count_50ms=0i,count_5ms=0i,count_5s=0i,count_inf=0i,count_total=0i,total_time=0i 1523048942000000000
```
