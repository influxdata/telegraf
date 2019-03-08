package diskio

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/require"
)

func TestDiskIO(t *testing.T) {
	type Result struct {
		stats map[string]disk.IOCountersStat
		err   error
	}
	type Metric struct {
		tags   map[string]string
		fields map[string]interface{}
	}

	tests := []struct {
		name    string
		devices []string
		result  Result
		err     error
		metrics []Metric
	}{
		{
			name: "minimal",
			result: Result{
				stats: map[string]disk.IOCountersStat{
					"sda": {
						ReadCount:    888,
						WriteCount:   5341,
						ReadBytes:    100000,
						WriteBytes:   200000,
						ReadTime:     7123,
						WriteTime:    9087,
						Name:         "sda",
						IoTime:       123552,
						SerialNumber: "ab-123-ad",
					},
				},
				err: nil,
			},
			err: nil,
			metrics: []Metric{
				{
					tags: map[string]string{
						"name":   "sda",
						"serial": "ab-123-ad",
					},
					fields: map[string]interface{}{
						"reads":            uint64(888),
						"writes":           uint64(5341),
						"read_bytes":       uint64(100000),
						"write_bytes":      uint64(200000),
						"read_time":        uint64(7123),
						"write_time":       uint64(9087),
						"io_time":          uint64(123552),
						"weighted_io_time": uint64(0),
						"iops_in_progress": uint64(0),
					},
				},
			},
		},
		{
			name:    "glob device",
			devices: []string{"sd*"},
			result: Result{
				stats: map[string]disk.IOCountersStat{
					"sda": {
						Name:      "sda",
						ReadCount: 42,
					},
					"vda": {
						Name:      "vda",
						ReadCount: 42,
					},
				},
				err: nil,
			},
			err: nil,
			metrics: []Metric{
				{
					tags: map[string]string{
						"name":   "sda",
						"serial": "unknown",
					},
					fields: map[string]interface{}{
						"reads": uint64(42),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mps system.MockPS
			mps.On("DiskIO").Return(tt.result.stats, tt.result.err)

			var acc testutil.Accumulator

			diskio := &DiskIO{
				ps:      &mps,
				Devices: tt.devices,
			}
			err := diskio.Gather(&acc)
			require.Equal(t, tt.err, err)

			for _, metric := range tt.metrics {
				for k, v := range metric.fields {
					require.True(t, acc.HasPoint("diskio", metric.tags, k, v),
						"missing point: diskio %v %q: %v", metric.tags, k, v)
				}
			}
			require.Equal(t, len(tt.metrics), int(acc.NMetrics()), "unexpected number of metrics")
			require.True(t, mps.AssertExpectations(t))
		})
	}
}
