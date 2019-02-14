package mongodb

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/influxdata/telegraf"
)

type MongodbData struct {
	StatLine      *StatLine
	Fields        map[string]interface{}
	Tags          map[string]string
	DbData        []DbData
	ShardHostData []DbData
}

type DbData struct {
	Name   string
	Fields map[string]interface{}
}

func NewMongodbData(statLine *StatLine, tags map[string]string) *MongodbData {
	return &MongodbData{
		StatLine: statLine,
		Tags:     tags,
		Fields:   make(map[string]interface{}),
		DbData:   []DbData{},
	}
}

var DefaultStats = map[string]string{
	"inserts":                   "InsertCnt",
	"inserts_per_sec":           "Insert",
	"queries":                   "QueryCnt",
	"queries_per_sec":           "Query",
	"updates":                   "UpdateCnt",
	"updates_per_sec":           "Update",
	"deletes":                   "DeleteCnt",
	"deletes_per_sec":           "Delete",
	"getmores":                  "GetMoreCnt",
	"getmores_per_sec":          "GetMore",
	"commands":                  "CommandCnt",
	"commands_per_sec":          "Command",
	"flushes":                   "FlushesCnt",
	"flushes_per_sec":           "Flushes",
	"flushes_total_time_ns":     "FlushesTotalTime",
	"vsize_megabytes":           "Virtual",
	"resident_megabytes":        "Resident",
	"queued_reads":              "QueuedReaders",
	"queued_writes":             "QueuedWriters",
	"active_reads":              "ActiveReaders",
	"active_writes":             "ActiveWriters",
	"net_in_bytes_count":        "NetInCnt",
	"net_in_bytes":              "NetIn",
	"net_out_bytes_count":       "NetOutCnt",
	"net_out_bytes":             "NetOut",
	"open_connections":          "NumConnections",
	"ttl_deletes":               "DeletedDocumentsCnt",
	"ttl_deletes_per_sec":       "DeletedDocuments",
	"ttl_passes":                "PassesCnt",
	"ttl_passes_per_sec":        "Passes",
	"cursor_timed_out":          "TimedOutC",
	"cursor_timed_out_count":    "TimedOutCCnt",
	"cursor_no_timeout":         "NoTimeoutC",
	"cursor_no_timeout_count":   "NoTimeoutCCnt",
	"cursor_pinned":             "PinnedC",
	"cursor_pinned_count":       "PinnedCCnt",
	"cursor_total":              "TotalC",
	"cursor_total_count":        "TotalCCnt",
	"document_deleted":          "DeletedD",
	"document_inserted":         "InsertedD",
	"document_returned":         "ReturnedD",
	"document_updated":          "UpdatedD",
	"connections_current":       "CurrentC",
	"connections_available":     "AvailableC",
	"connections_total_created": "TotalCreatedC",
}

var DefaultReplStats = map[string]string{
	"repl_inserts":          "InsertRCnt",
	"repl_inserts_per_sec":  "InsertR",
	"repl_queries":          "QueryRCnt",
	"repl_queries_per_sec":  "QueryR",
	"repl_updates":          "UpdateRCnt",
	"repl_updates_per_sec":  "UpdateR",
	"repl_deletes":          "DeleteRCnt",
	"repl_deletes_per_sec":  "DeleteR",
	"repl_getmores":         "GetMoreRCnt",
	"repl_getmores_per_sec": "GetMoreR",
	"repl_commands":         "CommandRCnt",
	"repl_commands_per_sec": "CommandR",
	"member_status":         "NodeType",
	"state":                 "NodeState",
	"repl_lag":              "ReplLag",
	"repl_oplog_window_sec": "OplogTimeDiff",
}

var DefaultClusterStats = map[string]string{
	"jumbo_chunks": "JumboChunksCount",
}

var DefaultShardStats = map[string]string{
	"total_in_use":     "TotalInUse",
	"total_available":  "TotalAvailable",
	"total_created":    "TotalCreated",
	"total_refreshing": "TotalRefreshing",
}

var ShardHostStats = map[string]string{
	"in_use":     "InUse",
	"available":  "Available",
	"created":    "Created",
	"refreshing": "Refreshing",
}

var MmapStats = map[string]string{
	"mapped_megabytes":     "Mapped",
	"non-mapped_megabytes": "NonMapped",
	"page_faults":          "FaultsCnt",
	"page_faults_per_sec":  "Faults",
}

var WiredTigerStats = map[string]string{
	"percent_cache_dirty": "CacheDirtyPercent",
	"percent_cache_used":  "CacheUsedPercent",
}

var WiredTigerExtStats = map[string]string{
	"wtcache_tracked_dirty_bytes":          "TrackedDirtyBytes",
	"wtcache_current_bytes":                "CurrentCachedBytes",
	"wtcache_max_bytes_configured":         "MaxBytesConfigured",
	"wtcache_app_threads_page_read_count":  "AppThreadsPageReadCount",
	"wtcache_app_threads_page_read_time":   "AppThreadsPageReadTime",
	"wtcache_app_threads_page_write_count": "AppThreadsPageWriteCount",
	"wtcache_bytes_written_from":           "BytesWrittenFrom",
	"wtcache_bytes_read_into":              "BytesReadInto",
	"wtcache_pages_evicted_by_app_thread":  "PagesEvictedByAppThread",
	"wtcache_pages_queued_for_eviction":    "PagesQueuedForEviction",
	"wtcache_pages_read_into":              "PagesReadIntoCache",
	"wtcache_pages_requested_from":         "PagesRequestedFromCache",
	"wtcache_server_evicting_pages":        "ServerEvictingPages",
	"wtcache_worker_thread_evictingpages":  "WorkerThreadEvictingPages",
	"wtcache_internal_pages_evicted":       "InternalPagesEvicted",
	"wtcache_modified_pages_evicted":       "ModifiedPagesEvicted",
	"wtcache_unmodified_pages_evicted":     "UnmodifiedPagesEvicted",
}

var DbDataStats = map[string]string{
	"collections":  "Collections",
	"objects":      "Objects",
	"avg_obj_size": "AvgObjSize",
	"data_size":    "DataSize",
	"storage_size": "StorageSize",
	"num_extents":  "NumExtents",
	"indexes":      "Indexes",
	"index_size":   "IndexSize",
	"ok":           "Ok",
}

func (d *MongodbData) AddDbStats() {
	for _, dbstat := range d.StatLine.DbStatsLines {
		dbStatLine := reflect.ValueOf(&dbstat).Elem()
		newDbData := &DbData{
			Name:   dbstat.Name,
			Fields: make(map[string]interface{}),
		}
		newDbData.Fields["type"] = "db_stat"
		for key, value := range DbDataStats {
			val := dbStatLine.FieldByName(value).Interface()
			newDbData.Fields[key] = val
		}
		d.DbData = append(d.DbData, *newDbData)
	}
}

func (d *MongodbData) AddShardHostStats() {
	for host, hostStat := range d.StatLine.ShardHostStatsLines {
		hostStatLine := reflect.ValueOf(&hostStat).Elem()
		newDbData := &DbData{
			Name:   host,
			Fields: make(map[string]interface{}),
		}
		newDbData.Fields["type"] = "shard_host_stat"
		for k, v := range ShardHostStats {
			val := hostStatLine.FieldByName(v).Interface()
			newDbData.Fields[k] = val
		}
		d.ShardHostData = append(d.ShardHostData, *newDbData)
	}
}

func (d *MongodbData) AddDefaultStats() {
	statLine := reflect.ValueOf(d.StatLine).Elem()
	d.addStat(statLine, DefaultStats)
	if d.StatLine.NodeType != "" {
		d.addStat(statLine, DefaultReplStats)
	}
	d.addStat(statLine, DefaultClusterStats)
	d.addStat(statLine, DefaultShardStats)
	if d.StatLine.StorageEngine == "mmapv1" || d.StatLine.StorageEngine == "rocksdb" {
		d.addStat(statLine, MmapStats)
	} else if d.StatLine.StorageEngine == "wiredTiger" {
		for key, value := range WiredTigerStats {
			val := statLine.FieldByName(value).Interface()
			percentVal := fmt.Sprintf("%.1f", val.(float64)*100)
			floatVal, _ := strconv.ParseFloat(percentVal, 64)
			d.add(key, floatVal)
		}
		d.addStat(statLine, WiredTigerExtStats)
	}
}

func (d *MongodbData) addStat(statLine reflect.Value, stats map[string]string) {
	for key, value := range stats {
		val := statLine.FieldByName(value).Interface()
		d.add(key, val)
	}
}

func (d *MongodbData) add(key string, val interface{}) {
	d.Fields[key] = val
}

func (d *MongodbData) flush(acc telegraf.Accumulator) {
	acc.AddFields(
		"mongodb",
		d.Fields,
		d.Tags,
		d.StatLine.Time,
	)
	d.Fields = make(map[string]interface{})

	for _, db := range d.DbData {
		d.Tags["db_name"] = db.Name
		acc.AddFields(
			"mongodb_db_stats",
			db.Fields,
			d.Tags,
			d.StatLine.Time,
		)
		db.Fields = make(map[string]interface{})
	}
	for _, host := range d.ShardHostData {
		d.Tags["hostname"] = host.Name
		acc.AddFields(
			"mongodb_shard_stats",
			host.Fields,
			d.Tags,
			d.StatLine.Time,
		)
		host.Fields = make(map[string]interface{})
	}
}
