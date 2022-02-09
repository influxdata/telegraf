package redis_sentinel

type configFieldType int32

const (
	configFieldTypeInteger configFieldType = iota
	configFieldTypeString
	configFieldTypeFloat
)

// Supported fields for "redis_sentinel_masters"
var measurementMastersFields = map[string]configFieldType{
	"config_epoch":            configFieldTypeInteger,
	"down_after_milliseconds": configFieldTypeInteger,
	"failover_timeout":        configFieldTypeInteger,
	"flags":                   configFieldTypeString,
	"info_refresh":            configFieldTypeInteger,
	"ip":                      configFieldTypeString,
	"last_ok_ping_reply":      configFieldTypeInteger,
	"last_ping_reply":         configFieldTypeInteger,
	"last_ping_sent":          configFieldTypeInteger,
	"link_pending_commands":   configFieldTypeInteger,
	"link_refcount":           configFieldTypeInteger,
	"num_other_sentinels":     configFieldTypeInteger,
	"num_slaves":              configFieldTypeInteger,
	"parallel_syncs":          configFieldTypeInteger,
	"port":                    configFieldTypeInteger,
	"quorum":                  configFieldTypeInteger,
	"role_reported":           configFieldTypeString,
	"role_reported_time":      configFieldTypeInteger,
}

// Supported fields for "redis_sentinel"
var measurementSentinelFields = map[string]configFieldType{
	"active_defrag_hits":              configFieldTypeInteger,
	"active_defrag_key_hits":          configFieldTypeInteger,
	"active_defrag_key_misses":        configFieldTypeInteger,
	"active_defrag_misses":            configFieldTypeInteger,
	"blocked_clients":                 configFieldTypeInteger,
	"client_recent_max_input_buffer":  configFieldTypeInteger,
	"client_recent_max_output_buffer": configFieldTypeInteger,
	"connected_clients":               configFieldTypeInteger, // Renamed to "clients"
	"evicted_keys":                    configFieldTypeInteger,
	"expired_keys":                    configFieldTypeInteger,
	"expired_stale_perc":              configFieldTypeFloat,
	"expired_time_cap_reached_count":  configFieldTypeInteger,
	"instantaneous_input_kbps":        configFieldTypeFloat,
	"instantaneous_ops_per_sec":       configFieldTypeInteger,
	"instantaneous_output_kbps":       configFieldTypeFloat,
	"keyspace_hits":                   configFieldTypeInteger,
	"keyspace_misses":                 configFieldTypeInteger,
	"latest_fork_usec":                configFieldTypeInteger,
	"lru_clock":                       configFieldTypeInteger,
	"migrate_cached_sockets":          configFieldTypeInteger,
	"pubsub_channels":                 configFieldTypeInteger,
	"pubsub_patterns":                 configFieldTypeInteger,
	"redis_version":                   configFieldTypeString,
	"rejected_connections":            configFieldTypeInteger,
	"sentinel_masters":                configFieldTypeInteger,
	"sentinel_running_scripts":        configFieldTypeInteger,
	"sentinel_scripts_queue_length":   configFieldTypeInteger,
	"sentinel_simulate_failure_flags": configFieldTypeInteger,
	"sentinel_tilt":                   configFieldTypeInteger,
	"slave_expires_tracked_keys":      configFieldTypeInteger,
	"sync_full":                       configFieldTypeInteger,
	"sync_partial_err":                configFieldTypeInteger,
	"sync_partial_ok":                 configFieldTypeInteger,
	"total_commands_processed":        configFieldTypeInteger,
	"total_connections_received":      configFieldTypeInteger,
	"total_net_input_bytes":           configFieldTypeInteger,
	"total_net_output_bytes":          configFieldTypeInteger,
	"uptime_in_seconds":               configFieldTypeInteger, // Renamed to "uptime_ns"
	"used_cpu_sys":                    configFieldTypeFloat,
	"used_cpu_sys_children":           configFieldTypeFloat,
	"used_cpu_user":                   configFieldTypeFloat,
	"used_cpu_user_children":          configFieldTypeFloat,
}

// Supported fields for "redis_sentinel_sentinels"
var measurementSentinelsFields = map[string]configFieldType{
	"down_after_milliseconds": configFieldTypeInteger,
	"flags":                   configFieldTypeString,
	"last_hello_message":      configFieldTypeInteger,
	"last_ok_ping_reply":      configFieldTypeInteger,
	"last_ping_reply":         configFieldTypeInteger,
	"last_ping_sent":          configFieldTypeInteger,
	"link_pending_commands":   configFieldTypeInteger,
	"link_refcount":           configFieldTypeInteger,
	"name":                    configFieldTypeString,
	"voted_leader":            configFieldTypeString,
	"voted_leader_epoch":      configFieldTypeInteger,
}

// Supported fields for "redis_sentinel_replicas"
var measurementReplicasFields = map[string]configFieldType{
	"down_after_milliseconds": configFieldTypeInteger,
	"flags":                   configFieldTypeString,
	"info_refresh":            configFieldTypeInteger,
	"last_ok_ping_reply":      configFieldTypeInteger,
	"last_ping_reply":         configFieldTypeInteger,
	"last_ping_sent":          configFieldTypeInteger,
	"link_pending_commands":   configFieldTypeInteger,
	"link_refcount":           configFieldTypeInteger,
	"master_host":             configFieldTypeString,
	"master_link_down_time":   configFieldTypeInteger,
	"master_link_status":      configFieldTypeString,
	"master_port":             configFieldTypeInteger,
	"name":                    configFieldTypeString,
	"role_reported":           configFieldTypeString,
	"role_reported_time":      configFieldTypeInteger,
	"slave_priority":          configFieldTypeInteger,
	"slave_repl_offset":       configFieldTypeInteger,
}
