package parquet

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	tests := []struct {
		name       string
		metrics    []telegraf.Metric
		numRows    int
		numColumns int
	}{
		{
			name: "basic single metric",
			metrics: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{},
					map[string]interface{}{
						"value": 1.0,
					},
					time.Now(),
				),
			},
			numRows:    1,
			numColumns: 2,
		},
		{
			name: "mix of tags and fields",
			metrics: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{
						"tag": "tag",
					},
					map[string]interface{}{
						"value": 1.0,
					},
					time.Now(),
				),
				metric.New(
					"test",
					map[string]string{
						"tag": "tag2",
					},
					map[string]interface{}{
						"value": 2.0,
					},
					time.Now(),
				),
			},
			numRows:    2,
			numColumns: 3,
		},
		{
			name: "null values",
			metrics: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{
						"host": "tag",
					},
					map[string]interface{}{
						"value_old": 1.0,
					},
					time.Now(),
				),
				metric.New(
					"test",
					map[string]string{
						"tag": "tag2",
					},
					map[string]interface{}{
						"value_new": 2.0,
					},
					time.Now(),
				),
			},
			numRows:    2,
			numColumns: 5,
		},
		{
			name: "data types",
			metrics: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{},
					map[string]interface{}{
						"int":     int(0),
						"int8":    int8(1),
						"int16":   int16(2),
						"int32":   int32(3),
						"int64":   int64(4),
						"uint":    uint(5),
						"uint8":   uint8(6),
						"uint16":  uint16(7),
						"uint32":  uint32(8),
						"uint64":  uint64(9),
						"float32": float32(10.0),
						"float64": float64(11.0),
						"string":  "string",
						"bool":    true,
					},
					time.Now(),
				),
			},
			numRows:    1,
			numColumns: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()
			plugin := &Parquet{
				Directory:          testDir,
				TimestampFieldName: defaultTimestampFieldName,
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			require.NoError(t, plugin.Write(tt.metrics))
			require.NoError(t, plugin.Close())

			// Read metrics from parquet file
			files, err := os.ReadDir(testDir)
			require.NoError(t, err)
			require.Len(t, files, 1)
			reader, err := file.OpenParquetFile(filepath.Join(testDir, files[0].Name()), false)
			require.NoError(t, err)
			defer reader.Close()

			metadata := reader.MetaData()
			require.Equal(t, tt.numRows, int(metadata.NumRows))
			require.Equal(t, tt.numColumns, metadata.Schema.NumColumns())
		})
	}
}

func TestPartialWrite(t *testing.T) {
	tests := []struct {
		name       string
		metrics    []telegraf.Metric
		numRows    int
		numColumns int
		expected   string
		accepted   []int
		rejected   []int
	}{
		{
			name: "create writer failed",
			metrics: []telegraf.Metric{
				metric.New(
					"test/sub",
					map[string]string{},
					map[string]interface{}{
						"value": 1.0,
					},
					time.Now(),
				),
			},
			expected: "failed to create writer for file",
		},
		{
			name: "schema mismatch",
			metrics: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{},
					map[string]interface{}{
						"value": int8(1),
					},
					time.Now(),
				),
				metric.New(
					"test",
					map[string]string{},
					map[string]interface{}{
						"value": "2",
					},
					time.Now(),
				),
				metric.New(
					"test",
					map[string]string{},
					map[string]interface{}{
						"value": int8(3),
					},
					time.Now(),
				),
			},
			numRows:    2,
			numColumns: 2,
			expected:   "invalid value 2 (string) for column \"value\" (int64)",
			accepted:   []int{0, 2},
			rejected:   []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := t.TempDir()

			plugin := &Parquet{
				Directory:          path,
				TimestampFieldName: defaultTimestampFieldName,
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()
			var perr *internal.PartialWriteError
			require.ErrorAs(t, plugin.Write(tt.metrics), &perr)
			require.NoError(t, plugin.Close())
			require.ErrorContains(t, perr.Err, tt.expected)
			require.ElementsMatch(t, perr.MetricsAccept, tt.accepted)
			require.ElementsMatch(t, perr.MetricsReject, tt.rejected)

			// Read metrics from parquet file
			files, err := os.ReadDir(path)
			require.NoError(t, err)
			if tt.numRows == 0 {
				require.Empty(t, files)
				return
			}
			require.Len(t, files, 1)

			reader, err := file.OpenParquetFile(filepath.Join(path, files[0].Name()), false)
			require.NoError(t, err)
			defer reader.Close()

			metadata := reader.MetaData()
			require.Equal(t, tt.numRows, int(metadata.NumRows), "wrong number of rows")
			require.Equal(t, tt.numColumns, metadata.Schema.NumColumns(), "wrong number of columns")
		})
	}
}

func TestRotation(t *testing.T) {
	metrics := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{
				"value": 1.0,
			},
			time.Now(),
		),
	}

	testDir := t.TempDir()
	plugin := &Parquet{
		Directory:          testDir,
		RotationInterval:   config.Duration(1 * time.Second),
		TimestampFieldName: defaultTimestampFieldName,
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.Eventually(t, func() bool {
		require.NoError(t, plugin.Write(metrics))
		files, err := os.ReadDir(testDir)
		require.NoError(t, err)
		return len(files) == 2
	}, 5*time.Second, time.Second)
	require.NoError(t, plugin.Close())
}

func TestOmitTimestamp(t *testing.T) {
	metrics := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{
				"value": 1.0,
			},
			time.Now(),
		),
	}

	testDir := t.TempDir()
	plugin := &Parquet{
		Directory: testDir,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(metrics))
	require.NoError(t, plugin.Close())

	files, err := os.ReadDir(testDir)
	require.NoError(t, err)
	require.Len(t, files, 1)
	reader, err := file.OpenParquetFile(filepath.Join(testDir, files[0].Name()), false)
	require.NoError(t, err)
	defer reader.Close()

	metadata := reader.MetaData()
	require.Equal(t, 1, int(metadata.NumRows))
	require.Equal(t, 1, metadata.Schema.NumColumns())
}

func TestTimestampDifferentName(t *testing.T) {
	metrics := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{
				"value": 1.0,
			},
			time.Now(),
		),
	}

	testDir := t.TempDir()
	plugin := &Parquet{
		Directory:          testDir,
		TimestampFieldName: "time",
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(metrics))
	require.NoError(t, plugin.Close())

	files, err := os.ReadDir(testDir)
	require.NoError(t, err)
	require.Len(t, files, 1)
	reader, err := file.OpenParquetFile(filepath.Join(testDir, files[0].Name()), false)
	require.NoError(t, err)
	defer reader.Close()

	metadata := reader.MetaData()
	require.Equal(t, 1, int(metadata.NumRows))
	require.Equal(t, 2, metadata.Schema.NumColumns())
}

func TestMissingValuesReadBackAsNull(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	now := time.Now()
	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"a": int64(1), "b": int64(2)}, now),
		metric.New("demo", nil, map[string]interface{}{"a": int64(3)}, now),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)

	reader, err := file.OpenParquetFile(written[0], false)
	require.NoError(t, err)
	defer reader.Close()

	schema := reader.MetaData().Schema
	for i := 0; i < schema.NumColumns(); i++ {
		column := schema.Column(i)
		if column.Name() == "timestamp" {
			continue
		}
		require.Equalf(t, int16(1), column.MaxDefinitionLevel(), "column %q is required, nulls would be written as zero", column.Name())
	}
}

func TestTimestampFieldNameCollisionKeepsOneColumn(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"timestamp": "x", "value": int64(1)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)

	reader, err := file.OpenParquetFile(written[0], false)
	require.NoError(t, err)
	defer reader.Close()

	schema := reader.MetaData().Schema
	names := make([]string, 0, schema.NumColumns())
	for i := 0; i < schema.NumColumns(); i++ {
		names = append(names, schema.Column(i).Name())
	}
	require.ElementsMatch(t, []string{"value", "timestamp"}, names)
	require.Equal(t, parquet.Types.Int64, schema.Column(schema.ColumnIndexByName("timestamp")).PhysicalType())
}

func TestInvalidFilename(t *testing.T) {
	tests := []struct {
		name   string
		metric string
		os     []string
	}{
		{
			name:   "too long",
			metric: strings.Repeat("a", 255),
		},
		{
			name:   "null byte",
			metric: "nul\x00byte",
		},
		{
			name:   "dots",
			metric: "..",
			os:     []string{"windows"},
		},
		{
			name:   "backslash",
			metric: `a\b`,
			os:     []string{"windows"},
		},
		{
			name:   "tab",
			metric: "a\tb",
			os:     []string{"windows"},
		},
		{
			name:   "a<b>c",
			metric: "brackets",
			os:     []string{"windows"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.os) != 0 && !slices.Contains(tt.os, runtime.GOOS) {
				t.Skip("Skipping due to unaffected OS...")
			}

			metrics := []telegraf.Metric{
				metric.New(
					tt.metric,
					map[string]string{},
					map[string]interface{}{"value": 1.0},
					time.Now(),
				),
			}

			testDir := t.TempDir()
			plugin := &Parquet{
				Directory:          testDir,
				TimestampFieldName: "time",
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()
			require.ErrorContains(t, plugin.Write(metrics), "failed to create file")
		})
	}
}

func TestCannotEscapeDirectory(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test")
	malicious := filepath.Join(tmp, "foo")
	maliciousAbs, err := filepath.Abs(malicious)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(path, 0700))
	require.NoError(t, os.MkdirAll(malicious, 0700))

	tests := []struct {
		name   string
		metric string
	}{
		{
			name:   "relative",
			metric: filepath.Join("..", "foo"),
		},
		{
			name:   "absolute",
			metric: maliciousAbs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := []telegraf.Metric{
				metric.New(
					tt.metric,
					map[string]string{},
					map[string]interface{}{"value": 1.0},
					time.Now(),
				),
			}

			testDir := t.TempDir()
			plugin := &Parquet{
				Directory:          testDir,
				TimestampFieldName: "time",
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			var perr *os.PathError
			require.ErrorAs(t, plugin.Write(metrics), &perr)
		})
	}
}
