package mongodb

import (
	"sort"
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
			UptimeNanos:      0,
			Insert:           0,
			Query:            0,
			Update:           0,
			UpdateCnt:        0,
			Delete:           0,
			GetMore:          0,
			Command:          0,
			Flushes:          0,
			FlushesCnt:       0,
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
			TimedOutC:        0,
			NoTimeoutC:       0,
			PinnedC:          0,
			TotalC:           0,
			DeletedD:         0,
			InsertedD:        0,
			ReturnedD:        0,
			UpdatedD:         0,
			CurrentC:         0,
			AvailableC:       0,
			TotalCreatedC:    0,
		},
		tags,
	)
	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key := range DefaultStats {
		assert.True(t, acc.HasFloatField("mongodb", key) || acc.HasInt64Field("mongodb", key), key)
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

	for key := range MmapStats {
		assert.True(t, acc.HasInt64Field("mongodb", key), key)
	}
}

func TestAddWiredTigerStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine:             "wiredTiger",
			CacheDirtyPercent:         0,
			CacheUsedPercent:          0,
			TrackedDirtyBytes:         0,
			CurrentCachedBytes:        0,
			MaxBytesConfigured:        0,
			AppThreadsPageReadCount:   0,
			AppThreadsPageReadTime:    0,
			AppThreadsPageWriteCount:  0,
			BytesWrittenFrom:          0,
			BytesReadInto:             0,
			PagesEvictedByAppThread:   0,
			PagesQueuedForEviction:    0,
			ServerEvictingPages:       0,
			WorkerThreadEvictingPages: 0,
			FaultsCnt:                 204,
		},
		tags,
	)

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key := range WiredTigerStats {
		assert.True(t, acc.HasFloatField("mongodb", key), key)
	}

	for key := range WiredTigerExtStats {
		assert.True(t, acc.HasFloatField("mongodb", key) || acc.HasInt64Field("mongodb", key), key)
	}

	assert.True(t, acc.HasInt64Field("mongodb", "page_faults"))
}

func TestAddShardStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			TotalInUse:      0,
			TotalAvailable:  0,
			TotalCreated:    0,
			TotalRefreshing: 0,
		},
		tags,
	)

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key := range DefaultShardStats {
		assert.True(t, acc.HasInt64Field("mongodb", key))
	}
}

func TestAddLatencyStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			CommandOpsCnt:  73,
			CommandLatency: 364,
			ReadOpsCnt:     113,
			ReadLatency:    201,
			WriteOpsCnt:    7,
			WriteLatency:   55,
		},
		tags,
	)

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)

	for key := range DefaultLatencyStats {
		assert.True(t, acc.HasInt64Field("mongodb", key))
	}
}

func TestAddShardHostStats(t *testing.T) {
	expectedHosts := []string{"hostA", "hostB"}
	hostStatLines := map[string]ShardHostStatLine{}
	for _, host := range expectedHosts {
		hostStatLines[host] = ShardHostStatLine{
			InUse:      0,
			Available:  0,
			Created:    0,
			Refreshing: 0,
		}
	}

	d := NewMongodbData(
		&StatLine{
			ShardHostStatsLines: hostStatLines,
		},
		map[string]string{}, // Use empty tags, so we don't break existing tests
	)

	var acc testutil.Accumulator
	d.AddShardHostStats()
	d.flush(&acc)

	var hostsFound []string
	for host := range hostStatLines {
		for key := range ShardHostStats {
			assert.True(t, acc.HasInt64Field("mongodb_shard_stats", key))
		}

		assert.True(t, acc.HasTag("mongodb_shard_stats", "hostname"))
		hostsFound = append(hostsFound, host)
	}
	sort.Strings(hostsFound)
	sort.Strings(expectedHosts)
	assert.Equal(t, hostsFound, expectedHosts)
}

func TestStateTag(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine: "",
			Time:          time.Now(),
			Insert:        0,
			Query:         0,
			NodeType:      "PRI",
			NodeState:     "PRIMARY",
			ReplSetName:   "rs1",
		},
		tags,
	)

	stateTags := make(map[string]string)
	stateTags["node_type"] = "PRI"
	stateTags["rs_name"] = "rs1"

	var acc testutil.Accumulator

	d.AddDefaultStats()
	d.flush(&acc)
	fields := map[string]interface{}{
		"active_reads":              int64(0),
		"active_writes":             int64(0),
		"commands":                  int64(0),
		"commands_per_sec":          int64(0),
		"deletes":                   int64(0),
		"deletes_per_sec":           int64(0),
		"flushes":                   int64(0),
		"flushes_per_sec":           int64(0),
		"flushes_total_time_ns":     int64(0),
		"getmores":                  int64(0),
		"getmores_per_sec":          int64(0),
		"inserts":                   int64(0),
		"inserts_per_sec":           int64(0),
		"member_status":             "PRI",
		"state":                     "PRIMARY",
		"net_in_bytes_count":        int64(0),
		"net_in_bytes":              int64(0),
		"net_out_bytes_count":       int64(0),
		"net_out_bytes":             int64(0),
		"open_connections":          int64(0),
		"queries":                   int64(0),
		"queries_per_sec":           int64(0),
		"queued_reads":              int64(0),
		"queued_writes":             int64(0),
		"repl_commands":             int64(0),
		"repl_commands_per_sec":     int64(0),
		"repl_deletes":              int64(0),
		"repl_deletes_per_sec":      int64(0),
		"repl_getmores":             int64(0),
		"repl_getmores_per_sec":     int64(0),
		"repl_inserts":              int64(0),
		"repl_inserts_per_sec":      int64(0),
		"repl_queries":              int64(0),
		"repl_queries_per_sec":      int64(0),
		"repl_updates":              int64(0),
		"repl_updates_per_sec":      int64(0),
		"repl_lag":                  int64(0),
		"resident_megabytes":        int64(0),
		"updates":                   int64(0),
		"updates_per_sec":           int64(0),
		"uptime_ns":                 int64(0),
		"vsize_megabytes":           int64(0),
		"ttl_deletes":               int64(0),
		"ttl_deletes_per_sec":       int64(0),
		"ttl_passes":                int64(0),
		"ttl_passes_per_sec":        int64(0),
		"jumbo_chunks":              int64(0),
		"total_in_use":              int64(0),
		"total_available":           int64(0),
		"total_created":             int64(0),
		"total_refreshing":          int64(0),
		"cursor_timed_out":          int64(0),
		"cursor_timed_out_count":    int64(0),
		"cursor_no_timeout":         int64(0),
		"cursor_no_timeout_count":   int64(0),
		"cursor_pinned":             int64(0),
		"cursor_pinned_count":       int64(0),
		"cursor_total":              int64(0),
		"cursor_total_count":        int64(0),
		"document_deleted":          int64(0),
		"document_inserted":         int64(0),
		"document_returned":         int64(0),
		"document_updated":          int64(0),
		"connections_current":       int64(0),
		"connections_available":     int64(0),
		"connections_total_created": int64(0),
	}
	acc.AssertContainsTaggedFields(t, "mongodb", fields, stateTags)
}
