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
