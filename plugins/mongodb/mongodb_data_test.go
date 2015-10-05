package mongodb

import (
	"testing"
	"time"

	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tags = make(map[string]string)

func TestAddNonReplStats(t *testing.T) {
	d := NewMongodbData(
		&StatLine{
			StorageEngine:  "",
			Time:           time.Now(),
			Insert:         0,
			Query:          0,
			Update:         0,
			Delete:         0,
			GetMore:        0,
			Command:        0,
			Flushes:        0,
			Virtual:        0,
			Resident:       0,
			QueuedReaders:  0,
			QueuedWriters:  0,
			ActiveReaders:  0,
			ActiveWriters:  0,
			NetIn:          0,
			NetOut:         0,
			NumConnections: 0,
		},
		tags,
	)
	var acc testutil.Accumulator

	d.AddDefaultStats(&acc)

	for key, _ := range DefaultStats {
		assert.True(t, acc.HasIntValue(key))
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

	d.AddDefaultStats(&acc)

	for key, _ := range MmapStats {
		assert.True(t, acc.HasIntValue(key))
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

	d.AddDefaultStats(&acc)

	for key, _ := range WiredTigerStats {
		assert.True(t, acc.HasFloatValue(key))
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

	stats := []string{"inserts_per_sec", "queries_per_sec"}

	stateTags := make(map[string]string)
	stateTags["state"] = "PRI"

	var acc testutil.Accumulator

	d.AddDefaultStats(&acc)

	for _, key := range stats {
		err := acc.ValidateTaggedValue(key, int64(0), stateTags)
		require.NoError(t, err)
	}
}
