package mongodb

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/influxdata/telegraf"
)

type MongodbData struct {
	StatLine *StatLine
	Fields   map[string]interface{}
	Tags     map[string]string
}

func NewMongodbData(statLine *StatLine, tags map[string]string) *MongodbData {
	if statLine.NodeType != "" && statLine.NodeType != "UNK" {
		tags["state"] = statLine.NodeType
	}
	return &MongodbData{
		StatLine: statLine,
		Tags:     tags,
		Fields:   make(map[string]interface{}),
	}
}

var DefaultStats = map[string]string{
	"inserts_per_sec":     "Insert",
	"queries_per_sec":     "Query",
	"updates_per_sec":     "Update",
	"deletes_per_sec":     "Delete",
	"getmores_per_sec":    "GetMore",
	"commands_per_sec":    "Command",
	"flushes_per_sec":     "Flushes",
	"vsize_megabytes":     "Virtual",
	"resident_megabytes":  "Resident",
	"queued_reads":        "QueuedReaders",
	"queued_writes":       "QueuedWriters",
	"active_reads":        "ActiveReaders",
	"active_writes":       "ActiveWriters",
	"net_in_bytes":        "NetIn",
	"net_out_bytes":       "NetOut",
	"open_connections":    "NumConnections",
	"ttl_deletes_per_sec": "DeletedDocuments",
	"ttl_passes_per_sec":  "Passes",
}

var DefaultReplStats = map[string]string{
	"repl_inserts_per_sec":  "InsertR",
	"repl_queries_per_sec":  "QueryR",
	"repl_updates_per_sec":  "UpdateR",
	"repl_deletes_per_sec":  "DeleteR",
	"repl_getmores_per_sec": "GetMoreR",
	"repl_commands_per_sec": "CommandR",
	"member_status":         "NodeType",
	"repl_lag":              "ReplLag",
}

var DefaultClusterStats = map[string]string{
	"jumbo_chunks": "JumboChunksCount",
}

var MmapStats = map[string]string{
	"mapped_megabytes":     "Mapped",
	"non-mapped_megabytes": "NonMapped",
	"page_faults_per_sec":  "Faults",
}

var WiredTigerStats = map[string]string{
	"percent_cache_dirty": "CacheDirtyPercent",
	"percent_cache_used":  "CacheUsedPercent",
}

func (d *MongodbData) AddDefaultStats() {
	statLine := reflect.ValueOf(d.StatLine).Elem()
	d.addStat(statLine, DefaultStats)
	if d.StatLine.NodeType != "" {
		d.addStat(statLine, DefaultReplStats)
	}
	d.addStat(statLine, DefaultClusterStats)
	if d.StatLine.StorageEngine == "mmapv1" {
		d.addStat(statLine, MmapStats)
	} else if d.StatLine.StorageEngine == "wiredTiger" {
		for key, value := range WiredTigerStats {
			val := statLine.FieldByName(value).Interface()
			percentVal := fmt.Sprintf("%.1f", val.(float64)*100)
			floatVal, _ := strconv.ParseFloat(percentVal, 64)
			d.add(key, floatVal)
		}
	}
}

func (d *MongodbData) addStat(
	statLine reflect.Value,
	stats map[string]string,
) {
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
}
