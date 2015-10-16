package mongodb

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/influxdb/telegraf/plugins"
)

type MongodbData struct {
	StatLine *StatLine
	Tags     map[string]string
}

func NewMongodbData(statLine *StatLine, tags map[string]string) *MongodbData {
	if statLine.NodeType != "" && statLine.NodeType != "UNK" {
		tags["state"] = statLine.NodeType
	}
	return &MongodbData{
		StatLine: statLine,
		Tags:     tags,
	}
}

var DefaultStats = map[string]string{
	"inserts_per_sec":    "Insert",
	"queries_per_sec":    "Query",
	"updates_per_sec":    "Update",
	"deletes_per_sec":    "Delete",
	"getmores_per_sec":   "GetMore",
	"commands_per_sec":   "Command",
	"flushes_per_sec":    "Flushes",
	"vsize_megabytes":    "Virtual",
	"resident_megabytes": "Resident",
	"queued_reads":       "QueuedReaders",
	"queued_writes":      "QueuedWriters",
	"active_reads":       "ActiveReaders",
	"active_writes":      "ActiveWriters",
	"net_in_bytes":       "NetIn",
	"net_out_bytes":      "NetOut",
	"open_connections":   "NumConnections",
}

var DefaultReplStats = map[string]string{
	"repl_inserts_per_sec":  "InsertR",
	"repl_queries_per_sec":  "QueryR",
	"repl_updates_per_sec":  "UpdateR",
	"repl_deletes_per_sec":  "DeleteR",
	"repl_getmores_per_sec": "GetMoreR",
	"repl_commands_per_sec": "CommandR",
	"member_status":         "NodeType",
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

func (d *MongodbData) AddDefaultStats(acc plugins.Accumulator) {
	statLine := reflect.ValueOf(d.StatLine).Elem()
	d.addStat(acc, statLine, DefaultStats)
	if d.StatLine.NodeType != "" {
		d.addStat(acc, statLine, DefaultReplStats)
	}
	if d.StatLine.StorageEngine == "mmapv1" {
		d.addStat(acc, statLine, MmapStats)
	} else if d.StatLine.StorageEngine == "wiredTiger" {
		for key, value := range WiredTigerStats {
			val := statLine.FieldByName(value).Interface()
			percentVal := fmt.Sprintf("%.1f", val.(float64)*100)
			floatVal, _ := strconv.ParseFloat(percentVal, 64)
			d.add(acc, key, floatVal)
		}
	}
}

func (d *MongodbData) addStat(acc plugins.Accumulator, statLine reflect.Value, stats map[string]string) {
	for key, value := range stats {
		val := statLine.FieldByName(value).Interface()
		d.add(acc, key, val)
	}
}

func (d *MongodbData) add(acc plugins.Accumulator, key string, val interface{}) {
	acc.AddFields(
		key,
		map[string]interface{}{
			"value": val,
		},
		d.Tags,
		d.StatLine.Time,
	)
}
