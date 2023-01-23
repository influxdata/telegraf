//go:build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/grok"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestRefreshFilePaths(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/**.log")},
	}
	err = r.Init()
	require.NoError(t, err)

	err = r.refreshFilePaths()
	require.NoError(t, err)
	require.Equal(t, 2, len(r.filenames))
}

func TestFileTag(t *testing.T) {
	acc := testutil.Accumulator{}
	wd, err := os.Getwd()
	require.NoError(t, err)
	r := File{
		Files:   []string{filepath.Join(wd, "dev/testfiles/json_a.log")},
		FileTag: "filename",
	}
	require.NoError(t, r.Init())

	r.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{}
		err := p.Init()
		return p, err
	})

	require.NoError(t, r.Gather(&acc))

	for _, m := range acc.Metrics {
		for key, value := range m.Tags {
			require.Equal(t, r.FileTag, key)
			require.Equal(t, filepath.Base(r.Files[0]), value)
		}
	}
}

func TestJSONParserCompile(t *testing.T) {
	var acc testutil.Accumulator
	wd, _ := os.Getwd()
	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/json_a.log")},
	}
	require.NoError(t, r.Init())

	r.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{TagKeys: []string{"parent_ignored_child"}}
		err := p.Init()
		return p, err
	})

	require.NoError(t, r.Gather(&acc))
	require.Equal(t, map[string]string{"parent_ignored_child": "hi"}, acc.Metrics[0].Tags)
	require.Equal(t, 5, len(acc.Metrics[0].Fields))
}

func TestGrokParser(t *testing.T) {
	wd, _ := os.Getwd()
	var acc testutil.Accumulator
	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/grok_a.log")},
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (telegraf.Parser, error) {
		parser := &grok.Parser{
			Patterns: []string{"%{COMMON_LOG_FORMAT}"},
			Log:      testutil.Logger{},
		}
		err := parser.Init()

		return parser, err
	})

	err = r.Gather(&acc)
	require.NoError(t, err)
	require.Len(t, acc.Metrics, 2)
}

func TestCharacterEncoding(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "1",
				"ip":   "12.122.114.5",
			},
			map[string]interface{}{
				"avg":    21.55,
				"best":   19.34,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  2.05,
				"worst":  26.83,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "2",
				"ip":   "192.205.32.238",
			},
			map[string]interface{}{
				"avg":    25.11,
				"best":   20.8,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  6.03,
				"worst":  38.85,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "3",
				"ip":   "152.195.85.133",
			},
			map[string]interface{}{
				"avg":    20.18,
				"best":   19.75,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  0.0,
				"worst":  20.78,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "4",
				"ip":   "93.184.216.34",
			},
			map[string]interface{}{
				"avg":    24.02,
				"best":   19.75,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  4.67,
				"worst":  32.41,
			},
			time.Unix(0, 0),
		),
	}

	tests := []struct {
		name   string
		plugin *File
		csv    csv.Parser
		file   string
	}{
		{
			name: "empty character_encoding with utf-8",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-8.csv"},
				CharacterEncoding: "",
			},
			csv: csv.Parser{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
		{
			name: "utf-8 character_encoding with utf-8",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-8.csv"},
				CharacterEncoding: "utf-8",
			},
			csv: csv.Parser{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
		{
			name: "utf-16le character_encoding with utf-16le",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-16le.csv"},
				CharacterEncoding: "utf-16le",
			},
			csv: csv.Parser{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
		{
			name: "utf-16be character_encoding with utf-16be",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-16be.csv"},
				CharacterEncoding: "utf-16be",
			},
			csv: csv.Parser{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.NoError(t, err)

			tt.plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := tt.csv
				err := parser.Init()
				return &parser, err
			})

			var acc testutil.Accumulator
			err = tt.plugin.Gather(&acc)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestStatefulParsers(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "1",
				"ip":   "12.122.114.5",
			},
			map[string]interface{}{
				"avg":    21.55,
				"best":   19.34,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  2.05,
				"worst":  26.83,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "2",
				"ip":   "192.205.32.238",
			},
			map[string]interface{}{
				"avg":    25.11,
				"best":   20.8,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  6.03,
				"worst":  38.85,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "3",
				"ip":   "152.195.85.133",
			},
			map[string]interface{}{
				"avg":    20.18,
				"best":   19.75,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  0.0,
				"worst":  20.78,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "4",
				"ip":   "93.184.216.34",
			},
			map[string]interface{}{
				"avg":    24.02,
				"best":   19.75,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  4.67,
				"worst":  32.41,
			},
			time.Unix(0, 0),
		),
	}

	tests := []struct {
		name   string
		plugin *File
		csv    csv.Parser
		file   string
		count  int
	}{
		{
			name: "read file twice",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-8.csv"},
				CharacterEncoding: "",
			},
			csv: csv.Parser{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
			count: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.NoError(t, err)

			tt.plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := tt.csv
				err := parser.Init()
				return &parser, err
			})

			var acc testutil.Accumulator
			for i := 0; i < tt.count; i++ {
				require.NoError(t, tt.plugin.Gather(&acc))

				testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
				acc.ClearMetrics()
			}
		})
	}
}

func TestCSVBehavior(t *testing.T) {
	// Setup the CSV parser creator function
	parserFunc := func() (telegraf.Parser, error) {
		parser := &csv.Parser{
			MetricName:     "file",
			HeaderRowCount: 1,
		}
		err := parser.Init()
		return parser, err
	}

	// Setup the plugin
	plugin := &File{
		Files: []string{filepath.Join("testdata", "csv_behavior_input.csv")},
	}
	plugin.SetParserFunc(parserFunc)
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"file",
			map[string]string{},
			map[string]interface{}{
				"a": int64(1),
				"b": int64(2),
			},
			time.Unix(0, 1),
		),
		metric.New(
			"file",
			map[string]string{},
			map[string]interface{}{
				"a": int64(3),
				"b": int64(4),
			},
			time.Unix(0, 2),
		),
		metric.New(
			"file",
			map[string]string{},
			map[string]interface{}{
				"a": int64(1),
				"b": int64(2),
			},
			time.Unix(0, 3),
		),
		metric.New(
			"file",
			map[string]string{},
			map[string]interface{}{
				"a": int64(3),
				"b": int64(4),
			},
			time.Unix(0, 4),
		),
	}

	var acc testutil.Accumulator
	// Run gather once
	require.NoError(t, plugin.Gather(&acc))
	// Run gather a second time
	require.NoError(t, plugin.Gather(&acc))
	require.Eventuallyf(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= uint64(len(expected))
	}, time.Second, 100*time.Millisecond, "Expected %d metrics found %d", len(expected), acc.NMetrics())

	// Check the result
	options := []cmp.Option{
		testutil.SortMetrics(),
		testutil.IgnoreTime(),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, options...)
}
