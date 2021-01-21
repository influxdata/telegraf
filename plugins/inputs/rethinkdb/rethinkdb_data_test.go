package rethinkdb

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var tags = make(map[string]string)

func TestAddEngineStats(t *testing.T) {
	engine := &Engine{
		ClientConns:   0,
		ClientActive:  0,
		QueriesPerSec: 0,
		TotalQueries:  0,
		ReadsPerSec:   0,
		TotalReads:    0,
		WritesPerSec:  0,
		TotalWrites:   0,
	}

	var acc testutil.Accumulator

	keys := []string{
		"active_clients",
		"clients",
		"queries_per_sec",
		"total_queries",
		"read_docs_per_sec",
		"total_reads",
		"written_docs_per_sec",
		"total_writes",
	}
	engine.AddEngineStats(keys, &acc, tags)

	for _, metric := range keys {
		assert.True(t, acc.HasInt64Field("rethinkdb_engine", metric))
	}
}

func TestAddEngineStatsPartial(t *testing.T) {
	engine := &Engine{
		ClientConns:   0,
		ClientActive:  0,
		QueriesPerSec: 0,
		ReadsPerSec:   0,
		WritesPerSec:  0,
	}

	var acc testutil.Accumulator

	keys := []string{
		"active_clients",
		"clients",
		"queries_per_sec",
		"read_docs_per_sec",
		"written_docs_per_sec",
	}

	missing_keys := []string{
		"total_queries",
		"total_reads",
		"total_writes",
	}
	engine.AddEngineStats(keys, &acc, tags)

	for _, metric := range missing_keys {
		assert.False(t, acc.HasInt64Field("rethinkdb", metric))
	}
}

func TestAddStorageStats(t *testing.T) {
	storage := &Storage{
		Cache: Cache{
			BytesInUse: 0,
		},
		Disk: Disk{
			ReadBytesPerSec:  0,
			ReadBytesTotal:   0,
			WriteBytesPerSec: 0,
			WriteBytesTotal:  0,
			SpaceUsage: SpaceUsage{
				Data:     0,
				Garbage:  0,
				Metadata: 0,
				Prealloc: 0,
			},
		},
	}

	var acc testutil.Accumulator

	keys := []string{
		"cache_bytes_in_use",
		"disk_read_bytes_per_sec",
		"disk_read_bytes_total",
		"disk_written_bytes_per_sec",
		"disk_written_bytes_total",
		"disk_usage_data_bytes",
		"disk_usage_garbage_bytes",
		"disk_usage_metadata_bytes",
		"disk_usage_preallocated_bytes",
	}

	storage.AddStats(&acc, tags)

	for _, metric := range keys {
		assert.True(t, acc.HasInt64Field("rethinkdb", metric))
	}
}
