package mongodb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatencyStats(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Bits:              0,
					Resident:          0,
					Virtual:           0,
					Supported:         false,
					Mapped:            0,
					MappedWithJournal: 0,
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Bits:              0,
					Resident:          0,
					Virtual:           0,
					Supported:         false,
					Mapped:            0,
					MappedWithJournal: 0,
				},
				OpLatencies: &OpLatenciesStats{
					Reads: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
					Writes: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
					Commands: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Equal(t, sl.CommandLatency, int64(0))
	require.Equal(t, sl.ReadLatency, int64(0))
	require.Equal(t, sl.WriteLatency, int64(0))
	require.Equal(t, sl.CommandOpsCnt, int64(0))
	require.Equal(t, sl.ReadOpsCnt, int64(0))
	require.Equal(t, sl.WriteOpsCnt, int64(0))
}

func TestLatencyStatsDiffZero(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Bits:              0,
					Resident:          0,
					Virtual:           0,
					Supported:         false,
					Mapped:            0,
					MappedWithJournal: 0,
				},
				OpLatencies: &OpLatenciesStats{
					Reads: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
					Writes: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
					Commands: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Bits:              0,
					Resident:          0,
					Virtual:           0,
					Supported:         false,
					Mapped:            0,
					MappedWithJournal: 0,
				},
				OpLatencies: &OpLatenciesStats{
					Reads: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
					Writes: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
					Commands: &LatencyStats{
						Ops:     0,
						Latency: 0,
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Equal(t, sl.CommandLatency, int64(0))
	require.Equal(t, sl.ReadLatency, int64(0))
	require.Equal(t, sl.WriteLatency, int64(0))
	require.Equal(t, sl.CommandOpsCnt, int64(0))
	require.Equal(t, sl.ReadOpsCnt, int64(0))
	require.Equal(t, sl.WriteOpsCnt, int64(0))
}

func TestLatencyStatsDiff(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Bits:              0,
					Resident:          0,
					Virtual:           0,
					Supported:         false,
					Mapped:            0,
					MappedWithJournal: 0,
				},
				OpLatencies: &OpLatenciesStats{
					Reads: &LatencyStats{
						Ops:     4189041956,
						Latency: 2255922322753,
					},
					Writes: &LatencyStats{
						Ops:     1691019457,
						Latency: 494478256915,
					},
					Commands: &LatencyStats{
						Ops:     1019150402,
						Latency: 59177710371,
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Bits:              0,
					Resident:          0,
					Virtual:           0,
					Supported:         false,
					Mapped:            0,
					MappedWithJournal: 0,
				},
				OpLatencies: &OpLatenciesStats{
					Reads: &LatencyStats{
						Ops:     4189049884,
						Latency: 2255946760057,
					},
					Writes: &LatencyStats{
						Ops:     1691021287,
						Latency: 494479456987,
					},
					Commands: &LatencyStats{
						Ops:     1019152861,
						Latency: 59177981552,
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Equal(t, sl.CommandLatency, int64(59177981552))
	require.Equal(t, sl.ReadLatency, int64(2255946760057))
	require.Equal(t, sl.WriteLatency, int64(494479456987))
	require.Equal(t, sl.CommandOpsCnt, int64(1019152861))
	require.Equal(t, sl.ReadOpsCnt, int64(4189049884))
	require.Equal(t, sl.WriteOpsCnt, int64(1691021287))
}

func TestLocksStatsNilWhenLocksMissingInOldStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenGlobalLockStatsMissingInOldStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenGlobalLockStatsEmptyInOldStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenCollectionLockStatsMissingInOldStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenCollectionLockStatsEmptyInOldStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
					"Collection": {},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenLocksMissingInNewStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenGlobalLockStatsMissingInNewStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenGlobalLockStatsEmptyInNewStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenCollectionLockStatsMissingInNewStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsNilWhenCollectionLockStatsEmptyInNewStat(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
					"Collection": {},
				},
			},
		},
		"foo",
		true,
		60,
	)

	require.Nil(t, sl.CollectionLocks)
}

func TestLocksStatsPopulated(t *testing.T) {
	sl := NewStatLine(
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
					"Collection": {
						AcquireWaitCount: &ReadWriteLockTimes{
							Read:  1,
							Write: 2,
						},
						AcquireCount: &ReadWriteLockTimes{
							Read:  5,
							Write: 10,
						},
						TimeAcquiringMicros: ReadWriteLockTimes{
							Read:  100,
							Write: 200,
						},
					},
				},
			},
		},
		MongoStatus{
			ServerStatus: &ServerStatus{
				Connections: &ConnectionStats{},
				Mem: &MemStats{
					Supported: false,
				},
				Locks: map[string]LockStats{
					"Global": {
						AcquireCount: &ReadWriteLockTimes{},
					},
					"Collection": {
						AcquireWaitCount: &ReadWriteLockTimes{
							Read:  2,
							Write: 4,
						},
						AcquireCount: &ReadWriteLockTimes{
							Read:  10,
							Write: 30,
						},
						TimeAcquiringMicros: ReadWriteLockTimes{
							Read:  250,
							Write: 310,
						},
					},
				},
			},
		},
		"foo",
		true,
		60,
	)

	expected := &CollectionLockStatus{
		ReadAcquireWaitsPercentage:  20,
		WriteAcquireWaitsPercentage: 10,
		ReadAcquireTimeMicros:       150,
		WriteAcquireTimeMicros:      55,
	}

	require.Equal(t, expected, sl.CollectionLocks)
}
