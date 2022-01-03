package v2

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
)

type ConversionFunc func(value sql.RawBytes) (interface{}, error)

func ParseInt(value sql.RawBytes) (interface{}, error) {
	v, err := strconv.ParseInt(string(value), 10, 64)

	// Ignore ErrRange.  When this error is set the returned value is "the
	// maximum magnitude integer of the appropriate bitSize and sign."
	if err, ok := err.(*strconv.NumError); ok && err.Err == strconv.ErrRange {
		return v, nil
	}

	return v, err
}

func ParseUint(value sql.RawBytes) (interface{}, error) {
	return strconv.ParseUint(string(value), 10, 64)
}

func ParseFloat(value sql.RawBytes) (interface{}, error) {
	return strconv.ParseFloat(string(value), 64)
}

func ParseBoolAsInteger(value sql.RawBytes) (interface{}, error) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.EqualFold(value, []byte("ON")) {
		return int64(1), nil
	}

	return int64(0), nil
}

func ParseString(value sql.RawBytes) (interface{}, error) {
	return string(value), nil
}

func ParseGTIDMode(value sql.RawBytes) (interface{}, error) {
	// https://dev.mysql.com/doc/refman/8.0/en/replication-mode-change-online-concepts.html
	v := string(value)
	switch v {
	case "OFF":
		return int64(0), nil
	case "ON":
		return int64(1), nil
	case "OFF_PERMISSIVE":
		return int64(0), nil
	case "ON_PERMISSIVE":
		return int64(1), nil
	default:
		return nil, fmt.Errorf("unrecognized gtid_mode: %q", v)
	}
}

func ParseValue(value sql.RawBytes) (interface{}, error) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.Equal(value, []byte("ON")) {
		return 1, nil
	}

	if bytes.EqualFold(value, []byte("NO")) || bytes.Equal(value, []byte("OFF")) {
		return 0, nil
	}

	if val, err := strconv.ParseInt(string(value), 10, 64); err == nil {
		return val, nil
	}
	if val, err := strconv.ParseUint(string(value), 10, 64); err == nil {
		return val, nil
	}
	if val, err := strconv.ParseFloat(string(value), 64); err == nil {
		return val, nil
	}

	if len(string(value)) > 0 {
		return string(value), nil
	}

	return nil, fmt.Errorf("unconvertible value: %q", string(value))
}

var GlobalStatusConversions = map[string]ConversionFunc{
	"innodb_available_undo_logs":    ParseUint,
	"innodb_buffer_pool_pages_misc": ParseUint,
	"innodb_data_pending_fsyncs":    ParseUint,
	"ssl_ctx_verify_depth":          ParseUint,
	"ssl_verify_depth":              ParseUint,

	// see https://galeracluster.com/library/documentation/galera-status-variables.html
	"wsrep_local_index":          ParseUint,
	"wsrep_local_send_queue_avg": ParseFloat,
}

var GlobalVariableConversions = map[string]ConversionFunc{
	// see https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html
	// see https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html
	"delay_key_write":                  ParseString, // ON, OFF, ALL
	"enforce_gtid_consistency":         ParseString, // ON, OFF, WARN
	"event_scheduler":                  ParseString, // YES, NO, DISABLED
	"gtid_mode":                        ParseGTIDMode,
	"have_openssl":                     ParseBoolAsInteger, // alias for have_ssl
	"have_ssl":                         ParseBoolAsInteger, // YES, DISABLED
	"have_symlink":                     ParseBoolAsInteger, // YES, NO, DISABLED
	"session_track_gtids":              ParseString,
	"session_track_transaction_info":   ParseString,
	"slave_skip_errors":                ParseString,
	"ssl_fips_mode":                    ParseString,
	"transaction_write_set_extraction": ParseString,
	"use_secondary_engine":             ParseString,
}

func ConvertGlobalStatus(key string, value sql.RawBytes) (interface{}, error) {
	if bytes.Equal(value, []byte("")) {
		return nil, nil
	}

	if conv, ok := GlobalStatusConversions[key]; ok {
		return conv(value)
	}

	return ParseValue(value)
}

func ConvertGlobalVariables(key string, value sql.RawBytes) (interface{}, error) {
	if bytes.Equal(value, []byte("")) {
		return nil, nil
	}

	if conv, ok := GlobalVariableConversions[key]; ok {
		return conv(value)
	}

	return ParseValue(value)
}
