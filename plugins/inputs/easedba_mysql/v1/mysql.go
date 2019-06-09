package easedba_v1

import (
	"database/sql"
	"github.com/influxdata/telegraf/plugins/inputs/easedba_mysql/global"
)

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
	"Com_stmt_execute":   "com_stmt_execute",
	"Com_call_procedure": "com_call_procedure",
	"Slow_queries":       "slow_queries",
}

var ConnectionMappings = map[string]string{
	"Threads_connected": "threads_connected",
	"Aborted_clients":   "aborted_clients",
	"Aborted_connects":  "aborted_connects",
	"Locked_connects":   "locked_connects",
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

var DbsizeMappings = map[string]string{
	"Binlog_cache_disk_use":      "binlog_cache_disk_use",
	"Binlog_stmt_cache_disk_use": "binlog_stmt_cache_disk_use",
	"Created_tmp_disk_tables":    "created_tmp_disk_tables",
	"Table_data_size":            "table_data_size",
	"Table_index_size":           "table_index_size",
	"Binary_log_size":            "binary_log_size",
}

var ReplicationMappings = map[string]string{
	"Slave_IO_Running":      "slave_io_running",
	"Slave_SQL_Running":     "slave_sql_running",
	"Seconds_Behind_Master": "seconds_Behind_Master",
	"Read_Master_Log_Pos":   "read_master_log_pos",
	"Exec_Master_Log_Pos":   "exec_master_log_pos",
	"SQL_Delay":             "sql_delay",
	"Last_SQL_Errno":        "last_sql_errno",
	"Last_IO_Errno":         "last_io_errno",
	"Last_SQL_Error":        "last_sql_error",
	"Last_IO_Error":         "last_io_error",
	"Master_position":       "master_position",
}

var SnapshotMappings = map[string]string{
	"Sql_snapshot":     "sql_snapshot",
	"Slow_query_count": "slow_query_count",
	"Long_trx_count":   "long_trx_count",
	"Trx_snapshot":     "trx_snapshot",
}

type RunningSql struct {
	ProcessId uint64         `json:pricess_id`
	User      string         `json:"user"`
	Host      string         `json:"host"`
	Db        sql.NullString `json:"db"`
	Time      int            `json:"time"`
	SqlText   sql.NullString `json:"sql_text"`
	State     sql.NullString `json:"state"`
}

type RunningSqls struct {
	RunningSqlList []RunningSql `json:"running_sql_list"`
}

type RunningTransaction struct {
	ProcessId           uint64         `json:"process_id"`
	ThreadId            uint64         `json:"thread_id"`
	TrxId               string         `json:"trx_id"`
	TrxState            string         `json:"trx_state"`
	TrxStarted          string         `json:"trx_started"`
	TrxWaitStarted      sql.NullString `json:"trx_wait_started"`
	TrxQuery            sql.NullString `json:"trx_query"`
	TrxIsolationLevel   string         `json:"trx_isolation_level"`
	Blocking_trx_id     sql.NullString `json:"blocking_trx_id"`
	Blocking_thread_id  sql.NullInt64  `json:"blocking_thread_id"`
	Blocking_process_id sql.NullInt64  `json:"blocking_process_id"`
	User                string         `json:"user"`
	Client              string         `json:"client"`
	Db                  sql.NullString `json:"db"`
}

type RunningTransactions struct {
	RunningTransactionList []RunningTransaction `json:"running_transaction_list"`
}

type TransactionHistory struct {
	ProcessId uint64         `json:"process_id"`
	ThreadId  uint64         `json:"thread_id"`
	SqlText   sql.NullString `json:"sql_text"`
	User      string         `json:"user"`
	Client    string         `json:"client"`
	Db        sql.NullString `json:"db"`
}

type TransactionHistories struct {
	TransactionHistoryList []TransactionHistory `json:"transaction_history_list"`
}

var InnodbRatio = map[string]string{
	"Com_select": "Innodb_rows_read",
	"Com_update": "Innodb_rows_updated",
	"Com_insert": "Innodb_rows_inserted",
	"Com_delete": "Innodb_rows_deleted",
}

// { key: schemaName, value: { key: attribute, value: value }
// some of the status attribute are accumulated in msyql. needs to
var GlobalStatus = map[string]*global.Status{}
