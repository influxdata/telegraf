package parquet

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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
				testutil.MustMetric(
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
				testutil.MustMetric(
					"test",
					map[string]string{
						"tag": "tag",
					},
					map[string]interface{}{
						"value": 1.0,
					},
					time.Now(),
				),
				testutil.MustMetric(
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
				testutil.MustMetric(
					"test",
					map[string]string{
						"host": "tag",
					},
					map[string]interface{}{
						"value_old": 1.0,
					},
					time.Now(),
				),
				testutil.MustMetric(
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
				testutil.MustMetric(
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
		testutil.MustMetric(
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
		testutil.MustMetric(
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
		testutil.MustMetric(
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
