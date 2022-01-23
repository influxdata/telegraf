package rethinkdb

import (
	"reflect"
	"time"

	"github.com/influxdata/telegraf"
)

type serverStatus struct {
	ID      string `gorethink:"id"`
	Network struct {
		Addresses  []Address `gorethink:"canonical_addresses"`
		Hostname   string    `gorethink:"hostname"`
		DriverPort int       `gorethink:"reql_port"`
	} `gorethink:"network"`
	Process struct {
		Version      string    `gorethink:"version"`
		RunningSince time.Time `gorethink:"time_started"`
	} `gorethink:"process"`
}

type Address struct {
	Host string `gorethink:"host"`
	Port int    `gorethink:"port"`
}

type stats struct {
	Engine Engine `gorethink:"query_engine"`
}

type Engine struct {
	ClientConns   int64 `gorethink:"client_connections,omitempty"`
	ClientActive  int64 `gorethink:"clients_active,omitempty"`
	QueriesPerSec int64 `gorethink:"queries_per_sec,omitempty"`
	TotalQueries  int64 `gorethink:"queries_total,omitempty"`
	ReadsPerSec   int64 `gorethink:"read_docs_per_sec,omitempty"`
	TotalReads    int64 `gorethink:"read_docs_total,omitempty"`
	WritesPerSec  int64 `gorethink:"written_docs_per_sec,omitempty"`
	TotalWrites   int64 `gorethink:"written_docs_total,omitempty"`
}

type tableStatus struct {
	ID   string `gorethink:"id"`
	DB   string `gorethink:"db"`
	Name string `gorethink:"name"`
}

type tableStats struct {
	Engine  Engine  `gorethink:"query_engine"`
	Storage Storage `gorethink:"storage_engine"`
}

type Storage struct {
	Cache Cache `gorethink:"cache"`
	Disk  Disk  `gorethink:"disk"`
}

type Cache struct {
	BytesInUse int64 `gorethink:"in_use_bytes"`
}

type Disk struct {
	ReadBytesPerSec  int64      `gorethink:"read_bytes_per_sec"`
	ReadBytesTotal   int64      `gorethink:"read_bytes_total"`
	WriteBytesPerSec int64      `gorethik:"written_bytes_per_sec"`
	WriteBytesTotal  int64      `gorethink:"written_bytes_total"`
	SpaceUsage       SpaceUsage `gorethink:"space_usage"`
}

type SpaceUsage struct {
	Data     int64 `gorethink:"data_bytes"`
	Garbage  int64 `gorethink:"garbage_bytes"`
	Metadata int64 `gorethink:"metadata_bytes"`
	Prealloc int64 `gorethink:"preallocated_bytes"`
}

var engineStats = map[string]string{
	"active_clients":       "ClientActive",
	"clients":              "ClientConns",
	"queries_per_sec":      "QueriesPerSec",
	"total_queries":        "TotalQueries",
	"read_docs_per_sec":    "ReadsPerSec",
	"total_reads":          "TotalReads",
	"written_docs_per_sec": "WritesPerSec",
	"total_writes":         "TotalWrites",
}

func (e *Engine) AddEngineStats(
	keys []string,
	acc telegraf.Accumulator,
	tags map[string]string,
) {
	engine := reflect.ValueOf(e).Elem()
	fields := make(map[string]interface{})
	for _, key := range keys {
		fields[key] = engine.FieldByName(engineStats[key]).Interface()
	}
	acc.AddFields("rethinkdb_engine", fields, tags)
}

func (s *Storage) AddStats(acc telegraf.Accumulator, tags map[string]string) {
	fields := map[string]interface{}{
		"cache_bytes_in_use":            s.Cache.BytesInUse,
		"disk_read_bytes_per_sec":       s.Disk.ReadBytesPerSec,
		"disk_read_bytes_total":         s.Disk.ReadBytesTotal,
		"disk_written_bytes_per_sec":    s.Disk.WriteBytesPerSec,
		"disk_written_bytes_total":      s.Disk.WriteBytesTotal,
		"disk_usage_data_bytes":         s.Disk.SpaceUsage.Data,
		"disk_usage_garbage_bytes":      s.Disk.SpaceUsage.Garbage,
		"disk_usage_metadata_bytes":     s.Disk.SpaceUsage.Metadata,
		"disk_usage_preallocated_bytes": s.Disk.SpaceUsage.Prealloc,
	}
	acc.AddFields("rethinkdb", fields, tags)
}
