package v1

import (
	"bytes"
	"database/sql"
	"strconv"
)

type Mapping struct {
	OnServer string
	InExport string
}

var Mappings = []*Mapping{
	{
		OnServer: "Aborted_",
		InExport: "aborted_",
	},
	{
		OnServer: "Bytes_",
		InExport: "bytes_",
	},
	{
		OnServer: "Com_",
		InExport: "commands_",
	},
	{
		OnServer: "Created_",
		InExport: "created_",
	},
	{
		OnServer: "Handler_",
		InExport: "handler_",
	},
	{
		OnServer: "Innodb_",
		InExport: "innodb_",
	},
	{
		OnServer: "Key_",
		InExport: "key_",
	},
	{
		OnServer: "Open_",
		InExport: "open_",
	},
	{
		OnServer: "Opened_",
		InExport: "opened_",
	},
	{
		OnServer: "Qcache_",
		InExport: "qcache_",
	},
	{
		OnServer: "Table_",
		InExport: "table_",
	},
	{
		OnServer: "Tokudb_",
		InExport: "tokudb_",
	},
	{
		OnServer: "Threads_",
		InExport: "threads_",
	},
	{
		OnServer: "Access_",
		InExport: "access_",
	},
	{
		OnServer: "Aria__",
		InExport: "aria_",
	},
	{
		OnServer: "Binlog__",
		InExport: "binlog_",
	},
	{
		OnServer: "Busy_",
		InExport: "busy_",
	},
	{
		OnServer: "Connection_",
		InExport: "connection_",
	},
	{
		OnServer: "Delayed_",
		InExport: "delayed_",
	},
	{
		OnServer: "Empty_",
		InExport: "empty_",
	},
	{
		OnServer: "Executed_",
		InExport: "executed_",
	},
	{
		OnServer: "Executed_",
		InExport: "executed_",
	},
	{
		OnServer: "Feature_",
		InExport: "feature_",
	},
	{
		OnServer: "Flush_",
		InExport: "flush_",
	},
	{
		OnServer: "Last_",
		InExport: "last_",
	},
	{
		OnServer: "Master_",
		InExport: "master_",
	},
	{
		OnServer: "Max_",
		InExport: "max_",
	},
	{
		OnServer: "Memory_",
		InExport: "memory_",
	},
	{
		OnServer: "Not_",
		InExport: "not_",
	},
	{
		OnServer: "Performance_",
		InExport: "performance_",
	},
	{
		OnServer: "Prepared_",
		InExport: "prepared_",
	},
	{
		OnServer: "Rows_",
		InExport: "rows_",
	},
	{
		OnServer: "Rpl_",
		InExport: "rpl_",
	},
	{
		OnServer: "Select_",
		InExport: "select_",
	},
	{
		OnServer: "Slave_",
		InExport: "slave_",
	},
	{
		OnServer: "Slow_",
		InExport: "slow_",
	},
	{
		OnServer: "Sort_",
		InExport: "sort_",
	},
	{
		OnServer: "Subquery_",
		InExport: "subquery_",
	},
	{
		OnServer: "Tc_",
		InExport: "tc_",
	},
	{
		OnServer: "Threadpool_",
		InExport: "threadpool_",
	},
	{
		OnServer: "wsrep_",
		InExport: "wsrep_",
	},
	{
		OnServer: "Uptime_",
		InExport: "uptime_",
	},
}

var ThroughtMappings = map[string]string{
	"Com_insert":         "com_insert",
	"Com_select":         "com_select",
	"Com_insert_select":  "com_insert_select",
	"Com_replace":        "com_replace",
	"Com_replace_select": "com_replace_select",
	"Com_update":         "com_update",
	"Com_update_multi":   "com_update_multi",
	"Com_delete":         "com_delete",
	"Com_delete_multi":   "com_delete_multi",
	"Com_commit":         "com_commit",
	"Com_rollback":       "com_rollback",
	"Com_stmt_exexute":   "com_stmt_exexute",
	"Com_call_procedure": "com_call_procedure",
}

var ConnectionMappings = map[string]string{
	"Connections":      "connections",
	"Aborted_clients":  "aborted_clients",
	"Aborted_connects": "aborted_connects",
	"Locked_connects":  "locked_connects",
}

var InnodbMappings = map[string]string{
	"Innodb_rows_read":                  "innodb_rows_read",
	"Innodb_rows_read_ratio":            "innodb_rows_read_ratio",
	"Innodb_rows_deleted":               "innodb_rows_deleted",
	"Innodb_rows_deleted_ratio":         "innodb_rows_deleted_ratio",
	"Innodb_rows_inserted":              "innodb_rows_inserted",
	"Innodb_rows_inserted_ratio":        "innodb_rows_inserted_ratio",
	"Innodb_rows_updated":               "innodb_rows_updated",
	"Innodb_rows_updated_ratio":         "innodb_rows_updated_ratio",
	"Innodb_buffer_pool_reads":          "innodb_buffer_pool_reads",
	"Innodb_buffer_pool_read_requests":  "innodb_buffer_pool_read_requests",
	"Innodb_buffer_pool_write_requests": "innodb_buffer_pool_write_requests",
	"Innodb_buffer_pool_pages_flushed":  "innodb_buffer_pool_pages_flushed",
	"Innodb_buffer_pool_wait_free":      "innodb_buffer_pool_wait_free",
	"Innodb_row_lock_current_waits":     "innodb_row_lock_current_waits",
}

var dbsizeMappings = map[string]string{
	"Binlog_cache_disk_use":      "binlog_cache_disk_use",
	"Binlog_stmt_cache_disk_use": "binlog_stmt_cache_disk_use",
	"Created_tmp_disk_tables":    "created_tmp_disk_tables",
	"Table_data_size":            "table_data_size",
	"Table_index_size":           "table_index_size",
	"Binary_log_size":            "binary_log_size",
}

var replicationMappings = map[string]string{
	"Slave_IO_State":           "slave_IO_State",
	"Slave_IO_Running":         "slave_IO_Running",
	"Slave_SQL_Running":        "slave_SQL_Running",
	"Seconds_Behind_Master":    "seconds_Behind_Master",
	"Read_Master_Log_Pos":      "read_Master_Log_Pos",
	"Exec_Master_Log_Pos":      "exec_Master_Log_Pos",
	"Retrieved_Gtid_Set":       "retrieved_Gtid_Set",
	"Executed_Gtid_Set":        "executed_Gtid_Set",
	"SQL_Delay":                "sQL_Delay",
	"Last_SQL_Errno":           "last_SQL_Errno",
	"Last_IO_Errno":            "last_IO_Errno",
	"Master_position":          "master_position",
	"Master_Executed_Gtid_Set": "master_Executed_Gtid_Set",
}

var snapshotMappings = map[string]string{
	"Sql_snapshot":     "sql_snapshot",
	"Slow_query_count": "slow_query_count",
	"Long_trx_count":   "long_trx_count",
	"Trx_snapshot":     "trx_snapshot",
}

func ParseValue(value sql.RawBytes) (float64, bool) {
	if bytes.Compare(value, []byte("Yes")) == 0 || bytes.Compare(value, []byte("ON")) == 0 {
		return 1, true
	}

	if bytes.Compare(value, []byte("No")) == 0 || bytes.Compare(value, []byte("OFF")) == 0 {
		return 0, true
	}
	n, err := strconv.ParseFloat(string(value), 64)
	return n, err == nil
}
