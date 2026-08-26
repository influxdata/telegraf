package parquet

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	type testcase struct {
		name       string
		metrics    []telegraf.Metric
		numRows    int
		numColumns int
	}

	var testcases = []testcase{
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

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			plugin := &Parquet{
				Directory:          testDir,
				TimestampFieldName: defaultTimestampFieldName,
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			require.NoError(t, plugin.Write(tc.metrics))
			require.NoError(t, plugin.Close())

			// Read metrics from parquet file
			files, err := os.ReadDir(testDir)
			require.NoError(t, err)
			require.Len(t, files, 1)
			reader, err := file.OpenParquetFile(filepath.Join(testDir, files[0].Name()), false)
			require.NoError(t, err)
			defer reader.Close()

			metadata := reader.MetaData()
			require.Equal(t, tc.numRows, int(metadata.NumRows))
			require.Equal(t, tc.numColumns, metadata.Schema.NumColumns())
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

func TestUsableAsFilename(t *testing.T) {
	tests := []struct {
		name   string
		usable bool
	}{
		{"cpu", true},
		{"disk.io", true},
		{"..", true},
		{"", false},
		{"../../etc/passwd", false},
		{`a\b`, false},
		{"nul\x00byte", false},
		{"tab\tnewline\n", false},
		{strings.Repeat("a", maxMeasurementLen), true},
		{strings.Repeat("a", maxMeasurementLen+1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.usable, usableAsFilename(tt.name))
		})
	}
}

func TestWindowsReservedCharacters(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}

	require.False(t, usableAsFilename(`a<b>c`))
}

func TestMeasurementNamesCannotEscapeDirectory(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Dir(dir)
	before, err := filepath.Glob(filepath.Join(outside, "*.parquet"))
	require.NoError(t, err)

	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	for _, name := range []string{"../../../../tmp/evil", "../evil", `a\..\..\b`, "nul\x00byte"} {
		m := metric.New(name, nil, map[string]interface{}{"value": int64(1)}, time.Now())
		require.ErrorAs(t, p.Write([]telegraf.Metric{m}), new(*internal.PartialWriteError))
	}
	require.NoError(t, p.Close())

	after, err := filepath.Glob(filepath.Join(outside, "*.parquet"))
	require.NoError(t, err)
	require.Equal(t, before, after)

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Empty(t, written)
}

func TestUnusableMeasurementNameKeepsOutputWriting(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.ErrorAs(t, p.Write([]telegraf.Metric{
		metric.New(strings.Repeat("a", 300), nil, map[string]interface{}{"value": int64(1)}, time.Now()),
	}), new(*internal.PartialWriteError))

	for i := 0; i < 8; i++ {
		require.NoError(t, p.Write([]telegraf.Metric{
			metric.New("good", nil, map[string]interface{}{"value": int64(i)}, time.Now()),
		}))
	}
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)
	require.Contains(t, filepath.Base(written[0]), "good")
}

func TestRejectedMetricsReportEveryIndexExactlyOnce(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	metrics := []telegraf.Metric{
		metric.New("../evil", nil, map[string]interface{}{"value": int64(1)}, time.Now()),
		metric.New("good", nil, map[string]interface{}{"value": int64(2)}, time.Now()),
		metric.New("nul\x00byte", nil, map[string]interface{}{"value": int64(3)}, time.Now()),
		metric.New("alsogood", nil, map[string]interface{}{"value": int64(4)}, time.Now()),
	}

	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, p.Write(metrics), &writeErr)
	require.Equal(t, []int{0, 2}, writeErr.MetricsReject)
	require.Equal(t, []int{1, 3}, writeErr.MetricsAccept)
	require.Len(t, writeErr.MetricsRejectErrors, 2)
	require.ElementsMatch(t,
		[]int{0, 1, 2, 3},
		append(slices.Clone(writeErr.MetricsAccept), writeErr.MetricsReject...),
	)
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 2)
}
