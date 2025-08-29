package diskio

import (
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/testutil"
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
						ReadCount:        888,
						WriteCount:       5341,
						ReadBytes:        100000,
						WriteBytes:       200000,
						ReadTime:         7123,
						WriteTime:        9087,
						MergedReadCount:  11,
						MergedWriteCount: 12,
						Name:             "sda",
						IoTime:           123552,
						SerialNumber:     "ab-123-ad",
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
						"merged_reads":     uint64(11),
						"merged_writes":    uint64(12),
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
			var mps psutil.MockPS
			mps.On("DiskIO").Return(tt.result.stats, tt.result.err)

			var acc testutil.Accumulator

			diskio := &DiskIO{
				Log:     testutil.Logger{},
				ps:      &mps,
				Devices: tt.devices,
			}
			require.NoError(t, diskio.Init())
			err := diskio.Gather(&acc)
			require.Equal(t, tt.err, err)

			for _, metric := range tt.metrics {
				for k, v := range metric.fields {
					require.True(t, acc.HasPoint("diskio", metric.tags, k, v),
						"missing point: diskio %v %q: %v", metric.tags, k, v)
				}
			}
			require.Len(t, tt.metrics, int(acc.NMetrics()), "unexpected number of metrics")
			require.True(t, mps.AssertExpectations(t))
		})
	}
}

func TestDiskIOUtil(t *testing.T) {
	cts := map[string]disk.IOCountersStat{
		"sda": {
			ReadCount:        888,
			WriteCount:       5341,
			ReadBytes:        100000,
			WriteBytes:       200000,
			ReadTime:         7123,
			WriteTime:        9087,
			MergedReadCount:  11,
			MergedWriteCount: 12,
			Name:             "sda",
			IoTime:           123552,
			SerialNumber:     "ab-123-ad",
		},
	}

	cts2 := map[string]disk.IOCountersStat{
		"sda": {
			ReadCount:        1000,
			WriteCount:       6000,
			ReadBytes:        200000,
			WriteBytes:       300000,
			ReadTime:         8123,
			WriteTime:        9187,
			MergedReadCount:  16,
			MergedWriteCount: 30,
			Name:             "sda",
			IoTime:           163552,
			SerialNumber:     "ab-123-ad",
		},
	}

	var acc testutil.Accumulator
	var mps psutil.MockPS
	mps.On("DiskIO").Return(cts, nil)
	diskio := &DiskIO{
		Log:     testutil.Logger{},
		Devices: []string{"sd*"},
		ps:      &mps,
	}
	require.NoError(t, diskio.Init())
	// gather
	require.NoError(t, diskio.Gather(&acc))
	// sleep
	time.Sleep(1 * time.Second)
	// gather twice
	mps2 := psutil.MockPS{}
	mps2.On("DiskIO").Return(cts2, nil)
	diskio.ps = &mps2

	err := diskio.Gather(&acc)
	require.NoError(t, err)
	require.True(t, acc.HasField("diskio", "io_util"), "miss io util")
	require.True(t, acc.HasField("diskio", "io_svctm"), "miss io_svctm")
	require.True(t, acc.HasField("diskio", "io_await"), "miss io_await")

	require.True(t, acc.HasFloatField("diskio", "io_util"), "io_util not have value")
	require.True(t, acc.HasFloatField("diskio", "io_svctm"), "io_svctm not have value")
	require.True(t, acc.HasFloatField("diskio", "io_await"), "io_await not have value")
}

func TestCounterWraparound(t *testing.T) {
	cts := map[string]disk.IOCountersStat{
		"sda": {
			ReadCount:  1000,
			WriteCount: 2000,
			ReadTime:   8000,
			WriteTime:  9000,
			IoTime:     30000,
			Name:       "sda",
		},
	}

	// Simulate wraparound - counters decrease significantly
	cts2 := map[string]disk.IOCountersStat{
		"sda": {
			ReadCount:  100,  // wrapped around
			WriteCount: 200,  // wrapped around
			ReadTime:   1000, // wrapped around
			WriteTime:  1500, // wrapped around
			IoTime:     3000, // wrapped around
			Name:       "sda",
		},
	}

	var acc testutil.Accumulator
	var mps psutil.MockPS
	mps.On("DiskIO").Return(cts, nil)
	diskio := &DiskIO{
		Log:     testutil.Logger{},
		Devices: []string{"sda"},
		ps:      &mps,
	}
	require.NoError(t, diskio.Init())

	// First gather to establish baseline
	require.NoError(t, diskio.Gather(&acc))

	// Second gather with wrapped counters
	mps2 := psutil.MockPS{}
	mps2.On("DiskIO").Return(cts2, nil)
	diskio.ps = &mps2

	require.NoError(t, diskio.Gather(&acc))

	// Should NOT have calculated fields due to wraparound detection
	require.False(t, acc.HasFloatField("diskio", "io_util"), "io_util should not be present on wraparound")
	require.False(t, acc.HasFloatField("diskio", "io_svctm"), "io_svctm should not be present on wraparound")
	require.False(t, acc.HasFloatField("diskio", "io_await"), "io_await should not be present on wraparound")

	// But basic counter fields should still be present
	require.True(t, acc.HasField("diskio", "reads"), "reads should be present")
	require.True(t, acc.HasField("diskio", "writes"), "writes should be present")
}
