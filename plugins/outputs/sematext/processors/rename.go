package processors

import (
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
)

const (
	winOSLabel  = "win"
	memLabel    = "mem"
	systemLabel = "system"
	netLabel    = "net"

	osLabel      = "os"
	memoryLabel  = "memory"
	hostLabel    = "host"
	networkLabel = "net"

	underscoreLabel = "_"
	dotLabel        = "."
)

var (
	measurementReplaces = map[string]string{
		"phpfpm":              "php",
		"mongodb":             "mongo",
		"mongodb_db_stats":    "mongo",
		"mongodb_col_stats":   "mongo",
		"mongodb_shard_stats": "mongo",
		"apache":              "apache",
		"nginx":               "nginx",
		"win_cpu":             "os.cpu",
		"win_disk":            "os.disk",
		"win_diskio":          "os.diskio",
		"win_mem":             "os.memory",
		"win_net":             "os.network",
		"win_swap":            "os.swap",
		"win_system":          "os.host",
		"win_eventlog":        "os.eventlog",
	}

	fieldReplaces = map[string]string{
		// apache
		"apache.BusyWorkers":          "workers.busy",
		"apache.BytesPerReq":          "bytes",
		"apache.ReqPerSec":            "requests",
		"apache.ConnsAsyncClosing":    "connections.async.closing",
		"apache.ConnsAsyncKeepAlive":  "connections.async.keepAlive",
		"apache.ConnsAsyncWriting":    "connections.async.writing",
		"apache.ConnsTotal":           "connections",
		"apache.IdleWorkers":          "workers.idle",
		"apache.scboard_closing":      "workers.closing",
		"apache.scboard_dnslookup":    "workers.dns",
		"apache.scboard_finishing":    "workers.finishing",
		"apache.scboard_idle_cleanup": "workers.cleanup",
		"apache.scboard_keepalive":    "workers.keepalive",
		"apache.scboard_logging":      "workers.logging",
		"apache.scboard_open":         "workers.open",
		"apache.scboard_reading":      "workers.reading",
		"apache.scboard_sending":      "workers.sending",
		"apache.scboard_starting":     "workers.starting",
		"apache.scboard_waiting":      "workers.waiting",
		"phpfpm.accepted_conn":        "fpm.requests.accepted.conns",
		"phpfpm.listen_queue":         "fpm.queue.listen",
		"phpfpm.max_listen_queue":     "fpm.queue.listen.max",
		"phpfpm.listen_queue_len":     "fpm.queue.listen.len",
		"phpfpm.idle_processes":       "fpm.process.idle",
		"phpfpm.active_processes":     "fpm.process.active",
		"phpfpm.total_processes":      "fpm.process.total",
		"phpfpm.max_active_processes": "fpm.process.active.max",
		"phpfpm.max_children_reached": "fpm.process.childrenReached.max",
		"phpfpm.slow_requests":        "fpm.requests.slow",
		// nginx
		"nginx.accepts":  "requests.connections.accepted",
		"nginx.handled":  "requests.connections.handled",
		"nginx.active":   "requests.connections.active",
		"nginx.reading":  "requests.connections.reading",
		"nginx.writing":  "requests.connections.writing",
		"nginx.waiting":  "requests.connections.waiting",
		"nginx.requests": "request.count",
		// mongodb
		"mongodb.flushes":                   "flushes",
		"mongodb.flushes_total_time_ns":     "flushes.time",
		"mongodb.document_inserted":         "documents.inserted",
		"mongodb.document_updated":          "documents.updated",
		"mongodb.document_deleted":          "documents.deleted",
		"mongodb.document_returned":         "documents.returned",
		"mongodb.resident_megabytes":        "memory.resident",
		"mongodb.vsize_megabytes":           "memory.virtual",
		"mongodb.mapped_megabytes":          "memory.mapped",
		"mongodb.inserts":                   "ops.insert",
		"mongodb.queries":                   "ops.query",
		"mongodb.updates":                   "ops.update",
		"mongodb.getmores":                  "ops.getmore",
		"mongodb.commands":                  "ops.command",
		"mongodb.repl_inserts":              "replica.ops.insert",
		"mongodb.repl_queries":              "replica.ops.query",
		"mongodb.repl_updates":              "replica.ops.update",
		"mongodb.repl_deletes":              "replica.ops.delete",
		"mongodb.repl_getmores":             "replica.ops.getmore",
		"mongodb.repl_commands":             "replica.ops.command",
		"mongodb.count_command_failed":      "commands.failed",
		"mongodb.count_command_total":       "commands.total",
		"mongodb_db_stats.data_size":        "database.data.size",
		"mongodb_db_stats.storage_size":     "database.storage.size",
		"mongodb_db_stats.index_size":       "database.index.size",
		"mongodb_db_stats.collections":      "database.collections",
		"mongodb_db_stats.objects":          "database.objects",
		"mongodb_db_stats.avg_obj_size":     "database.avg_obj_size",
		"mongodb_db_stats.indexes":          "database.indexes",
		"mongodb_db_stats.num_extents":      "database.num_extents",
		"mongodb_db_stats.ok":               "database.ok",
		"mongodb.connections_current":       "network.connections",
		"mongodb.connections_total_created": "network.connections.total",
		"mongodb.net_in_bytes":              "network.transfer.rx.rate",
		"mongodb.net_out_bytes":             "network.transfer.tx.rate",
		// mongodb_col_stats -> these appear like they map to the same thing, but "from" side is actually
		// "name.metricName" and "to" side is just the new "metricName"
		"mongodb_col_stats.avg_obj_size":     "mongodb_col_stats.avg_obj_size",
		"mongodb_col_stats.count":            "mongodb_col_stats.count",
		"mongodb_col_stats.ok":               "mongodb_col_stats.ok",
		"mongodb_col_stats.size":             "mongodb_col_stats.size",
		"mongodb_col_stats.storage_size":     "mongodb_col_stats.storage_size",
		"mongodb_col_stats.total_index_size": "mongodb_col_stats.total_index_size",
		// mongodb_shard_stats -> same logic as for mongodb_col_stats
		"mongodb_shard_stats.in_use":     "mongodb_shard_stats.in_use",
		"mongodb_shard_stats.available":  "mongodb_shard_stats.available",
		"mongodb_shard_stats.created":    "mongodb_shard_stats.created",
		"mongodb_shard_stats.refreshing": "mongodb_shard_stats.refreshing",

		"mongodb.tcmalloc_heap_size":                     "tcmalloc_heap_size",
		"mongodb.tcmalloc_current_allocated_bytes":       "tcmalloc_current_allocated_bytes",
		"mongodb.tcmalloc_total_free_bytes":              "tcmalloc_total_free_bytes",
		"mongodb.tcmalloc_pageheap_free_bytes":           "tcmalloc_pageheap_free_bytes",
		"mongodb.tcmalloc_pageheap_unmapped_bytes":       "tcmalloc_pageheap_unmapped_bytes",
		"mongodb.tcmalloc_pageheap_committed_bytes":      "tcmalloc_pageheap_committed_bytes",
		"mongodb.tcmalloc_central_cache_free_bytes":      "tcmalloc_central_cache_free_bytes",
		"mongodb.tcmalloc_thread_cache_free_bytes":       "tcmalloc_thread_cache_free_bytes",
		"mongodb.tcmalloc_transfer_cache_free_bytes":     "tcmalloc_transfer_cache_free_bytes",
		"mongodb.tcmalloc_pageheap_scavenge_count":       "tcmalloc_pageheap_scavenge_count",
		"mongodb.tcmalloc_pageheap_commit_count":         "tcmalloc_pageheap_commit_count",
		"mongodb.tcmalloc_pageheap_decommit_count":       "tcmalloc_pageheap_decommit_count",
		"mongodb.tcmalloc_pageheap_total_decommit_bytes": "tcmalloc_pageheap_total_decommit_bytes",
		"mongodb.tcmalloc_pageheap_reserve_count":        "tcmalloc_pageheap_reserve_count",
		"mongodb.tcmalloc_pageheap_total_reserve_bytes":  "tcmalloc_pageheap_total_reserve_bytes",

		"mongodb.latency_reads":    "latency_reads",
		"mongodb.latency_writes":   "latency_writes",
		"mongodb.latency_commands": "latency_commands",

		"mongodb.cursor_timed_out_count":  "cursor_timed_out_count",
		"mongodb.cursor_total_count":      "cursor_total_count",
		"mongodb.cursor_pinned_count":     "cursor_pinned_count",
		"mongodb.cursor_no_timeout_count": "cursor_no_timeout_count",

		"mongodb.storage_freelist_search_bucket_exhausted": "storage_freelist_search_bucket_exhausted",
		"mongodb.storage_freelist_search_requests":         "storage_freelist_search_requests",
		"mongodb.storage_freelist_search_scanned":          "storage_freelist_search_scanned",

		"mongodb.wtcache_tracked_dirty_bytes":         "wtcache_tracked_dirty_bytes",
		"mongodb.wtcache_current_bytes":               "wtcache_current_bytes",
		"mongodb.wtcache_bytes_written_from":          "wtcache_bytes_written_from",
		"mongodb.wtcache_bytes_read_into":             "wtcache_bytes_read_into",
		"mongodb.wtcache_pages_read_into":             "wtcache_pages_read_into",
		"mongodb.wtcache_pages_written_from":          "wtcache_pages_written_from",
		"mongodb.wtcache_pages_requested_from":        "wtcache_pages_requested_from",
		"mongodb.wtcache_internal_pages_evicted":      "wtcache_internal_pages_evicted",
		"mongodb.wtcache_modified_pages_evicted":      "wtcache_modified_pages_evicted",
		"mongodb.wtcache_unmodified_pages_evicted":    "wtcache_unmodified_pages_evicted",
		"mongodb.wtcache_worker_thread_evictingpages": "wtcache_worker_thread_evictingpages",
		"mongodb.wtcache_pages_evicted_by_app_thread": "wtcache_pages_evicted_by_app_thread",
		"mongodb.wtcache_pages_queued_for_eviction":   "wtcache_pages_queued_for_eviction",

		// windows
		"win_cpu.Percent_DPC_Time":                    "percentage.dpc.time",
		"win_cpu.Percent_Idle_Time":                   "percentage.idle.time",
		"win_cpu.Percent_Interrupt_Time":              "percentage.interrupt.time",
		"win_cpu.Percent_Privileged_Time":             "percentage.privileged.time",
		"win_cpu.Percent_Processor_Time":              "percentage.processor.time",
		"win_cpu.Percent_User_Time":                   "percentage.user.time",
		"win_disk.Percent_Free_Space":                 "percentage.free.bytes",
		"win_diskio.Disk_Read_Bytes_persec":           "read.bytes",
		"win_diskio.Disk_Write_Bytes_persec":          "write.bytes",
		"win_mem.Available_Bytes":                     "free",
		"win_mem.Modified_Page_List_Bytes":            "modified.page.list.bytes",
		"win_mem.Standby_Cache_Core_Bytes":            "standby.cache.core.bytes",
		"win_mem.Standby_Cache_Normal_Priority_Bytes": "standby.cache.normal.priority.bytes",
		"win_mem.Standby_Cache_Reserve_Bytes":         "standby.cache.reserve.bytes",
		"win_net.Bytes_Received_persec":               "rx",
		"win_net.Bytes_Sent_persec":                   "tx",
		"win_swap.Percent_Usage":                      "percentage.usage",
		"win_system.Processor_Queue_Length":           "processor.queue.length",
		"win_eventlog.Version":                        "version",
		"win_eventlog.EventRecordID":                  "eventrecordid",
		"win_eventlog.Task":                           "task",
	}
)

// Rename processor renames the measurement (metric) names
// to match the existing metric names sent by Node.js agents
type Rename struct{}

// NewRename builds a new rename processor.
func NewRename() BatchProcessor { return &Rename{} }

// Process performs a lookup in the local maps of metric/field names
// and replaces the metric name with the new name.
func (r *Rename) Process(points []telegraf.Metric) []telegraf.Metric {
	for _, point := range points {
		originalName := point.Name()
		replace, ok := measurementReplaces[originalName]
		if !ok {
			replace = ChangeNames(originalName)
		}
		point.SetName(replace)
		removedFields := make([]string, 0)
		for _, field := range point.FieldList() {
			key := originalName + "." + field.Key
			replace, ok := fieldReplaces[key]
			if !ok {
				replace = ChangeNames(key)
			}
			// we can't remove the fields
			// while iterating because it
			// produces unwanted effects
			// e.g. metrics that have the
			// mapping are not renamed.
			// That's why we have to remove
			// them in a separate loop
			if replace != field.Key {
				removedFields = append(removedFields, field.Key)
			}
			point.AddField(replace, field.Value)
		}
		for _, f := range removedFields {
			point.RemoveField(f)
		}
	}
	return points
}

// Different windows versions have different metrics
// and the user can decide which metrics to ship in telegraf config
// since we can't get all metrics and directly replace their names
// this part will make sure that there are no incompatible symbols
// inside the metric nanmes
func ChangeNames(name string) string {
	baseNameChanger := strings.NewReplacer(winOSLabel, osLabel, memLabel, memoryLabel, systemLabel, hostLabel, netLabel, networkLabel)
	sanitizedChars := strings.NewReplacer(",", "", ":", "", "+", "", "&", "", "(", "", ")", "")
	extraChars := regexp.MustCompile("__+")
	//change some common labels to match sematext ones
	//for metrics that we haven't covered
	name = baseNameChanger.Replace(name)
	//remove unsupported characters
	name = sanitizedChars.Replace(name)
	//remove extra __ created by removing sanitized chars
	name = extraChars.ReplaceAllString(name, "")
	//remove _ at the end of the field name (if it exists)
	if len(name) > 1 && strings.Compare(name[len(name)-1:], underscoreLabel) == 0 {
		name = name[:len(name)-1]
	}
	//finally, change all underscores to dots
	return strings.ToLower(strings.Replace(name, underscoreLabel, dotLabel, -1))
}

func (Rename) Close() {}
