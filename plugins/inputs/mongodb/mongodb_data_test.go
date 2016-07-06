package mongodb

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var tags = make(map[string]string)

func TestAddNonReplStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine:    "",
			Time:             time.Now(),
			Insert:           0,
			Query:            0,
			Update:           0,
			Delete:           0,
			GetMore:          0,
			Command:          0,
			Flushes:          0,
			Virtual:          0,
			Resident:         0,
			QueuedReaders:    0,
			QueuedWriters:    0,
			ActiveReaders:    0,
			ActiveWriters:    0,
			NetIn:            0,
			NetOut:           0,
			NumConnections:   0,
			Passes:           0,
			DeletedDocuments: 0,
		},
		tags,
	)
	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key, _ := range DefaultStats {
		assert.True(t, acc.HasIntField("mongodb", key))
	}
}

func TestAddReplStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine: "mmapv1",
			Mapped:        0,
			NonMapped:     0,
			Faults:        0,
		},
		tags,
	)

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key, _ := range MmapStats {
		assert.True(t, acc.HasIntField("mongodb", key))
	}
}

func TestAddWiredTigerStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine:     "wiredTiger",
			CacheDirtyPercent: 0,
			CacheUsedPercent:  0,
		},
		tags,
	)

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key, _ := range WiredTigerStats {
		assert.True(t, acc.HasFloatField("mongodb", key))
	}
}

func TestStateTag(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine: "",
			Time:          time.Now(),
			Insert:        0,
			Query:         0,
			NodeType:      "PRI",
		},
		tags,
	)

	stateTags := make(map[string]string)
	stateTags["state"] = "PRI"

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)
	fields := map[string]interface{}{
		"active_reads":          int64(0),
		"active_writes":         int64(0),
		"commands_per_sec":      int64(0),
		"deletes_per_sec":       int64(0),
		"flushes_per_sec":       int64(0),
		"getmores_per_sec":      int64(0),
		"inserts_per_sec":       int64(0),
		"member_status":         "PRI",
		"net_in_bytes":          int64(0),
		"net_out_bytes":         int64(0),
		"open_connections":      int64(0),
		"queries_per_sec":       int64(0),
		"queued_reads":          int64(0),
		"queued_writes":         int64(0),
		"repl_commands_per_sec": int64(0),
		"repl_deletes_per_sec":  int64(0),
		"repl_getmores_per_sec": int64(0),
		"repl_inserts_per_sec":  int64(0),
		"repl_queries_per_sec":  int64(0),
		"repl_updates_per_sec":  int64(0),
		"repl_lag":              int64(0),
		"resident_megabytes":    int64(0),
		"updates_per_sec":       int64(0),
		"vsize_megabytes":       int64(0),
		"ttl_deletes_per_sec":   int64(0),
		"ttl_passes_per_sec":    int64(0),
		"jumbo_chunks":          int64(0),
	}
	acc.AssertContainsTaggedFields(t, "mongodb", fields, stateTags)
}
