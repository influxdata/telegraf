package mongodb

import (
	"testing"
	//"time"

	//"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, sl.CommandLatency, int64(0))
	assert.Equal(t, sl.ReadLatency, int64(0))
	assert.Equal(t, sl.WriteLatency, int64(0))
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

	assert.Equal(t, sl.CommandLatency, int64(0))
	assert.Equal(t, sl.ReadLatency, int64(0))
	assert.Equal(t, sl.WriteLatency, int64(0))
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

	assert.Equal(t, sl.CommandLatency, int64( (59177981552 - 59177710371) / (1019152861 - 1019150402) ))
	assert.Equal(t, sl.ReadLatency, int64( (2255946760057 - 2255922322753) / (4189049884 - 4189041956) ))
	assert.Equal(t, sl.WriteLatency, int64( (494479456987 - 494478256915) / (1691021287 - 1691019457) ))
}
