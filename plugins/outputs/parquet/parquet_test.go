package parquet

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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

func TestMetricToFile(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"cpu", "cpu"},
		{"../../etc/passwd", ".._.._etc_passwd"},
		{`a/b\c`, "a_b_c"},
		{"nul\x00byte", "nul_byte"},
		{"tab\tnewline\n", "tab_newline_"},
		{strings.Repeat("a", 300), strings.Repeat("a", maxMeasurementLen)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parquet{Log: testutil.Logger{}}
			require.Equal(t, tt.expected, p.metricToFile(tt.name))
		})
	}
}

func TestWindowsReservedCharacters(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}

	p := &Parquet{Log: testutil.Logger{}}
	require.Equal(t, "a_b_c", p.metricToFile(`a<b>c`))
}

func TestMeasurementNamesCannotEscapeDirectory(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Dir(dir)
	before, err := filepath.Glob(filepath.Join(outside, "*.parquet"))
	require.NoError(t, err)

	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	for _, name := range []string{"../../../../tmp/evil", "..", "../evil", `a\..\..\b`} {
		m := metric.New(name, nil, map[string]interface{}{"value": int64(1)}, time.Now())
		require.NoError(t, p.Write([]telegraf.Metric{m}))
	}
	require.NoError(t, p.Close())

	after, err := filepath.Glob(filepath.Join(outside, "*.parquet"))
	require.NoError(t, err)
	require.Equal(t, before, after)

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 4)
}

func TestLongMeasurementNameKeepsOutputWriting(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New(strings.Repeat("a", 300), nil, map[string]interface{}{"value": int64(1)}, time.Now()),
	}))
	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("good", nil, map[string]interface{}{"value": int64(2)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 2)
	for _, name := range written {
		require.LessOrEqual(t, len(filepath.Base(name)), 255)
	}
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

func TestConflictingValueTypesDoNotPanic(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"k": int64(1)}, time.Now()),
	}))
	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", map[string]string{"k": "v"}, map[string]interface{}{"other": int64(2)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)

	var total int64
	for _, name := range written {
		reader, err := file.OpenParquetFile(name, false)
		require.NoError(t, err)
		total += reader.NumRows()
		require.NoError(t, reader.Close())
	}
	require.Equal(t, int64(2), total)
}

func TestEveryConvertibleTypeRoundTrips(t *testing.T) {
	values := map[string]interface{}{
		"int8": int8(1), "int16": int16(2), "int32": int32(3), "int64": int64(4), "int": 5,
		"uint8": uint8(6), "uint16": uint16(7), "uint32": uint32(8), "uint64": uint64(9), "uint": uint(10),
		"float32": float32(11), "float64": float64(12),
		"string": "thirteen", "bool": true,
	}

	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			datatype, err := goToArrowType(value)
			require.NoError(t, err)

			builder := array.NewBuilder(memory.DefaultAllocator, datatype)
			defer builder.Release()

			require.True(t, appendValue(builder, value))
			require.Equal(t, 0, builder.NullN())
		})
	}
}

func TestAppendValueNullsWhatItCannotWrite(t *testing.T) {
	builder := array.NewBuilder(memory.DefaultAllocator, arrow.PrimitiveTypes.Int64)
	defer builder.Release()

	require.False(t, appendValue(builder, "not an int"))
	require.False(t, appendValue(builder, []string{"unsupported"}))
	require.True(t, appendValue(builder, nil))
	require.Equal(t, 3, builder.NullN())
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

func TestColumnOrderIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New(
			"demo",
			map[string]string{"zone": "a"},
			map[string]interface{}{"delta": int64(1), "alpha": int64(2), "charlie": int64(3)},
			time.Now(),
		),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)

	require.Equal(t, []string{"alpha", "charlie", "delta", "zone", "timestamp"}, columnNames(t, written[0]))
}

func TestUnsupportedFieldTypeDoesNotStallOutput(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	poisoned := metric.New("demo", nil, map[string]interface{}{"value": int64(1)}, time.Now())
	poisoned.AddField("unsupported", struct{ x int }{1})
	require.NoError(t, p.Write([]telegraf.Metric{poisoned}))

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"value": int64(2)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)

	reader, err := file.OpenParquetFile(written[0], false)
	require.NoError(t, err)
	defer reader.Close()
	require.Equal(t, int64(2), reader.NumRows())

	schema := reader.MetaData().Schema
	names := make([]string, 0, schema.NumColumns())
	for i := 0; i < schema.NumColumns(); i++ {
		names = append(names, schema.Column(i).Name())
	}
	require.Equal(t, []string{"value", "timestamp"}, names)
}

func TestExistingFilesAreNeverOverwritten(t *testing.T) {
	dir := t.TempDir()

	for i := 1; i <= 3; i++ {
		p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
		require.NoError(t, p.Init())
		require.NoError(t, p.Write([]telegraf.Metric{
			metric.New("demo", nil, map[string]interface{}{"value": int64(i)}, time.Now()),
		}))
		require.NoError(t, p.Close())
	}

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 3)

	var total int64
	for _, name := range written {
		reader, err := file.OpenParquetFile(name, false)
		require.NoError(t, err)
		total += reader.NumRows()
		require.NoError(t, reader.Close())
	}
	require.Equal(t, int64(3), total)
}

func TestRotationSurvivesAMissingFile(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{
		Directory:          dir,
		RotationInterval:   config.Duration(time.Nanosecond),
		TimestampFieldName: "timestamp",
		Log:                testutil.Logger{},
	}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"value": int64(1)}, time.Now()),
	}))

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)
	require.NoError(t, os.Remove(written[0]))

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"value": int64(2)}, time.Now()),
	}))
	require.NoError(t, p.Close())
}

func TestFieldTakesPrecedenceOverTagOfTheSameName(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", map[string]string{"k": "from tag"}, map[string]interface{}{"other": int64(0)}, time.Now()),
		metric.New("demo", nil, map[string]interface{}{"k": int64(1)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)

	reader, err := file.OpenParquetFile(written[0], false)
	require.NoError(t, err)
	defer reader.Close()

	schema := reader.MetaData().Schema
	require.Equal(t, parquet.Types.Int64, schema.Column(schema.ColumnIndexByName("k")).PhysicalType())
}

func columnNames(t *testing.T, path string) []string {
	t.Helper()

	reader, err := file.OpenParquetFile(path, false)
	require.NoError(t, err)
	defer reader.Close()

	schema := reader.MetaData().Schema
	names := make([]string, 0, schema.NumColumns())
	for i := 0; i < schema.NumColumns(); i++ {
		names = append(names, schema.Column(i).Name())
	}

	return names
}

func TestNewColumnStartsANewFile(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"a": int64(1)}, time.Now()),
	}))
	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"a": int64(2), "b": int64(3)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 2)

	schemas := make([][]string, 0, len(written))
	for _, name := range written {
		schemas = append(schemas, columnNames(t, name))
	}
	require.ElementsMatch(t, [][]string{{"a", "timestamp"}, {"a", "b", "timestamp"}}, schemas)
}

func TestOmittedFieldKeepsTheSameFile(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"a": int64(1), "b": int64(2)}, time.Now()),
	}))
	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"a": int64(3)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)
	require.Equal(t, []string{"a", "b", "timestamp"}, columnNames(t, written[0]))
}

func TestMaxColumnsCapsTheSchema(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", MaxColumns: 3, Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	fields := make(map[string]interface{}, 10)
	for i := 0; i < 10; i++ {
		fields[fmt.Sprintf("f%02d", i)] = int64(i)
	}
	require.NoError(t, p.Write([]telegraf.Metric{metric.New("demo", nil, fields, time.Now())}))

	require.NoError(t, p.Write([]telegraf.Metric{
		metric.New("demo", nil, map[string]interface{}{"late": int64(1)}, time.Now()),
	}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, written, 1)
	require.Equal(t, []string{"f00", "f01", "f02", "timestamp"}, columnNames(t, written[0]))
}

func TestMaxColumnsZeroMeansUnlimited(t *testing.T) {
	dir := t.TempDir()
	p := &Parquet{Directory: dir, TimestampFieldName: "timestamp", MaxColumns: 0, Log: testutil.Logger{}}
	require.NoError(t, p.Init())

	fields := make(map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		fields[fmt.Sprintf("f%02d", i)] = int64(i)
	}
	require.NoError(t, p.Write([]telegraf.Metric{metric.New("demo", nil, fields, time.Now())}))
	require.NoError(t, p.Close())

	written, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	require.NoError(t, err)
	require.Len(t, columnNames(t, written[0]), 51)
}
