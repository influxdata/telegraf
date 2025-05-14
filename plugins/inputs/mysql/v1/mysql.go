package v1

import (
	"bytes"
	"database/sql"
	"strconv"
)

// Mapping represents a mapping between server and export names.
type Mapping struct {
	OnServer string
	InExport string
}

// Mappings is a list of predefined mappings between server and export names.
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

// ParseValue parses a SQL raw byte value into a float64.
// It converts "Yes"/"ON" to 1, "No"/"OFF" to 0, and attempts to parse other values as float64.
func ParseValue(value sql.RawBytes) (float64, error) {
	if bytes.Equal(value, []byte("Yes")) || bytes.Equal(value, []byte("ON")) {
		return 1, nil
	}

	if bytes.Equal(value, []byte("No")) || bytes.Equal(value, []byte("OFF")) {
		return 0, nil
	}
	n, err := strconv.ParseFloat(string(value), 64)
	return n, err
}
